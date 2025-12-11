package relay

// ResponsesRequest represents an OpenAI Responses API request.
// Based on the official OpenAI Responses API specification.
type ResponsesRequest struct {
	Model              string                 `json:"model" binding:"required"`
	Input              interface{}            `json:"input" binding:"required"` // Can be string or array of messages
	PreviousResponseID *string                `json:"previous_response_id,omitempty"`
	Tools              []Tool                 `json:"tools,omitempty"`
	Instructions       *string                `json:"instructions,omitempty"`
	Temperature        *float64               `json:"temperature,omitempty"`
	TopP               *float64               `json:"top_p,omitempty"`
	MaxOutputTokens    *int                   `json:"max_output_tokens,omitempty"`
	ToolChoice         interface{}            `json:"tool_choice,omitempty"` // Can be string or object
	ParallelToolCalls  *bool                  `json:"parallel_tool_calls,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	Stream             bool                   `json:"stream,omitempty"`
	Background         *bool                  `json:"background,omitempty"`
	Reasoning          *ReasoningConfig       `json:"reasoning,omitempty"`
	Include            []string               `json:"include,omitempty"`
	Store              *bool                  `json:"store,omitempty"`
}

// Tool represents a tool definition in the Responses API.
type Tool struct {
	Type     string      `json:"type" binding:"required"`
	Function *Function   `json:"function,omitempty"`
	MCP      *MCPConfig  `json:"mcp,omitempty"`
	Settings interface{} `json:"settings,omitempty"`
}

// Function represents a function tool definition.
type Function struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// MCPConfig represents MCP tool configuration.
type MCPConfig struct {
	ServerName string `json:"server_name" binding:"required"`
	ToolName   string `json:"tool_name" binding:"required"`
}

// ReasoningConfig represents reasoning effort configuration.
type ReasoningConfig struct {
	Effort string `json:"effort,omitempty"` // low, medium, high
}

// ResponsesResponse represents the Responses API response structure.
type ResponsesResponse struct {
	ID       string                 `json:"id"`
	Object   string                 `json:"object"`
	Created  int64                  `json:"created"`
	Model    string                 `json:"model"`
	Output   interface{}            `json:"output,omitempty"`  // Can be string or array
	Choices  []ResponseChoice       `json:"choices,omitempty"` // For compatibility
	Usage    *Usage                 `json:"usage,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Status   *string                `json:"status,omitempty"` // For background tasks
	Error    *ResponseError         `json:"error,omitempty"`
}

// ResponseChoice represents a choice in the response.
type ResponseChoice struct {
	Index        int         `json:"index"`
	Message      *Message    `json:"message,omitempty"`
	Delta        *Message    `json:"delta,omitempty"`
	FinishReason *string     `json:"finish_reason,omitempty"`
	Output       interface{} `json:"output,omitempty"`
}

// ResponseError represents an error in the response.
type ResponseError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}
