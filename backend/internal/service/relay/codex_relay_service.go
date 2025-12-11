package relay

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/account"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/billing"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/circuitbreaker"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/proxy"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/scheduler"
)

// CodexRelayService handles Codex API requests and relays them to upstream APIs.
type CodexRelayService interface {
	// HandleRequest processes a Codex-style request.
	// reqBody is the normalized internal request (CodexRequest).
	// rawBody is the original JSON body from the client (Responses API shape),
	// used when forwarding to upstream endpoints that expect the original schema
	// such as the ChatGPT Codex Responses API.
	HandleRequest(c *gin.Context, apiKey *model.APIKey, reqBody *CodexRequest, rawBody []byte, requestPath string) error
}

type codexRelayService struct {
	schedulerSvc   scheduler.SchedulerService
	accountSvc     account.CodexAccountService
	usageCollector billing.UsageCollector
	clientManager  proxy.ProxyClientManager
	cbManager      *circuitbreaker.Manager
	logger         *zap.Logger
	logPayloads    bool
}

// oauthRefreshThreshold defines how close to the access token expiry
// we consider it reasonable to perform a proactive refresh.
// Codex OAuth tokens typically have a multi-day lifetime, so refreshing
// only when there is less than 24 hours remaining avoids unnecessary
// refresh attempts while still keeping tokens fresh.
const oauthRefreshThreshold = 24 * time.Hour

// NewCodexRelayService creates a new Codex relay service.
func NewCodexRelayService(
	schedulerSvc scheduler.SchedulerService,
	accountSvc account.CodexAccountService,
	usageCollector billing.UsageCollector,
	clientManager proxy.ProxyClientManager,
	logger *zap.Logger,
	logPayloads bool,
) CodexRelayService {
	return &codexRelayService{
		schedulerSvc:   schedulerSvc,
		accountSvc:     accountSvc,
		usageCollector: usageCollector,
		clientManager:  clientManager,
		cbManager:      circuitbreaker.NewManager(),
		logger:         logger,
		logPayloads:    logPayloads,
	}
}
