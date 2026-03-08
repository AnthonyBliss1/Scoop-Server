package main

import (
	"flag"

	"github.com/anthonybliss1/Scoop-Server/cmd"
	"github.com/anthonybliss1/Scoop-Server/types"
	"github.com/anthonybliss1/Scoop-Server/utils"
	"github.com/fatih/color"
)

var red = color.New(color.FgRed)

func main() {
	// TODO: should use -tls-mode = off | manual | self | acme
	// manual would require user to provide cert and pkey paths
	// self would generate self signed cert to use
	// acme would use autocert for domains

	port := flag.Int("port", 2767, "http server port number")
	deploy := flag.Bool("deploy", false, "deploy systemd or launchd service")

	tlsMode := flag.String("tls-mode", "off", "off | manual | self | acme")

	var privateIP string
	cert := flag.String("cert", "", "tls cert file path")
	pKey := flag.String("key", "", "path to private key file")

	domain := flag.String("domain", "", "public domain for ACME")
	email := flag.String("email", "", "contact emailf for ACME")

	// help := flag.Bool("help", false, "show all flags and descriptions")

	flag.Parse()

	switch *tlsMode {
	case "off":
		// do nothing

	case "manual":
		if *cert == "" {
			red.Println("> -tsl-mode=manual requires -cert")
			return
		}

		if *pKey == "" {
			red.Println("> -tls-mode=manual requires -key")
			return
		}

		// validate cert and key provided by user
		if err := utils.ValidateCertAndKey(*cert, *pKey); err != nil {
			red.Printf("> Failed to validate certificate and key: %q\n", err)
			return
		}

	case "self":
		// generate self-signed cert and key
		cPath, kPath, pI, err := utils.GenerateSelfCertAndKey()
		if err != nil {
			red.Printf("> %s\n", err)
			return
		}

		*cert = cPath
		*pKey = kPath
		privateIP = pI

	case "acme":
		if *domain == "" {
			red.Println("> -tsl-mode=acme requires -domain")
			return
		}

		if *email == "" {
			red.Println("> -tsl-mode=acme requires -email")
			return
		}

	default:
		red.Printf("> Invalid use of -tls-mode: %q\n", *tlsMode)
		return
	}

	opts := types.Options{
		Port:      *port,
		Deploy:    *deploy,
		TLSMode:   *tlsMode,
		Cert:      *cert,
		PKey:      *pKey,
		Domain:    *domain,
		Email:     *email,
		PrivateIP: privateIP,
	}

	switch *deploy {
	case true:
		cmd.StartDeploy(opts)

	default:
		cmd.StartServer(opts)
	}
}
