package handlers

import (
	"encoding/json"
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
	serverCollections := filepath.Join(base, "Scoop", "Collections")
	serverDNS := filepath.Join(base, "Scoop", "DNS")

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
