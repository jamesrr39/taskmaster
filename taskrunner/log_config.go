package taskrunner

type Compression string

const (
	CompressionNone Compression = "none"
	CompressionZstd Compression = "zstd"
)

type LogConfig struct {
	Compression Compression `json:"compression" required:"true"`
}
