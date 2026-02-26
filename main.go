package main

import (
	"fmt"
	"net/http"

	"github.com/anthonybliss1/Scoop-Server/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/sync", handlers.ReadServerData)

	r.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Upload Scoop save data to sync server"))
	})

	fmt.Println("[ Starting Scoop Server... ]")

	// should allow user to change the port
	http.ListenAndServe("0.0.0.0:2767", r)
}
