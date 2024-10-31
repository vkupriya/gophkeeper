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

const (
	UnknownSecret int32 = iota
	TextSecret
	BinarySecret
	CardSecret
	FileSecret
)

func TypeToProto(st string) int32 {
	switch st {
	case "text":
		return TextSecret
	case "binary":
		return BinarySecret
	case "card":
		return CardSecret
	case "file":
		return FileSecret
	case "unknown":
		return UnknownSecret
	default:
		return UnknownSecret
	}
}

func ProtoToType(i int32) string {
	switch i {
	case TextSecret:
		return "text"
	case BinarySecret:
		return "binary"
	case CardSecret:
		return "card"
	case FileSecret:
		return "file"
	case UnknownSecret:
		return "unknown"
	default:
		return "unknown"
	}
}
