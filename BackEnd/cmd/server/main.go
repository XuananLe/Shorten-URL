package main

import (
	"encoding/json"
	"flag"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	log "github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
	"shorten-url/backend/pkg/services"
	"shorten-url/backend/pkg/stores"
	"shorten-url/backend/pkg/utils"
	"strings"
	"sync"
)

func main() {
	analyticsService := flag.Bool("analyticsService", false, "Track the number of link clicks")
	logging := flag.Bool("logging", false, "Enable Logging All Request")
	flag.Parse()
	utils.LoadEnv()
	stores.InitRedis("41943040", "volatile-lru")
	stores.InitPostgres()
	defer stores.PostgresClient.DB.Close()
	
	services.NewUrlService(stores.RedisCluster, stores.PostgresClient)
	r := chi.NewRouter()

	// Global Middleware
	if *logging {
		r.Use(middleware.Logger)
	}

	// TODO: Specific Origin from FE. Eg. Only Allow https://apm.shorten.com/....
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
		// TODO: Allow user to create at max 100 request per minute

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

	portArray := utils.Config.Server.ServerPort
	for _, port := range portArray {
		wg.Add(1)
		go func(port string) {
			defer wg.Done()
			utils.StartServer(port, r)
		}(port)
	}

	wg.Wait()

}
