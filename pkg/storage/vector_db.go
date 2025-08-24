// pkg/storage/vector_db.go
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/turtacn/agenticai/pkg/types"
)

type vectClient struct {
	storeType StoreType
	conn      interface{} // *milvus/grpc.Client | *qdrant.Client
}

func NewVectorStore(cfg Config) (VectorStore, error) {
	switch cfg.Type {
	case StoreTypeMilvus:
		return newMilvus(cfg)
	case StoreTypeQdrant:
		return newQdrant(cfg)
	default:
		return nil, errors.New("unsupported vector store type")
	}
}

func newMilvus(cfg Config) (VectorStore, error) {
	// 连接 & 检查
	return &vectClient{storeType: StoreTypeMilvus}, nil // stub
}

func newQdrant(cfg Config) (VectorStore, error) {
	return &vectClient{storeType: StoreTypeQdrant}, nil // stub
}

func (v *vectClient) Upsert(ctx context.Context, vec *types.Vector) error {
	// switch by v.storeType ...
	return nil
}

func (v *vectClient) Search(ctx context.Context, vec []float32, k int) ([]*types.SearchResult, error) {
	return nil, fmt.Errorf("not implemented")
}

/* 下面实现其他 Store interface 空方法以快速编译通过 */
func (v *vectClient) Reader(_ context.Context, _ string) (io.ReadCloser, error)   { return nil, nil }
func (v *vectClient) Writer(_ context.Context, _ string) (io.WriteCloser, error)  { return nil, nil }
func (v *vectClient) Delete(_ context.Context, _ string) error                    { return nil }
func (v *vectClient) List(_ context.Context, _ string) ([]string, error)          { return nil, nil }
func (v *vectClient) Close() error {
	switch t := v.conn.(type) {
	case io.Closer:
		return t.Close()
	default:
		return nil
	}
}
//Personal.AI order the ending
