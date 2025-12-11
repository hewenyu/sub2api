package redis

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

const (
	acquireSemaphoreScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local now = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])
local req_id = ARGV[4]

redis.call('ZREMRANGEBYSCORE', key, '-inf', now)

local count = redis.call('ZCARD', key)

if count < limit then
    redis.call('ZADD', key, now + ttl, req_id)
    redis.call('EXPIRE', key, ttl * 2)
    return 1
else
    return 0
end
`

	releaseSemaphoreScript = `
local key = KEYS[1]
local req_id = ARGV[1]

return redis.call('ZREM', key, req_id)
`

	getSemaphoreCountScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])

redis.call('ZREMRANGEBYSCORE', key, '-inf', now)

return redis.call('ZCARD', key)
`
)

type LuaScriptManager struct {
	client *redis.Client
	shas   sync.Map
}

func NewLuaScriptManager(client *redis.Client) *LuaScriptManager {
	return &LuaScriptManager{
		client: client,
	}
}

func (m *LuaScriptManager) LoadScripts(ctx context.Context) error {
	scripts := map[string]string{
		"acquire": acquireSemaphoreScript,
		"release": releaseSemaphoreScript,
		"count":   getSemaphoreCountScript,
	}

	for name, script := range scripts {
		sha, err := m.client.ScriptLoad(ctx, script).Result()
		if err != nil {
			return fmt.Errorf("failed to load script %s: %w", name, err)
		}
		m.shas.Store(name, sha)
	}

	return nil
}

func (m *LuaScriptManager) EvalSHA(ctx context.Context, name string, keys []string, args ...interface{}) (interface{}, error) {
	sha, ok := m.shas.Load(name)
	if !ok {
		return nil, fmt.Errorf("script %s not loaded", name)
	}

	result, err := m.client.EvalSha(ctx, sha.(string), keys, args...).Result()
	if err != nil && strings.Contains(err.Error(), "NOSCRIPT") {
		if loadErr := m.LoadScripts(ctx); loadErr != nil {
			return nil, loadErr
		}
		sha, _ = m.shas.Load(name)
		result, err = m.client.EvalSha(ctx, sha.(string), keys, args...).Result()
	}

	return result, err
}
