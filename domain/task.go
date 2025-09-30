package domain

type Compression string

const (
	CompressionNone Compression = "none"
	CompressionZstd Compression = "zstd"
)

type LogConfig struct {
	Compression Compression `json:"compression" required:"true"`
}

type Task struct {
	Name   string    `json:"name" required:"true"`
	Script string    `json:"script" required:"true"`
	Log    LogConfig `json:"log" required:"true"`
}
