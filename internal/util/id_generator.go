package util

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type IDGenerator struct {
	counter atomic.Int64
}

var (
	idGeneratorOnce sync.Once
	idGeneratorInst *IDGenerator
)

func NewIDGenerator(start int64) *IDGenerator {
	idGeneratorOnce.Do(func() {
		g := &IDGenerator{}
		g.counter.Store(start)
		idGeneratorInst = g
	})
	return idGeneratorInst
}

func (g *IDGenerator) Next(prefix string) string {
	n := g.counter.Add(1)
	return fmt.Sprintf("%s-%d", prefix, n)
}
