package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/proxy"
)

type ProxyHandler struct {
	proxyService proxy.ProxyConfigService
	logger       *zap.Logger
}

func NewProxyHandler(proxyService proxy.ProxyConfigService, logger *zap.Logger) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxyService,
		logger:       logger,
	}
}

type ProxyResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Enabled     bool    `json:"enabled"`
	Protocol    string  `json:"protocol"`
	Host        string  `json:"host"`
	Port        int     `json:"port"`
	Username    *string `json:"username,omitempty"`
	HasPassword bool    `json:"has_password"`
	IsDefault   bool    `json:"is_default"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`
}

func (h *ProxyHandler) toProxyResponse(p *model.ProxyConfig) ProxyResponse {
	return ProxyResponse{
		ID:          p.ID,
		Name:        p.Name,
		Enabled:     p.Enabled,
		Protocol:    p.Protocol,
		Host:        p.Host,
		Port:        p.Port,
		Username:    p.Username,
		HasPassword: p.Password != nil && *p.Password != "",
		IsDefault:   p.IsDefault,
		CreatedAt:   p.CreatedAt.Unix(),
		UpdatedAt:   p.UpdatedAt.Unix(),
	}
}

func (h *ProxyHandler) ListProxies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := repository.ProxyConfigFilters{}
	if enabled := c.Query("enabled"); enabled != "" {
		if e, err := strconv.ParseBool(enabled); err == nil {
			filters.Enabled = &e
		}
	}
	if isDefault := c.Query("is_default"); isDefault != "" {
		if d, err := strconv.ParseBool(isDefault); err == nil {
			filters.IsDefault = &d
		}
	}
	if protocol := c.Query("protocol"); protocol != "" {
		filters.Protocol = &protocol
	}

	proxies, total, err := h.proxyService.ListProxies(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list proxies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list proxies"})
		return
	}

	items := make([]ProxyResponse, len(proxies))
	for i, p := range proxies {
		items[i] = h.toProxyResponse(p)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": items,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      total,
				"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		},
	})
}

func (h *ProxyHandler) CreateProxy(c *gin.Context) {
	var req proxy.CreateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create proxy request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	proxyConfig, err := h.proxyService.CreateProxy(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create proxy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.toProxyResponse(proxyConfig),
	})
}

func (h *ProxyHandler) GetProxy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	proxyConfig, err := h.proxyService.GetProxy(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get proxy", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.toProxyResponse(proxyConfig),
	})
}

func (h *ProxyHandler) UpdateProxy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	var req proxy.UpdateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update proxy request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.proxyService.UpdateProxy(c.Request.Context(), id, &req); err != nil {
		h.logger.Error("Failed to update proxy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Proxy updated successfully",
	})
}

func (h *ProxyHandler) DeleteProxy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	if err := h.proxyService.DeleteProxy(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete proxy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete proxy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Proxy deleted successfully",
	})
}

func (h *ProxyHandler) SetDefaultProxy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	if err := h.proxyService.SetDefaultProxy(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to set default proxy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Default proxy set successfully",
	})
}

func (h *ProxyHandler) TestProxy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	result, err := h.proxyService.TestProxyWithGeolocation(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to test proxy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

type ProxyNameInfo struct {
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	Protocol  string `json:"protocol"`
	IsDefault bool   `json:"is_default"`
}

func (h *ProxyHandler) GetProxyNames(c *gin.Context) {
	enabled := true
	filters := repository.ProxyConfigFilters{
		Enabled: &enabled,
	}

	proxies, _, err := h.proxyService.ListProxies(c.Request.Context(), filters, 1, 1000)
	if err != nil {
		h.logger.Error("Failed to list proxy names", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list proxy names"})
		return
	}

	names := make([]ProxyNameInfo, len(proxies))
	for i, p := range proxies {
		names[i] = ProxyNameInfo{
			Name:      p.Name,
			Enabled:   p.Enabled,
			Protocol:  p.Protocol,
			IsDefault: p.IsDefault,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    names,
	})
}
