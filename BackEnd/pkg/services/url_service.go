package services

import (
	"context"
	"encoding/json"
	"fmt"
	"shorten-url/backend/pkg/db/sqlc"
	"shorten-url/backend/pkg/stores"
	"shorten-url/backend/pkg/utils"
	"sync"
	"time"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

type User struct {
	UserID string `json:"userId"`
}
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
	cacheMutex sync.Mutex

}

var UrlServiceInstance *UrlService = &UrlService{}

func NewUrlService(redisClient *redis.ClusterClient, postgresClient *stores.Postgres) *UrlService {
	UrlServiceInstance = &UrlService{
		ctx:            context.Background(),
		cacheTimeout:   24 * time.Hour,
		redisClient:    redisClient,
		postgresClient: postgresClient,
	}
	return UrlServiceInstance
}

func (s *UrlService) GetURL(shortenedURL string) (*CachedURL, error) {
	cachedData, err := s.getFromCache(shortenedURL)
	if err == nil {
		return cachedData, nil
	} 

	url, err := s.getFromDB(shortenedURL)
	if err != nil {
		return nil, err
	}

	if err := s.setCache(shortenedURL, url); err != nil {
		log.Errorf("Failed to set cache for %s: %v", shortenedURL, err)
	}

	return url, nil
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
		}
	}()

	return updatedURL, nil
}

func (s *UrlService) CreateURL(originalURL string, userIDStr string) (sqlc.Url, error) {
	baseHash := utils.Hash(originalURL)
	shortenedURL := baseHash
	counter := 0
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return sqlc.Url{}, fmt.Errorf("invalid user ID: %v", err)
	}
	for {
		urlParams := sqlc.InsertURLParams{
			Shortened: shortenedURL,
			Original:  originalURL,
			UserID: pgtype.UUID{
				Bytes: userID,
				Valid: true,
			},
		}
		newURL, err := s.postgresClient.Queries.InsertURL(s.ctx, urlParams)
		if err != nil {
			if utils.IsPgUniqueViolation(err) {
				counter++
				shortenedURL = fmt.Sprintf("%s%d", baseHash, counter)
				continue 
			}
			return sqlc.Url{}, fmt.Errorf("failed to create URL in database: %v", err)
		}
		cachedURL := &CachedURL{
			Original:  originalURL,
			Clicks:    0,
			CreatedAt: time.Now(),
			UserID:    userIDStr,
		}
		
		if err := s.setCache(shortenedURL, cachedURL); err != nil {
			log.Errorf("Failed to set cache for new URL %s: %v", shortenedURL, err)
		}
		return newURL, nil
	}
}



func (s *UrlService) CreateUser(userIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	err = s.postgresClient.Queries.InsertUser(context.Background(), utils.ConvertFromUuidPg(userID))

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

func (s *UrlService) DeleteExpiredURLs() error {
	expiredUrls, err := s.postgresClient.Queries.GetExpiredURLs(s.ctx)
	if err != nil {
		log.Errorf("Failed to get expired URLs: %v", err)
	} else {
		for _, url := range expiredUrls {
			err := s.redisClient.Del(s.ctx, url.Shortened).Err()
			if err != nil {
				log.Errorf("Failed to delete expired URL from cache: %v", err)
			}
		}
	}
	err = s.postgresClient.Queries.DeleteExpiredURLs(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to delete expired URLs: %v", err)
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


