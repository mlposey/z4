package storage

import (
	"fmt"
	"github.com/mlposey/z4/iden"
	"time"
)

type IDGenerator struct {
	indexes *IndexStore
}

func NewGenerator(indexes *IndexStore) *IDGenerator {
	return &IDGenerator{
		indexes: indexes,
	}
}

func (g *IDGenerator) ID(queue string, ts time.Time) (iden.TaskID, error) {
	idx, err := g.indexes.Next(queue)
	if err != nil {
		return iden.TaskID{}, fmt.Errorf("failed to get next index: %w", err)
	}
	return iden.New(ts, idx), nil
}
