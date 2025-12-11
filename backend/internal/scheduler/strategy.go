package scheduler

import (
	"context"
	"fmt"
	"sort"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
)

// SelectionStrategy defines the interface for account selection strategies.
type SelectionStrategy interface {
	Select(ctx context.Context, candidates []*model.CodexAccount, selectionCtx SelectionContext) (*model.CodexAccount, error)
}

// SelectionContext provides context for the selection decision.
type SelectionContext struct {
	APIKey         *model.APIKey
	SessionHash    string
	RequestedModel string
}

// PriorityStrategy implements multi-level sorting selection strategy.
// Sorting order:
//  1. Priority (descending) - higher priority first
//  2. Concurrency (ascending) - lower concurrency first
//  3. LastUsedAt (ascending) - least recently used first
type PriorityStrategy struct {
	concurrencyRepo redis.ConcurrencyRepository
}

// NewPriorityStrategy creates a new priority-based selection strategy.
func NewPriorityStrategy(concurrencyRepo redis.ConcurrencyRepository) SelectionStrategy {
	return &PriorityStrategy{
		concurrencyRepo: concurrencyRepo,
	}
}

func (s *PriorityStrategy) Select(ctx context.Context, candidates []*model.CodexAccount, selectionCtx SelectionContext) (*model.CodexAccount, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available accounts")
	}

	// Get concurrency count for each account
	concurrencyMap := make(map[int64]int64)
	for _, account := range candidates {
		count, err := s.concurrencyRepo.GetCount(ctx, account.ID)
		if err != nil {
			// Ignore error, default to 0
			count = 0
		}
		concurrencyMap[account.ID] = count
	}

	// Multi-level sorting
	sort.Slice(candidates, func(i, j int) bool {
		accountI := candidates[i]
		accountJ := candidates[j]

		// 1. Priority (descending - higher is better)
		if accountI.Priority != accountJ.Priority {
			return accountI.Priority > accountJ.Priority
		}

		// 2. Concurrency (ascending - lower is better)
		concurrencyI := concurrencyMap[accountI.ID]
		concurrencyJ := concurrencyMap[accountJ.ID]
		if concurrencyI != concurrencyJ {
			return concurrencyI < concurrencyJ
		}

		// 3. LastUsedAt (ascending - earlier is better)
		if accountI.LastUsedAt == nil && accountJ.LastUsedAt == nil {
			return false // Both never used, maintain original order
		}
		if accountI.LastUsedAt == nil {
			return true // i never used, prefer it
		}
		if accountJ.LastUsedAt == nil {
			return false // j never used, prefer it
		}
		return accountI.LastUsedAt.Before(*accountJ.LastUsedAt)
	})

	// Return the first account after sorting
	return candidates[0], nil
}

// RoundRobinStrategy implements simple round-robin selection.
type RoundRobinStrategy struct {
	currentIndex int
}

// NewRoundRobinStrategy creates a new round-robin selection strategy.
func NewRoundRobinStrategy() SelectionStrategy {
	return &RoundRobinStrategy{currentIndex: 0}
}

func (s *RoundRobinStrategy) Select(ctx context.Context, candidates []*model.CodexAccount, selectionCtx SelectionContext) (*model.CodexAccount, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available accounts")
	}

	selected := candidates[s.currentIndex%len(candidates)]
	s.currentIndex++

	return selected, nil
}
