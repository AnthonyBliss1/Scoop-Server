package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/anthonybliss1/Scoop-Server/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	port := flag.Int("port", 2767, "http server port number")

	flag.Parse()

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/sync", handlers.ReadServerData)

	r.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Upload Scoop save data to sync server"))
	})

	fmt.Printf("[ Starting Scoop Server on Port %d ... ]\n", *port)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	http.ListenAndServe(addr, r)
}
