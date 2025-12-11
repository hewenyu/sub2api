package proxy

// CreateProxyRequest represents the request to create a new proxy configuration.
type CreateProxyRequest struct {
	Name     string  `json:"name" validate:"required,min=1,max=100"`
	Enabled  bool    `json:"enabled"`
	Protocol string  `json:"protocol" validate:"required,oneof=http https socks5"`
	Host     string  `json:"host" validate:"required,max=255"`
	Port     int     `json:"port" validate:"required,min=1,max=65535"`
	Username *string `json:"username,omitempty" validate:"omitempty,max=255"`
	Password *string `json:"password,omitempty"`
}

// UpdateProxyRequest represents the request to update an existing proxy configuration.
// All fields are optional (pointers). Only provided fields will be updated.
// Password field: nil = no change, empty string = remove password, value = update password.
type UpdateProxyRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Enabled  *bool   `json:"enabled,omitempty"`
	Protocol *string `json:"protocol,omitempty" validate:"omitempty,oneof=http https socks5"`
	Host     *string `json:"host,omitempty" validate:"omitempty,max=255"`
	Port     *int    `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	Username *string `json:"username,omitempty" validate:"omitempty,max=255"`
	Password *string `json:"password,omitempty"`
}

// ProxyTestResult represents the result of testing a proxy connection.
type ProxyTestResult struct {
	Success     bool    `json:"success"`
	Message     string  `json:"message"`
	IP          *string `json:"ip,omitempty"`
	Country     *string `json:"country,omitempty"`
	Region      *string `json:"region,omitempty"`
	City        *string `json:"city,omitempty"`
	ISP         *string `json:"isp,omitempty"`
	ResponseMS  *int64  `json:"response_ms,omitempty"`
	Error       *string `json:"error,omitempty"`
	GeoProvider *string `json:"geo_provider,omitempty"`
}
