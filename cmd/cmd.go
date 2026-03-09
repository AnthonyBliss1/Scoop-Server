package cmd

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/anthonybliss1/Scoop-Server/handlers"
	"github.com/anthonybliss1/Scoop-Server/types"
	"github.com/anthonybliss1/Scoop-Server/utils"

	"github.com/fatih/color"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	green = color.New(color.FgGreen)
	red   = color.New(color.FgRed)
)

func StartServer(o types.Options) {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/sync", handlers.ReadServerData)
	r.Post("/upload", handlers.WriteServerData)

	text := `
░█▀▀░█▀▀░█▀█░█▀█░█▀█░░░█▀▀░█▀▀░█▀▄░█░█░█▀▀░█▀▄
░▀▀█░█░░░█░█░█░█░█▀▀░░░▀▀█░█▀▀░█▀▄░▀▄▀░█▀▀░█▀▄
░▀▀▀░▀▀▀░▀▀▀░▀▀▀░▀░░░░░▀▀▀░▀▀▀░▀░▀░░▀░░▀▀▀░▀░▀`

	color.Green(text)

	addr := fmt.Sprintf("0.0.0.0:%d", o.Port)

	switch o.TLSMode {
	case "off":
		green.Printf("\n[ Scoop Server running on Port %d ... ]\n", o.Port)
		http.ListenAndServe(addr, r)

	case "manual":
		green.Println("\n[ TLS enabled ... ]")
		green.Printf("[ Scoop Server running on Port %d ... ]\n", o.Port)
		http.ListenAndServeTLS(addr, o.Cert, o.PKey, r)

	case "self":
		// generated cert and key paths are written to o.Cert and o.PKey
		green.Println("\n[ TLS enabled ... ]")
		green.Printf("[ Scoop Server running at https://%s:%d ... ]\n", o.PrivateIP, o.Port)
		http.ListenAndServeTLS(addr, o.Cert, o.PKey, r)

	case "acme":
		green.Println("\n[ TLS managed with autocert ... ]")

		// start server on port 80 to handle ACME challenge
		go func() {
			if err := http.ListenAndServe(":80", o.ACManager.HTTPHandler(nil)); err != nil {
				red.Printf("> %q\n", err)
				return
			}
		}()

		s := &http.Server{
			Addr:    ":443",
			Handler: r,
			TLSConfig: &tls.Config{
				GetCertificate: o.ACManager.GetCertificate,
			},
		}

		green.Printf("[ Scoop Server running on Port 443 ... ]\n")
		if err := s.ListenAndServeTLS("", ""); err != nil {
			red.Printf("> %q\n", err)
			return
		}

	// this should never fire
	default:
		log.Fatalf("Failed assertion for -tls-mode: %q\n", o.TLSMode)
	}
}

func StartDeploy(o types.Options) {
	os := runtime.GOOS

	switch os {
	case "darwin":
		// TODO: Add macos support (launchD)
		red.Println("> LaunchD currently not supported! ")
		return

	case "linux":
		green.Println("[ Linux OS identified ... ]")

		if err := utils.DeploySystemD(o); err != nil {
			red.Println(err)
			return
		}

		green.Println("\n> Scoop-Service successfully deployed!")
		return
	}
}
