package scheduler

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository/redis"
)

type HealthAwareStrategy struct {
	concurrencyRepo redis.ConcurrencyRepository
}

func NewHealthAwareStrategy(concurrencyRepo redis.ConcurrencyRepository) SelectionStrategy {
	return &HealthAwareStrategy{
		concurrencyRepo: concurrencyRepo,
	}
}

func (s *HealthAwareStrategy) Select(ctx context.Context, candidates []*model.CodexAccount, selectionCtx SelectionContext) (*model.CodexAccount, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available accounts")
	}

	type accountScore struct {
		account *model.CodexAccount
		score   float64
	}

	scores := make([]accountScore, 0, len(candidates))

	for _, account := range candidates {
		score := s.calculateHealthScore(ctx, account)
		scores = append(scores, accountScore{account: account, score: score})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	return scores[0].account, nil
}

func (s *HealthAwareStrategy) calculateHealthScore(ctx context.Context, account *model.CodexAccount) float64 {
	score := float64(account.Priority)

	if account.RateLimitedUntil != nil && account.RateLimitedUntil.After(time.Now()) {
		score -= 1000
	}

	concurrency, err := s.concurrencyRepo.GetCount(ctx, account.ID)
	if err == nil {
		score -= float64(concurrency) * 10
	}

	return score
}
