package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	systemDTemp = `[Unit]
	Description=scoop-server
	After=network.target

	[Service]
	User=%s
	WorkingDirectory=%s
	ExecStart=%s
	Restart=on-failure

	[Install]
	WantedBy=multi-user.target
	`
)

var lxHome = regexp.MustCompile(`^/home/([^/]+)/`)

func DeploySystemD(port *int) error {
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

	// add port to exec path
	bPath = fmt.Sprintf("%s -port=%d", bPath, *port)

	bDir := filepath.Dir(bPath)

	fmt.Println("[ Binary path identified ... ]")

	unitText := fmt.Sprintf(systemDTemp, user, bDir, bPath)

	fmt.Println("[ Creating unit file ... ]")

	servicePath := "/etc/systemd/system/scoop-server.service"
	if err := os.WriteFile(servicePath, []byte(unitText), 0o644); err != nil {
		return err
	}

	fmt.Println("[ Reloading systemctl daemon ... ]")

	if out, err := exec.Command("systemctl", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("daemon-reload failed: %q (%s)", err, string(out))
	}

	fmt.Println("[ Enabling service ... ]")

	if out, err := exec.Command("systemctl", "enable", "scoop-server.service").CombinedOutput(); err != nil {
		if !strings.Contains(string(out), "is enabled") {
			return fmt.Errorf("enable failed: %q (%s)", err, string(out))
		}
	}

	fmt.Println("[ Restarting service ... ]")

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
			fmt.Println("> Please enter the user below: ")

			fmt.Print("> ")
			u, err := ScanUser(scanner)
			if err != nil {
				return err
			}

			*user = u

		default:
			fmt.Println("> Please confirm the user")
			fmt.Printf("> Is %s correct? (y/n)\n", *user)

			fmt.Print("> ")
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
