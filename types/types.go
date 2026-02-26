package types

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
}

type Response struct {
	Status      string `json:"status"`
	StatusCode  int    `json:"status_code"`
	Headers     []KV   `json:"headers"`
	Body        string `json:"body"`
	Duration    int64  `json:"duration"`
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
}
