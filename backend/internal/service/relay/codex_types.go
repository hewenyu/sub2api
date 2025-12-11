package relay

// CodexRequest represents a Codex API request.
type CodexRequest struct {
	Model          string    `json:"model" binding:"required"`
	Messages       []Message `json:"messages" binding:"required"`
	Temperature    *float64  `json:"temperature,omitempty"`
	MaxTokens      *int      `json:"max_tokens,omitempty"`
	Stream         bool      `json:"stream"`
	ConversationID *string   `json:"conversation_id,omitempty"`
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// CodexResponse represents a Codex API response.
type CodexResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
}

// Choice represents a response choice.
type Choice struct {
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"`
	Delta        *Message `json:"delta,omitempty"`
	FinishReason *string  `json:"finish_reason,omitempty"`
}

// Usage represents token usage information.
type Usage struct {
	// Standard OpenAI-style fields
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`

	// Responses API style fields
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`

	// Detailed token breakdown for prompt caching.
	InputTokensDetails   *UsageTokensDetails `json:"input_tokens_details,omitempty"`
	PromptTokensDetails  *UsageTokensDetails `json:"prompt_tokens_details,omitempty"`
	CacheCreation        *CacheCreationUsage `json:"cache_creation,omitempty"`
	CacheReadInputTokens int                 `json:"cache_read_input_tokens,omitempty"`
	// Some variants use cache_creation_input_tokens at the top level.
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
}

// UsageTokensDetails captures detailed token information such as cached tokens
// and cache creation tokens.
type UsageTokensDetails struct {
	CachedTokens             int `json:"cached_tokens,omitempty"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheCreationTokens      int `json:"cache_creation_tokens,omitempty"`
}

// CacheCreationUsage captures more detailed cache creation breakdown (e.g. 5m/1h).
type CacheCreationUsage struct {
	Ephemeral5mInputTokens int `json:"ephemeral_5m_input_tokens,omitempty"`
	Ephemeral1hInputTokens int `json:"ephemeral_1h_input_tokens,omitempty"`
}
