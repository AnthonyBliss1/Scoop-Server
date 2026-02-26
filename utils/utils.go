package utils

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/anthonybliss1/Scoop-Server/types"
)

func GrabCollections(path string) (Collections []types.Collection, err error) {
	// Make the folder structure if it doesnt exist,
	// if path already exist then MkdirAll returns nil (does nothing)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, err
	}

	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, coll := range dirEntries {
		var c types.Collection

		file := filepath.Join(path, coll.Name())

		// file extension safeguard
		// (all collection files are json)
		ext := filepath.Ext(file)
		if ext != ".json" {
			continue
		}

		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(b, &c); err != nil {
			return nil, err
		}

		// add to collections
		Collections = append(Collections, c)
	}

	return Collections, nil
}

func GrabDNSOverrides(path string) (DNS []types.DNSOverride, err error) {
	// Make the folder structure if it doesnt exist,
	// if path already exist then MkdirAll returns nil (does nothing)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, err
	}

	// All DNS Overrides are stored in a single file
	path = filepath.Join(path, "overrides.json")

	b, err := os.ReadFile(path)
	if err != nil {
		// on error, make sure the file is created
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			return nil, err
		}

		// attempt read again
		b, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}

	// empty file safeguard
	if len(b) == 0 {
		return []types.DNSOverride{}, nil
	}

	if err := json.Unmarshal(b, &DNS); err != nil {
		return nil, err
	}

	return DNS, nil
}
