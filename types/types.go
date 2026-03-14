package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/acme/autocert"
)

type Method string

const (
	Get    Method = "GET"
	Post   Method = "POST"
	Put    Method = "PUT"
	Patch  Method = "PATCH"
	Delete Method = "DELETE"
	Empty  Method = ""
)

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type Request struct {
	Method  Method `json:"method"`
	URL     string `json:"url"`
	Headers []KV   `json:"headers"`
	QParams []KV   `json:"query_params"`
	Body    string `json:"body"`
}

type Response struct {
	Status      string `json:"status"`
	StatusCode  int    `json:"status_code"`
	Headers     []KV   `json:"headers"`
	Body        string `json:"body"`
	Duration    int64  `json:"duration"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

type Scoop struct {
	Name     string   `json:"name"`
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Collection struct {
	Name   string  `json:"name"`
	Scoops []Scoop `json:"scoops"`
}

type DNSOverride struct {
	Variable string `json:"variable"`
	IPV4     string `json:"ipv4"`
}

type ServerPayload struct {
	Collections []Collection  `json:"collections"`
	DNS         []DNSOverride `json:"dns"`

	mu sync.Mutex
}

// all potential flags consolidated into this struct

type Options struct {
	Port      int
	Deploy    bool
	TLSMode   string
	Cert      string
	PKey      string
	Domain    string
	PrivateIP string
	ACManager *autocert.Manager
}

func (s *ServerPayload) PopulateCollections(path string) error {
	// Make the folder structure if it doesnt exist,
	// if path already exist then MkdirAll returns nil (does nothing)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}

	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, coll := range dirEntries {
		var c Collection

		file := filepath.Join(path, coll.Name())

		// file extension safeguard
		// (all collection files are json)
		ext := filepath.Ext(file)
		if ext != ".json" {
			continue
		}

		b, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, &c); err != nil {
			return err
		}

		// add to collections
		s.mu.Lock()
		s.Collections = append(s.Collections, c)
		s.mu.Unlock()
	}

	return nil
}

func (s *ServerPayload) WriteCollections(path string) error {
	// Make the folder structure if it doesnt exist,
	// if path already exist then MkdirAll returns nil (does nothing)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, c := range s.Collections {
		b, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			return err
		}

		fn := fmt.Sprintf("%s.json", c.Name)
		newFPath := filepath.Join(path, fn)

		if err := os.WriteFile(newFPath, b, 0o644); err != nil {
			return err
		}
	}

	return nil
}

func (s *ServerPayload) PopulateDNSOverrides(path string) error {
	// Make the folder structure if it doesnt exist,
	// if path already exist then MkdirAll returns nil (does nothing)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}

	// All DNS Overrides are stored in a single file
	path = filepath.Join(path, "overrides.json")

	b, err := os.ReadFile(path)
	if err != nil {
		// on error, make sure the file is created
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			return err
		}

		// instead of reading an empty file, will set b to an empty slice
		// which is handled below
		b = []byte{}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// empty file safeguard
	if len(b) == 0 {
		s.DNS = []DNSOverride{}
		return nil
	}

	if err := json.Unmarshal(b, &s.DNS); err != nil {
		return err
	}

	return nil
}

func (s *ServerPayload) WriteDNSOverrides(path string) error {
	// Make the folder structure if it doesnt exist,
	// if path already exist then MkdirAll returns nil (does nothing)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}

	// All DNS Overrides are stored in a single file
	path = filepath.Join(path, "overrides.json")

	s.mu.Lock()
	b, err := json.MarshalIndent(s.DNS, "", "  ")
	if err != nil {
		return err
	}
	s.mu.Unlock()

	// overwrite the server file with payload data
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return err
	}

	return nil
}
