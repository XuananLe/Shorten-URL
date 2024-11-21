package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"shorten-url/backend/pkg/config"
	"shorten-url/backend/pkg/services"
	"shorten-url/backend/pkg/stores"
	"shorten-url/backend/pkg/utils"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/robfig/cron"
	_ "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	chitrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go-chi/chi.v5"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type FeatureFlags struct {
	AnalyticsService bool `json:"analyticsService"`
	Logging          bool `json:"logging"`
	Monitoring       bool `json:"monitoring"`
	RateLimiting     bool `json:"rateLimiting"`
}

func loadFeatureFlags(configFile string) (FeatureFlags, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return FeatureFlags{}, err
	}
	var flags FeatureFlags
	err = json.Unmarshal(data, &flags)
	return flags, err
}

func main() {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 10000;
	flags, err := loadFeatureFlags("feature.json")
	if err != nil {
		log.Fatalf("Failed to load feature flags: %v", err)
	}

	if flags.AnalyticsService {
		fmt.Println("Analytics Service is enabled.")
	} else {
		fmt.Println("Analytics Service is disabled.")
	}
	if flags.Logging {
		fmt.Println("Logging is enabled.")
	} else {
		fmt.Println("Logging is disabled.")
	}

	if flags.Monitoring {
		fmt.Println("Monitoring is enabled.")
		tracer.Start()
		defer tracer.Stop()
	} else {
		fmt.Println("Monitoring is disabled.")
	}
	if flags.RateLimiting {
		fmt.Println("Rate Limiting is enabled.")
	} else {
		fmt.Println("Rate Limiting is disabled.")
	}

	port := flag.String("port", "3002", "Port to run the server on")

	flag.Parse()

	_ = make([]byte, 1<<30)
	c := cron.New()

	c.AddFunc("@hourly", func() {
		err := services.UrlServiceInstance.DeleteExpiredURLs()
		if err != nil {
			log.Error(err)
		}
	})
	c.Start()
	defer c.Stop()

	config.LoadEnv()
	stores.InitRedis("2gb", "volatile-lru")
	stores.InitPostgres()
	stores.InitRabbitMQ()
	stores.RabbitMQClient.DeclareQueue("queue-based-load-leveling-" + *port)
	services.NewUrlService(stores.RedisCluster, stores.PostgresClient, stores.RabbitMQClient)


	defer stores.PostgresClient.DB.Close()
	defer stores.RedisCluster.Close()
	defer stores.CloseRabbitMQ()

	numConsumers := 10
	for i := 1; i <= numConsumers; i++ {
		go services.UrlServiceInstance.ProcessQueueBatch(*port,fmt.Sprintf("consumer-%d", i), 100, 5*time.Second)
	}

	r := chi.NewRouter()

	if flags.Logging {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)

	if flags.Monitoring {
		r.Use(chitrace.Middleware(chitrace.WithServiceName("chi-server")))
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*", "ws://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.StripSlashes)

	r.Get("/short/{id}", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now();
		shortenedURL := chi.URLParam(r, "id")
		if shortenedURL == "" {
			http.Error(w, "Missing ID", http.StatusBadRequest)
			return
		}

		originalURL, err := services.UrlServiceInstance.GetURL(shortenedURL)
		if err != nil || originalURL == nil {
			http.Error(w, "Not Found Your URL", http.StatusNotFound)
			return
		}

		if flags.AnalyticsService {
			updatedURL, err := services.UrlServiceInstance.IncrementClicks(shortenedURL)
			if err != nil {
				log.Error(err)
				originalURL = updatedURL
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if (time.Since(start) >= 100 * time.Millisecond) {
			log.Fatal("Damn boi, we timed out")
		}
		json.NewEncoder(w).Encode(map[string]any{
			"originalUrl": originalURL.Original,
		})
	})

	r.Group(func(r chi.Router) {
		if flags.RateLimiting {
			r.Use(httprate.Limit(
				1000,
				10*time.Second,
				httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
			))
		}

		r.Post("/create", func(w http.ResponseWriter, r *http.Request) {
			url := r.URL.Query().Get("url")
			userId := r.URL.Query().Get("userId")
			if url == "" {
				http.Error(w, "Missing URL parameter", http.StatusBadRequest)
				return
			}
			if userId == "" {
				http.Error(w, "Missing UserId parameter", http.StatusBadRequest)
				return
			}
			newUrl, err := services.UrlServiceInstance.CreateURL(*port, url, userId)
			if err != nil {
				log.Errorf("Failed to create URL: %v", err)
				http.Error(w, "Failed to create URL", http.StatusInternalServerError)
				log.Fatal(err)
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"shortUrl": newUrl,
			})
		})
	})

	r.Get("/history/{userId}", func(w http.ResponseWriter, r *http.Request) {
		userId := chi.URLParam(r, "userId")
		if userId == "" {
			http.Error(w, "Missing UserId parameter", http.StatusBadRequest)
			return
		}

		urls := services.UrlServiceInstance.GetURLs(userId)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(urls)
	})

	r.Delete("/short/{id}", func(w http.ResponseWriter, r *http.Request) {
		shortenedURL := chi.URLParam(r, "id")
		if shortenedURL == "" {
			http.Error(w, "Missing ID", http.StatusBadRequest)
			return
		}

		err := services.UrlServiceInstance.DeleteURL(shortenedURL)
		if err != nil {
			http.Error(w, "Failed to delete URL", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("URL deleted successfully"))
	})

	r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		var user services.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		log.Printf("Received userId: %s", user.UserID)
		if err := services.UrlServiceInstance.CreateUser(user.UserID); err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write([]byte("User ID created successfully"))
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Route does not exist"))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(405)
		w.Write([]byte("Method is not valid"))
	})

	utils.StartServer(*port, r)

	select {}
}
