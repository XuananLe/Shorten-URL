package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"shorten-url/backend/pkg/services"
	"shorten-url/backend/pkg/stores"
	"shorten-url/backend/pkg/utils"
	"strconv"
	log "github.com/sirupsen/logrus"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	utils.LoadEnv()
	stores.InitRedis("41943040", "volatile-lru")
	stores.InitPostgres()
	services.NewUrlService(stores.RedisClient, stores.PostgresClient)
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.Logger)

	// TODO: Specific Origin from FE. Eg. Only Allow https://apm.shorten.com/....
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(middleware.StripSlashes)
	r.Use(middleware.Recoverer)

	// MAIN ROUTE
	r.Get("/short/{id}", func(w http.ResponseWriter, r *http.Request) {
		shortenedURL := chi.URLParam(r, "id")
		if shortenedURL == "" {
			http.Error(w, "Missing ID", http.StatusBadRequest)
			return
		}

		originalURL, _ := services.UrlServiceInstance.GetURL(shortenedURL)

		if originalURL == nil {
			http.Error(w, "Not Found Your URL", http.StatusNotFound)
			return
		}

		if err := services.UrlServiceInstance.IncrementClicks(shortenedURL); err != nil {
			log.Error(err)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]string{
			"originalUrl": originalURL.Original,
		})
		// http.Redirect(w, r, originalURL.Original, http.StatusSeeOther)
	})

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

		shortenedID := utils.Hash(url)
		err := services.UrlServiceInstance.CreateURL(shortenedID, url, userId)
		if err != nil {
			log.Errorf("Failed to create URL: %v", err)
			http.Error(w, "Failed to create URL", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"shortUrl": shortenedID,
		})
	})

	r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		var user services.User;
		err := json.NewDecoder(r.Body).Decode(&user);
		if err != nil {
            http.Error(w, "Invalid request payload", http.StatusBadRequest)
            return
        }
        log.Printf("Received userId: %s", user.UserID)
		if err := services.UrlServiceInstance.CreateUser(user.UserID); err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest);
			return;
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
	err := http.ListenAndServe(":"+utils.Config.Server.SERVER_PORT, r)
	if err != nil {
		log.Fatal(err)
	} 
	fmt.Printf("\"Serving at port\": %v\n", "Serving at port "+ strconv.Itoa(3000))
	
}
