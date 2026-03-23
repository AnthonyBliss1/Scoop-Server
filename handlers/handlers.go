package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/anthonybliss1/Scoop-Server/types"
	"github.com/anthonybliss1/Scoop-Server/utils"
)

var (
	key        string
	handlersMU *sync.Mutex
)

func init() {
	key, _ = utils.GetAPIKey()
}

func APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")

		if apiKey == "" || apiKey != key {
			http.Error(w, "Unauthorized: Invalid or missing API key", http.StatusUnauthorized)
			return
		}

		// proceed
		next.ServeHTTP(w, r)
	})
}

func ReadServerData(w http.ResponseWriter, r *http.Request) {
	base, err := os.UserConfigDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Search Scoop-Server Config
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(b))
}

func WriteServerData(w http.ResponseWriter, r *http.Request) {
	handlersMU = &sync.Mutex{}

	base, err := os.UserConfigDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Search Scoop-Server Config
	serverCollections := filepath.Join(base, "Scoop-Server", "Collections")
	serverDNS := filepath.Join(base, "Scoop-Server", "DNS")

	// wipe just coll dir (dns already does an overwrite)
	// this would obv not scale, will come back to this and do some tmp dir indiana jones switch
	// mutex is a bit of a bandaid here

	handlersMU.Lock()
	defer handlersMU.Unlock()

	if err := os.RemoveAll(serverCollections); err != nil {
		// wont handle the error if its a does not exist error
		if !os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

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

func CheckHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server healthy"))
}
