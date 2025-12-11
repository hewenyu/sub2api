package scheduler

import (
	"context"
	"fmt"
	"hash/fnv"
	"slices"
	"sort"
	"sync"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

type ConsistentHashStrategy struct {
	mu       sync.RWMutex
	ring     []uint32
	accounts map[uint32]*model.CodexAccount
	replicas int
}

func NewConsistentHashStrategy() SelectionStrategy {
	return &ConsistentHashStrategy{
		ring:     make([]uint32, 0),
		accounts: make(map[uint32]*model.CodexAccount),
		replicas: 150,
	}
}

func (s *ConsistentHashStrategy) Select(ctx context.Context, candidates []*model.CodexAccount, selectionCtx SelectionContext) (*model.CodexAccount, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available accounts")
	}

	s.mu.Lock()
	s.buildRing(candidates)
	s.mu.Unlock()

	key := selectionCtx.SessionHash
	if key == "" {
		key = fmt.Sprintf("apikey-%d", selectionCtx.APIKey.ID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	hash := s.hashKey(key)
	idx := sort.Search(len(s.ring), func(i int) bool {
		return s.ring[i] >= hash
	})

	if idx == len(s.ring) {
		idx = 0
	}

	return s.accounts[s.ring[idx]], nil
}

func (s *ConsistentHashStrategy) buildRing(candidates []*model.CodexAccount) {
	s.ring = make([]uint32, 0)
	s.accounts = make(map[uint32]*model.CodexAccount)

	for _, account := range candidates {
		for i := 0; i < s.replicas; i++ {
			key := fmt.Sprintf("%d-%d", account.ID, i)
			hash := s.hashKey(key)
			s.ring = append(s.ring, hash)
			s.accounts[hash] = account
		}
	}

	slices.Sort(s.ring)
}

func (s *ConsistentHashStrategy) hashKey(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}
