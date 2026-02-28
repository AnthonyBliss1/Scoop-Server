package cmd

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/anthonybliss1/Scoop-Server/handlers"
	"github.com/anthonybliss1/Scoop-Server/utils"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func StartServer(port *int) {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/sync", handlers.ReadServerData)
	r.Post("/upload", handlers.WriteServerData)

	fmt.Printf("[ Starting Scoop Server on Port %d ... ]\n", *port)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	http.ListenAndServe(addr, r)
}

func StartDeploy(port *int) {
	os := runtime.GOOS

	switch os {
	case "darwin":
		// TODO: Add macos support (launchD)
		fmt.Println("> LaunchD currently not supported! ")

	case "linux":
		fmt.Println("[ Linux OS identified ... ]")

		if err := utils.DeploySystemD(port); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("\n> Scoop-Service successfully deployed!")
	}
}
