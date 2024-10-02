package main

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Anh yeu em vcl \n"))
		w.Write([]byte("OK"))
    })
	err := http.ListenAndServe(":3000", r);
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("\"Serving at port\": %v\n", "Serving at port" + string(3000))
	}
}