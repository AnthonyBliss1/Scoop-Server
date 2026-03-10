package utils

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/anthonybliss1/Scoop-Server/types"
	"github.com/fatih/color"
	"golang.org/x/crypto/acme/autocert"
)

const (
	systemDTemp = `[Unit]
Description=scoop-server
After=network.target

[Service]
User=%s
WorkingDirectory=%s
ExecStart=%s
StandardOutput=append:%s
StandardError=append:%s
Restart=on-failure

[Install]
WantedBy=multi-user.target
`
)

var (
	green = color.New(color.FgGreen)
	blue  = color.New(color.FgBlue)

	lxHome = regexp.MustCompile(`^/home/([^/]+)/`)
)

func DeploySystemD(o types.Options) error {
	bPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find binary path: %q", err)
	}

	var user string
	if m := lxHome.FindStringSubmatch(bPath); len(m) == 2 {
		user = m[1]
	}

	fmt.Println()
	if err := ConfirmUser(&user); err != nil {
		return err
	}
	fmt.Println()

	// add custom port to exec path
	if o.Port != 2767 {
		bPath = fmt.Sprintf("%s -port=%d", bPath, o.Port)
	}

	// modify exec start command if tls is enabled
	// since the 'self' options creates the cert and key before this deploy step,
	// the exec function can just be manual and pointed to the created cert and key
	// instead of creating the cert and key everytime the service is started
	if o.TLSMode == "manual" || o.TLSMode == "self" {
		bPath = fmt.Sprintf("%s -tls-mode=manual -cert=%s -key=%s", bPath, o.Cert, o.PKey)
	}

	if o.TLSMode == "acme" {
		bPath = fmt.Sprintf("%s -tls-mode=acme -domain=%s", bPath, o.Domain)
	}

	bDir := filepath.Dir(bPath)

	green.Println("[ Binary path identified ... ]")

	logPath := "/var/log/scoop-server/scoop-server.log"
	unitText := fmt.Sprintf(systemDTemp, user, bDir, bPath, logPath, logPath)

	green.Println("[ Creating unit file ... ]")

	servicePath := "/etc/systemd/system/scoop-server.service"
	if err := os.WriteFile(servicePath, []byte(unitText), 0o644); err != nil {
		return err
	}

	green.Println("[ Reloading systemctl daemon ... ]")

	if out, err := exec.Command("systemctl", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("daemon-reload failed: %q (%s)", err, string(out))
	}

	green.Println("[ Enabling service ... ]")

	if out, err := exec.Command("systemctl", "enable", "scoop-server.service").CombinedOutput(); err != nil {
		if !strings.Contains(string(out), "is enabled") {
			return fmt.Errorf("enable failed: %q (%s)", err, string(out))
		}
	}

	green.Println("[ Restarting service ... ]")

	if out, err := exec.Command("systemctl", "restart", "scoop-server.service").CombinedOutput(); err != nil {
		return fmt.Errorf("restart failed: %q (%s)", err, string(out))
	}

	return nil
}

func ConfirmUser(user *string) error {
	scanner := bufio.NewScanner(os.Stdin)

	var ok bool

	for !ok {
		switch *user {
		case "":
			blue.Println("> Please enter the user below: ")

			blue.Print("> ")
			u, err := ScanUser(scanner)
			if err != nil {
				return err
			}

			*user = u

		default:
			blue.Println("> Please confirm the user")
			blue.Printf("> Is %s correct? (y/n)\n", *user)

			blue.Print("> ")
			ok = ScanConfirm(scanner)

			if !ok {
				*user = ""
			}
		}
	}

	return nil
}

func ScanUser(scanner *bufio.Scanner) (u string, err error) {
	if scanner.Scan() {
		u = scanner.Text()
		u = strings.TrimSpace(u)

		if _, err := user.Lookup(u); err != nil {
			return "", err
		}
	}

	return u, nil
}

func ScanConfirm(scanner *bufio.Scanner) bool {
	var input string

	if scanner.Scan() {
		input = scanner.Text()
		input = strings.ToLower(input)
	}

	return input == "y"
}

func GenerateSelfCertAndKey() (cPath string, kPath string, localIP string, err error) {
	localIP = GetLocalIP().String()

	hosts := []string{"localhost", "127.0.0.1", localIP}
	cPath = "cert.pem"
	kPath = "key.pem"

	var dnsNames []string
	var ipAddrs []net.IP

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			ipAddrs = append(ipAddrs, ip)
		} else {
			dnsNames = append(dnsNames, h)
		}
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate privateKey: %q", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(10, 0, 0) // valid for 10 years

	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate serialNumber: %q", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		DNSNames:    dnsNames,
		IPAddresses: ipAddrs,

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	b, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate Certificate: %q", err)
	}

	certOutput, err := os.Create(cPath)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create Certificate Output: %q", err)
	}
	defer certOutput.Close()

	if err := pem.Encode(certOutput, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: b,
	}); err != nil {
		return "", "", "", fmt.Errorf("failed to encode certificate PEM: %q", err)
	}

	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to marshal private key: %q", err)
	}

	keyOutput, err := os.OpenFile(kPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to open key file: %q", err)
	}
	defer keyOutput.Close()

	if err := pem.Encode(keyOutput, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	}); err != nil {
		return "", "", "", fmt.Errorf("failed to encode key PEM: %q", err)
	}

	// validate cert / key that were created
	if err := ValidateCertAndKey(cPath, kPath); err != nil {
		return "", "", "", fmt.Errorf("cannot validate generated cert/key pair: %q", err)
	}

	return cPath, kPath, localIP, nil
}

func ConfigureAutoCert(domain string) *autocert.Manager {
	return &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache("certs"),
	}
}

func ValidateCertAndKey(certPath string, keyPath string) error {
	_, err := tls.LoadX509KeyPair(certPath, keyPath)

	return err
}

func GetLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}
