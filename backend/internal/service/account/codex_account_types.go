package account

// CreateCodexAccountRequest represents a request to create a Codex account.
type CreateCodexAccountRequest struct {
	Name            string  `json:"name" binding:"required"`
	AccountType     string  `json:"account_type" binding:"required,oneof=openai-oauth openai-responses"`
	Email           *string `json:"email"`
	APIKey          *string `json:"api_key"` // Only for openai-responses type
	BaseAPI         string  `json:"base_api"`
	CustomUserAgent *string `json:"custom_user_agent"`
	DailyQuota      float64 `json:"daily_quota"`
	QuotaResetTime  string  `json:"quota_reset_time"`
	Priority        int     `json:"priority"`
	Schedulable     bool    `json:"schedulable"`
	ProxyName       *string `json:"proxy_name"`
}
