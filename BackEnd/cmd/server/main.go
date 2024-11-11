package main

import (
	"encoding/json"
	"flag"
	"net/http"
	_ "net/http/pprof"
	"shorten-url/backend/pkg/config"
	"shorten-url/backend/pkg/services"
	"shorten-url/backend/pkg/stores"
	"shorten-url/backend/pkg/utils"
	"strings"
	"sync"
	"time"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	_ "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

func main() {
	analyticsService := flag.Bool("analyticsService", false, "Track the number of link clicks")
	logging := flag.Bool("logging", false, "Enable Logging All Request")
	rateLimiting := flag.Bool("ratelimit", false, "Enable Rate Limiting")
	flag.Parse()
	services.NewUrlService(stores.RedisCluster, stores.PostgresClient)

	// Bypass Garbage Collection
	_ = make([]byte, 1<<30)
	// c := cron.New()



	// c.AddFunc("@hourly", func() {
	// 	err := services.UrlServiceInstance.DeleteExpiredURLs()
	// 	if err != nil {
	// 		log.Error(err)
	// 	}
	// })
	// c.Start()

	config.LoadEnv()
	stores.InitRedis("41943040", "volatile-lru")
	stores.InitPostgres()
	defer stores.PostgresClient.DB.Close()

	services.NewUrlService(stores.RedisCluster, stores.PostgresClient)
	r := chi.NewRouter()

	if *logging {
		r.Use(middleware.Logger)
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*", "ws://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(
		middleware.Maybe(middleware.StripSlashes, func(r *http.Request) bool {
			return !strings.HasPrefix(r.URL.Path, "/debug/")
		}),
	)

	r.Use(middleware.StripSlashes)
	r.Use(middleware.Recoverer)

	r.Get("/short/{id}", func(w http.ResponseWriter, r *http.Request) {
		shortenedURL := chi.URLParam(r, "id")
		if shortenedURL == "" {
			http.Error(w, "Missing ID", http.StatusBadRequest)
			return
		}
		// nem cai message vao trong rabbit queue
		

		originalURL, err := services.UrlServiceInstance.GetURL(shortenedURL)
		if err != nil || originalURL == nil {
			http.Error(w, "Not Found Your URL", http.StatusNotFound)
			return
		}

		if *analyticsService {
			updatedURL, err := services.UrlServiceInstance.IncrementClicks(shortenedURL)
			if err != nil {
				log.Error(err)
				originalURL = updatedURL
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"originalUrl": originalURL.Original,
		})
	})

	r.Group(func(r chi.Router) {
		if *rateLimiting {
			r.Use(httprate.Limit(
				1000,             
				10 * time.Second, 
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

			newUrl, err := services.UrlServiceInstance.CreateURL(url, userId)
			if err != nil {
				log.Errorf("Failed to create URL: %v", err)
				http.Error(w, "Failed to create URL", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"shortUrl":  newUrl.Shortened,
				"createdAt": newUrl.CreatedAt,
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

	var wg sync.WaitGroup

	portArray := config.AppConfig.Server.Ports
	for _, port := range portArray {
		wg.Add(1)
		go func(port string) {
			defer wg.Done()
			utils.StartServer(port, r)
		}(port)
	}

	wg.Wait()

	select {}
}
