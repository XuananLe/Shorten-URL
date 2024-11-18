package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"shorten-url/backend/pkg/db/sqlc"
	"shorten-url/backend/pkg/stores"
	"shorten-url/backend/pkg/utils"
	"sync"
	"time"
)

const (
	batchSize  = 100
	maxRetries = 3
	baseDelay  = 100 * time.Millisecond
)

type CachedURL struct {
	Original  string    `json:"original"`
	Clicks    int       `json:"clicks"`
	CreatedAt time.Time `json:"created_at"`
	ExpiredAt time.Time `json:"expired_at,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
}

type UrlService struct {
	ctx            context.Context
	cacheTimeout   time.Duration
	redisClient    *redis.ClusterClient
	postgresClient *stores.Postgres
	RabbitMQClient *stores.RabbitMQ
	cacheMutex     sync.RWMutex
	errorChan      chan error
	instanceId     string
}
type URLMessage struct {
	OriginalURL string `json:"original_url"`
	Shortened   string `json:"shortened"`
	UserID      string `json:"user_id"`
	Counter     int    `json:"counter"`
}

var UrlServiceInstance *UrlService

func NewUrlService(redisClient *redis.ClusterClient, postgresClient *stores.Postgres, RabbitMQClient *stores.RabbitMQ) *UrlService {
	UrlServiceInstance = &UrlService{
		ctx:            context.Background(),
		cacheTimeout:   24 * time.Hour,
		redisClient:    redisClient,
		postgresClient: postgresClient,
		RabbitMQClient: stores.RabbitMQClient,
		cacheMutex:     sync.RWMutex{},
		errorChan:      make(chan error, 100),
		instanceId:     uuid.New().String()[0:8],
	}

	go UrlServiceInstance.handleErrors()

	return UrlServiceInstance
}

func (s *UrlService) GetURL(shortenedURL string) (*CachedURL, error) {

	var cachedData *CachedURL
	var err error

	cachedData, err = s.getFromCache(shortenedURL)
	if err == nil {
		return cachedData, nil
	}

	url, err := s.getFromDB(shortenedURL)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func (s *UrlService) CreateURL(port string, originalURL string, userIDStr string) (string, error) {
	baseHash := utils.Hash(originalURL)
	shortenedURL := baseHash

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %v", err)
	}

	message := URLMessage{
		OriginalURL: originalURL,
		Shortened:   shortenedURL,
		UserID:      userID.String(),
		Counter:     0,
	}

	messageBody, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %v", err)
	}

	err = stores.RabbitMQClient.Channel.Publish(
		"",
		"queue-based-load-leveling-" + port,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to publish message: %v", err)
	}

	return shortenedURL, nil
}

func (s *UrlService) ProcessQueueBatch(port string, consumerTag string, batchSize int, batchTimeout time.Duration) {
	msgs, err := stores.RabbitMQClient.Channel.Consume(
		"queue-based-load-leveling-" + port, 
		consumerTag,                
		false,                       
		false,                     
		false,                       
		false,                      
		nil,  
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}


	batch := make([]URLMessage, 0, batchSize)
	timer := time.NewTimer(batchTimeout)

	for {
		select {
		case msg := <-msgs:
			var urlMessage URLMessage
			if err := json.Unmarshal(msg.Body, &urlMessage); err != nil {
				log.Printf("Consumer %s: Failed to unmarshal message: %v", consumerTag, err)
				msg.Nack(false, false)
				continue
			}

			batch = append(batch, urlMessage)

			msg.Ack(false)

			if len(batch) >= batchSize {
				s.processBatch(batch)
				batch = batch[:0]
				timer.Reset(batchTimeout)
			}

		case <-timer.C:
			if len(batch) > 0 {
				s.processBatch(batch)
				batch = batch[:0]
			}
			timer.Reset(batchTimeout)
		}
	}
}

func (s *UrlService) processBatch(batch []URLMessage) {
	if len(batch) == 0 {
		return
	}

	params := sqlc.BatchInsertURLsParams{
		Column1: make([]string, len(batch)),
		Column2: make([]string, len(batch)),
		Column3: make([]int64, len(batch)),
		Column4: make([]pgtype.Timestamptz, len(batch)),
		Column5: make([]pgtype.Timestamptz, len(batch)),
		Column6: make([]pgtype.UUID, len(batch)),
	}

	for i, message := range batch {
		params.Column1[i] = message.Shortened
		params.Column2[i] = message.OriginalURL
		params.Column3[i] = int64(message.Counter)
		params.Column4[i] = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		params.Column5[i] = pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour * 100), Valid: true}
		params.Column6[i] = pgtype.UUID{
			Bytes: uuid.MustParse(message.UserID),
			Valid: true,
		}
	}

	err := s.postgresClient.Queries.BatchInsertURLs(s.ctx, params)
	if err != nil {
		log.Printf("Failed to insert batch: %v", err)
	} else {
		fmt.Println("Successful append size ", len(batch));
	}
}

func (s *UrlService) IncrementClicks(shortenedURL string) (*CachedURL, error) {
	var updatedURL *CachedURL

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	cachedURL, err := s.getFromCache(shortenedURL)
	if err == nil {
		cachedURL.Clicks++
		updatedURL = cachedURL

		if err := s.setCache(shortenedURL, cachedURL); err != nil {
			log.Errorf("Failed to update clicks in cache for %s: %v", shortenedURL, err)
		}
	} else {
		dbURL, err := s.getFromDB(shortenedURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get URL from database: %v", err)
		}
		dbURL.Clicks++
		updatedURL = dbURL

		if err := s.setCache(shortenedURL, dbURL); err != nil {
			log.Errorf("Failed to set cache for %s: %v", shortenedURL, err)
		}
	}

	go func() {
		if err := s.postgresClient.Queries.IncrementClicks(s.ctx, shortenedURL); err != nil {
			log.Errorf("Failed to increment clicks in database for %s: %v", shortenedURL, err)
			s.errorChan <- fmt.Errorf("DB click increment failed for %s: %w", shortenedURL, err)
		}
	}()

	return updatedURL, nil
}

func (s *UrlService) DeleteURL(shortenedURL string) error {
	go func() {
		if err := s.postgresClient.Queries.DeleteURL(s.ctx, shortenedURL); err != nil {
			log.Errorf("failed to delete URL from database: %v", err)
		}
	}()

	if err := s.redisClient.Del(s.ctx, shortenedURL).Err(); err != nil {
		return fmt.Errorf("failed to delete URL from cache: %v", err)
	}

	return nil
}

func (s *UrlService) GetURLs(userId string) []CachedURL {
	userID, err := uuid.Parse(userId)
	if err != nil {
		log.Errorf("Failed to parse user ID: %v", err)
		return []CachedURL{}
	}

	urls, err := s.postgresClient.Queries.GetURLsByUser(s.ctx, utils.ConvertFromUuidPg(userID))
	if err != nil {
		log.Errorf("Failed to get URLs by user: %v", err)
		return []CachedURL{}
	}

	var response []CachedURL
	for _, url := range urls {
		response = append(response, CachedURL{
			Original:  url.Original,
			Clicks:    int(url.Clicks.Int64),
			CreatedAt: url.CreatedAt.Time,
			ExpiredAt: url.ExpiredAt.Time,
		})
	}

	return response
}

func (s *UrlService) DeleteExpiredURLs() error {
	expiredUrls, err := s.postgresClient.Queries.GetExpiredURLs(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to get expired URLs: %w", err)
	}

	for i := 0; i < len(expiredUrls); i += batchSize {
		end := i + batchSize
		if end > len(expiredUrls) {
			end = len(expiredUrls)
		}
		batch := expiredUrls[i:end]

		var wg sync.WaitGroup
		for _, url := range batch {
			wg.Add(1)
			go func(shortened string) {
				defer wg.Done()
				if err := s.redisClient.Del(s.ctx, shortened).Err(); err != nil {
					s.errorChan <- fmt.Errorf("failed to delete from cache: %s: %w", shortened, err)
				}
			}(url.Shortened)
		}
		wg.Wait()
	}

	if err := s.postgresClient.Queries.DeleteExpiredURLs(s.ctx); err != nil {
		return fmt.Errorf("failed to delete expired URLs from DB: %w", err)
	}

	return nil
}

func (s *UrlService) getFromCache(shortenedURL string) (*CachedURL, error) {
	data, err := s.redisClient.Get(s.ctx, shortenedURL).Result()
	if err != nil {
		return nil, err
	}

	var cachedURL CachedURL
	if err := json.Unmarshal([]byte(data), &cachedURL); err != nil {
		return nil, err
	}

	return &cachedURL, nil
}

func (s *UrlService) getFromDB(shortenedURL string) (*CachedURL, error) {
	url, err := s.postgresClient.Queries.GetOriginated(s.ctx, shortenedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get URL from database: %v", err)
	}

	var userIDStr string
	if url.UserID.Valid {
		userUUID, err := uuid.FromBytes(url.UserID.Bytes[:])
		if err != nil {
			return nil, fmt.Errorf("failed to convert UUID bytes: %v", err)
		}
		userIDStr = userUUID.String()
	}

	return &CachedURL{
		Original:  url.Original,
		Clicks:    int(url.Clicks.Int64),
		CreatedAt: url.CreatedAt.Time,
		ExpiredAt: url.ExpiredAt.Time,
		UserID:    userIDStr,
	}, nil
}
func (s *UrlService) setCache(shortenedURL string, url *CachedURL) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	return s.redisClient.Set(s.ctx, shortenedURL, data, s.cacheTimeout).Err()
}

func (s *UrlService) StartCacheSyncWorker(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			s.syncCacheWithDB()
		}
	}()
}

func (s *UrlService) syncCacheWithDB() {
	keys, err := s.redisClient.Keys(s.ctx, "*").Result()
	if err != nil {
		log.Errorf("Failed to get keys from Redis: %v", err)
		return
	}

	for _, key := range keys {
		cachedURL, err := s.getFromCache(key)
		if err != nil {
			continue
		}

		err = s.postgresClient.Queries.UpdateURL(s.ctx, sqlc.UpdateURLParams{
			Shortened: key,
			Clicks:    pgtype.Int8{Int64: int64(cachedURL.Clicks), Valid: true},
		})
		if err != nil {
			log.Errorf("Failed to sync URL %s to database: %v", key, err)
		}
	}
}

func (s *UrlService) handleErrors() {
	for err := range s.errorChan {
		log.Errorf("Async operation error: %v", err)
	}
}
