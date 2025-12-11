package proxy

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
)

var (
	// proxyNameRegex validates proxy name: 1-100 chars, alphanumeric, underscore, hyphen
	// Must not start with hyphen or underscore
	proxyNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,99}$`)

	// Valid proxy protocols
	validProtocols = map[string]bool{
		"http":   true,
		"https":  true,
		"socks5": true,
	}
)

// ValidateProxyName validates proxy configuration name.
// Rules: 1-100 characters, alphanumeric/underscore/hyphen, cannot start with - or _
func ValidateProxyName(name string) error {
	if name == "" {
		return errors.New("proxy name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("proxy name too long: max 100 characters, got %d", len(name))
	}

	if !proxyNameRegex.MatchString(name) {
		return errors.New("proxy name must contain only alphanumeric characters, underscores, and hyphens, and cannot start with - or _")
	}

	return nil
}

// ValidateProtocol validates proxy protocol.
// Valid protocols: http, https, socks5
func ValidateProtocol(protocol string) error {
	if protocol == "" {
		return errors.New("proxy protocol cannot be empty")
	}

	protocol = strings.ToLower(protocol)
	if !validProtocols[protocol] {
		return fmt.Errorf("invalid proxy protocol: %s (must be http, https, or socks5)", protocol)
	}

	return nil
}

// ValidateHost validates proxy host (IP address, domain name, or localhost).
func ValidateHost(host string) error {
	if host == "" {
		return errors.New("proxy host cannot be empty")
	}

	if len(host) > 255 {
		return fmt.Errorf("proxy host too long: max 255 characters, got %d", len(host))
	}

	// Check if it's a valid IP address
	if ip := net.ParseIP(host); ip != nil {
		return nil
	}

	// Check if it's "localhost"
	if strings.ToLower(host) == "localhost" {
		return nil
	}

	// Check if it's a valid domain name (basic validation)
	// Domain can contain alphanumeric, hyphens, and dots
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	if !domainRegex.MatchString(host) {
		return fmt.Errorf("invalid proxy host: must be a valid IP address, domain name, or localhost")
	}

	return nil
}

// ValidatePort validates proxy port number.
// Valid range: 1-65535
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid proxy port: %d (must be between 1 and 65535)", port)
	}
	return nil
}

// ValidateCreateRequest validates a CreateProxyRequest.
func ValidateCreateRequest(req *CreateProxyRequest) error {
	if req == nil {
		return errors.New("create proxy request cannot be nil")
	}

	if err := ValidateProxyName(req.Name); err != nil {
		return fmt.Errorf("invalid name: %w", err)
	}

	if err := ValidateProtocol(req.Protocol); err != nil {
		return fmt.Errorf("invalid protocol: %w", err)
	}

	if err := ValidateHost(req.Host); err != nil {
		return fmt.Errorf("invalid host: %w", err)
	}

	if err := ValidatePort(req.Port); err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	// Username can be nil (optional), but if provided must not be empty
	if req.Username != nil && *req.Username == "" {
		return errors.New("username cannot be empty string (use nil for no authentication)")
	}

	// If username is provided, check length
	if req.Username != nil && len(*req.Username) > 255 {
		return fmt.Errorf("username too long: max 255 characters, got %d", len(*req.Username))
	}

	return nil
}

// ValidateUpdateRequest validates an UpdateProxyRequest.
func ValidateUpdateRequest(req *UpdateProxyRequest) error {
	if req == nil {
		return errors.New("update proxy request cannot be nil")
	}

	// At least one field must be provided
	if req.Name == nil && req.Enabled == nil && req.Protocol == nil &&
		req.Host == nil && req.Port == nil && req.Username == nil && req.Password == nil {
		return errors.New("at least one field must be provided for update")
	}

	// Validate each field if provided
	if req.Name != nil {
		if err := ValidateProxyName(*req.Name); err != nil {
			return fmt.Errorf("invalid name: %w", err)
		}
	}

	if req.Protocol != nil {
		if err := ValidateProtocol(*req.Protocol); err != nil {
			return fmt.Errorf("invalid protocol: %w", err)
		}
	}

	if req.Host != nil {
		if err := ValidateHost(*req.Host); err != nil {
			return fmt.Errorf("invalid host: %w", err)
		}
	}

	if req.Port != nil {
		if err := ValidatePort(*req.Port); err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
	}

	// Username validation: can be nil (no change), empty string (remove), or value
	if req.Username != nil && len(*req.Username) > 255 {
		return fmt.Errorf("username too long: max 255 characters, got %d", len(*req.Username))
	}

	return nil
}
