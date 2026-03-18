package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AntiCheatMiddleware struct {
	redis *redis.Client
}

func NewAntiCheatMiddleware(redis *redis.Client) *AntiCheatMiddleware {
	return &AntiCheatMiddleware{redis: redis}
}

func (m *AntiCheatMiddleware) RateLimitSwipe() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.redis == nil {
			c.Next()
			return
		}

		userIDStr, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userID, ok := userIDStr.(uuid.UUID)
		if !ok {
			c.Next()
			return
		}

		// Sliding window or fixed window (fixed is easier for MVP)
		key := fmt.Sprintf("swipe_limit:%s:%d", userID.String(), time.Now().Unix()/60) // 1 minute window
		
		ctx := context.Background()
		count, err := m.redis.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			m.redis.Expire(ctx, key, 2*time.Minute)
		}

		// Limit to 60 swipes per minute (configurable)
		if count > 60 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "swipe spam detected, please slow down",
			})
			return
		}

		c.Next()
	}
}
