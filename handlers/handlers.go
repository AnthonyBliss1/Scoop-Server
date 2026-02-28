package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/anthonybliss1/Scoop-Server/types"
)

func ReadServerData(w http.ResponseWriter, r *http.Request) {
	base, err := os.UserConfigDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Search Scoop-Server Config
	// Using the actual Scoop dir for testing
	// will change to Scoop-Server/Collections & /DNS
	serverCollections := filepath.Join(base, "Scoop-Server", "Collections")
	serverDNS := filepath.Join(base, "Scoop-Server", "DNS")

	var payload types.ServerPayload

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(2)

	// Grab All Collections
	go func() {
		defer wg.Done()
		if err := payload.PopulateCollections(serverCollections); err != nil {
			errCh <- err
		}
	}()

	// Grab DNS Overrides
	go func() {
		defer wg.Done()
		if err := payload.PopulateDNSOverrides(serverDNS); err != nil {
			errCh <- err
		}
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(b))
}

func WriteServerData(w http.ResponseWriter, r *http.Request) {
	base, err := os.UserConfigDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Search Scoop-Server Config
	// Using the actual Scoop dir for testing
	// will change to Scoop-Server/Collections & /DNS
	serverCollections := filepath.Join(base, "Scoop-Server", "Collections")
	serverDNS := filepath.Join(base, "Scoop-Server", "DNS")

	var payload types.ServerPayload

	// read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// unmarshal to payload struct
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(2)

	// write collections data to server files
	go func() {
		defer wg.Done()
		if err := payload.WriteCollections(serverCollections); err != nil {
			errCh <- err
		}
	}()

	go func() {
		defer wg.Done()
		if err := payload.WriteDNSOverrides(serverDNS); err != nil {
			errCh <- err
		}
	}()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server files sucessfully updated"))
}
