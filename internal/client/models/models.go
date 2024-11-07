package models

type Secret struct {
	Name    string
	Type    string
	Meta    string
	Data    []byte
	Version int64
}

type SecretPrint struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Meta    string `json:"meta"`
	Data    string `json:"data"`
	Version int64  `json:"version"`
}

type SecretList []SecretItem

type SecretItem struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version int64  `json:"version"`
}
