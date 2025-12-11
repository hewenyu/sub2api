package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Wei-Shaw/sub2api/backend/internal/shutdown"
)

func RequestTrackerMiddleware(tracker *shutdown.RequestTracker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tracker.IsShuttingDown() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Server is shutting down",
			})
			c.Abort()
			return
		}

		tracker.Start()
		defer tracker.End()

		c.Next()
	}
}
