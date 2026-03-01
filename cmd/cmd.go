package cmd

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/anthonybliss1/Scoop-Server/handlers"
	"github.com/anthonybliss1/Scoop-Server/utils"
	"github.com/fatih/color"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	green = color.New(color.FgGreen)
	red   = color.New(color.FgRed)
)

func StartServer(port *int) {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/sync", handlers.ReadServerData)
	r.Post("/upload", handlers.WriteServerData)

	text := `
░█▀▀░█▀▀░█▀█░█▀█░█▀█░░░█▀▀░█▀▀░█▀▄░█░█░█▀▀░█▀▄
░▀▀█░█░░░█░█░█░█░█▀▀░░░▀▀█░█▀▀░█▀▄░▀▄▀░█▀▀░█▀▄
░▀▀▀░▀▀▀░▀▀▀░▀▀▀░▀░░░░░▀▀▀░▀▀▀░▀░▀░░▀░░▀▀▀░▀░▀`

	color.Green(text)
	green.Printf("\n[ Starting Scoop Server on Port %d ... ]\n", *port)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	http.ListenAndServe(addr, r)
}

func StartDeploy(port *int) {
	os := runtime.GOOS

	switch os {
	case "darwin":
		// TODO: Add macos support (launchD)
		red.Println("> LaunchD currently not supported! ")

	case "linux":
		green.Println("[ Linux OS identified ... ]")

		if err := utils.DeploySystemD(port); err != nil {
			red.Println(err)
			return
		}

		green.Println("\n> Scoop-Service successfully deployed!")
	}
}
