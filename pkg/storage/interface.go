// pkg/storage/interface.go
package storage

import (
	"context"
	"io"

	"github.com/turtacn/agenticai/pkg/types"
)

type StoreType string

const (
	StoreTypeS3       StoreType = "s3"
	StoreTypeMinIO    StoreType = "minio"
	StoreTypeRedis    StoreType = "redis"
	StoreTypeMilvus   StoreType = "milvus"
	StoreTypeQdrant   StoreType = "qdrant"
	StoreTypePostgres StoreType = "postgres"
)

type Config struct {
	Type   StoreType
	Params map[string]string // host, bucket, collection...
}

type Store interface {
	Reader(ctx context.Context, key string) (io.ReadCloser, error)
	Writer(ctx context.Context, key string) (io.WriteCloser, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]string, error)
	Close() error
}

type VectorStore interface {
	Store
	Upsert(ctx context.Context, vec *types.Vector) error
	Search(ctx context.Context, vec []float32, k int) ([]*types.SearchResult, error)
}
//Personal.AI order the ending
