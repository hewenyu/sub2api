package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
)

// HTTPClientConfig defines configuration for HTTP client creation.
type HTTPClientConfig struct {
	Timeout         time.Duration
	KeepAlive       time.Duration
	MaxIdleConns    int
	MaxConnsPerHost int
}

// DefaultHTTPClientConfig returns default configuration for standard requests.
func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		Timeout:         30 * time.Second,
		KeepAlive:       30 * time.Second,
		MaxIdleConns:    100,
		MaxConnsPerHost: 10,
	}
}

// StreamingHTTPClientConfig returns configuration optimized for streaming requests.
func StreamingHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		Timeout:         300 * time.Second,
		KeepAlive:       90 * time.Second,
		MaxIdleConns:    100,
		MaxConnsPerHost: 10,
	}
}

// NewHTTPClient creates an HTTP client with optional proxy configuration.
// Returns a default client if proxy is nil.
func NewHTTPClient(proxyConfig *model.ProxyConfig, config HTTPClientConfig, encryptionKey string) (*http.Client, error) {
	if proxyConfig == nil {
		return &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        config.MaxIdleConns,
				MaxConnsPerHost:     config.MaxConnsPerHost,
				IdleConnTimeout:     config.KeepAlive,
				DisableKeepAlives:   false,
				DisableCompression:  false,
				ForceAttemptHTTP2:   true,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		}, nil
	}

	proxyURL, err := buildProxyURL(proxyConfig, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build proxy URL: %w", err)
	}

	var transport http.RoundTripper

	if proxyConfig.Protocol == "socks5" {
		// Extract authentication credentials from proxy URL
		var auth *proxy.Auth
		if proxyURL.User != nil {
			password, _ := proxyURL.User.Password()
			auth = &proxy.Auth{
				User:     proxyURL.User.Username(),
				Password: password,
			}
		}
		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
		}
		transport = &http.Transport{
			Dial:                dialer.Dial,
			MaxIdleConns:        config.MaxIdleConns,
			MaxConnsPerHost:     config.MaxConnsPerHost,
			IdleConnTimeout:     config.KeepAlive,
			DisableKeepAlives:   false,
			DisableCompression:  false,
			ForceAttemptHTTP2:   true,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	} else {
		transport = &http.Transport{
			Proxy:               http.ProxyURL(proxyURL),
			MaxIdleConns:        config.MaxIdleConns,
			MaxConnsPerHost:     config.MaxConnsPerHost,
			IdleConnTimeout:     config.KeepAlive,
			DisableKeepAlives:   false,
			DisableCompression:  false,
			ForceAttemptHTTP2:   true,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}, nil
}

// buildProxyURL constructs a proxy URL from a ProxyConfig.
// Format: protocol://[username:password@]host:port
func buildProxyURL(proxyConfig *model.ProxyConfig, encryptionKey string) (*url.URL, error) {
	var password string
	if proxyConfig.Password != nil && *proxyConfig.Password != "" {
		decrypted, err := crypto.AES256Decrypt(*proxyConfig.Password, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password: %w", err)
		}
		password = decrypted
	}

	var userInfo *url.Userinfo
	if proxyConfig.Username != nil && *proxyConfig.Username != "" {
		if password != "" {
			userInfo = url.UserPassword(*proxyConfig.Username, password)
		} else {
			userInfo = url.User(*proxyConfig.Username)
		}
	}

	return &url.URL{
		Scheme: proxyConfig.Protocol,
		Host:   fmt.Sprintf("%s:%d", proxyConfig.Host, proxyConfig.Port),
		User:   userInfo,
	}, nil
}
