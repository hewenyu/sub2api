package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/middleware"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/admin"
)

type LoginHandler struct {
	adminService admin.AdminService
	adminRepo    repository.AdminRepository
	logger       *zap.Logger
}

func NewLoginHandler(adminService admin.AdminService, adminRepo repository.AdminRepository, logger *zap.Logger) *LoginHandler {
	return &LoginHandler{
		adminService: adminService,
		adminRepo:    adminRepo,
		logger:       logger,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string    `json:"token"`
	Admin AdminInfo `json:"admin"`
}

type AdminInfo struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	IsActive  bool   `json:"is_active"`
	CreatedAt int64  `json:"created_at"`
}

func (h *LoginHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Authenticate
	token, admin, err := h.adminService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Build response
	resp := LoginResponse{
		Token: token,
		Admin: AdminInfo{
			ID:        admin.ID,
			Username:  admin.Username,
			Email:     admin.Email,
			IsActive:  admin.IsActive,
			CreatedAt: admin.CreatedAt.Unix(),
		},
	}

	c.JSON(http.StatusOK, resp)
}

func (h *LoginHandler) GetInfo(c *gin.Context) {
	adminID := middleware.GetAdminID(c)
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
		})
		return
	}

	// Get admin info from database
	admin, err := h.adminRepo.GetByID(c.Request.Context(), adminID)
	if err != nil {
		h.logger.Error("Failed to get admin info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get admin info",
		})
		return
	}

	c.JSON(http.StatusOK, AdminInfo{
		ID:        admin.ID,
		Username:  admin.Username,
		Email:     admin.Email,
		IsActive:  admin.IsActive,
		CreatedAt: admin.CreatedAt.Unix(),
	})
}
