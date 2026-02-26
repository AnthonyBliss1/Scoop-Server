package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/anthonybliss1/Scoop-Server/types"
	"github.com/anthonybliss1/Scoop-Server/utils"
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

	// Grab All Collections
	collections, err := utils.GrabCollections(serverCollections)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Grab DNS Overrides
	dns, err := utils.GrabDNSOverrides(serverDNS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload := types.ServerPayload{Collections: collections, DNS: dns}

	b, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(b))
}
