package utils

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func StartServer(port string, router *chi.Mux) {	
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatal(err.Error())
	}
}
