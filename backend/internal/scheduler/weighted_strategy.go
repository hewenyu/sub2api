package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

type WeightedRoundRobinStrategy struct {
	mu               sync.Mutex
	currentWeights   map[int64]int
	effectiveWeights map[int64]int
}

func NewWeightedRoundRobinStrategy() SelectionStrategy {
	return &WeightedRoundRobinStrategy{
		currentWeights:   make(map[int64]int),
		effectiveWeights: make(map[int64]int),
	}
}

func (s *WeightedRoundRobinStrategy) Select(ctx context.Context, candidates []*model.CodexAccount, selectionCtx SelectionContext) (*model.CodexAccount, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available accounts")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	totalWeight := 0
	for _, account := range candidates {
		weight := account.Priority
		if weight <= 0 {
			weight = 1
		}

		if _, exists := s.effectiveWeights[account.ID]; !exists {
			s.effectiveWeights[account.ID] = weight
		}

		if _, exists := s.currentWeights[account.ID]; !exists {
			s.currentWeights[account.ID] = 0
		}

		s.currentWeights[account.ID] += s.effectiveWeights[account.ID]
		totalWeight += s.effectiveWeights[account.ID]
	}

	var selected *model.CodexAccount
	maxWeight := -1

	for _, account := range candidates {
		if s.currentWeights[account.ID] > maxWeight {
			maxWeight = s.currentWeights[account.ID]
			selected = account
		}
	}

	if selected != nil {
		s.currentWeights[selected.ID] -= totalWeight
	}

	return selected, nil
}
