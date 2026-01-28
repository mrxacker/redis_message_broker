// chat-service/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mrxacker/redis_message_broker/shared"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Chat Service: Connected to Redis")
}

type SendMessageRequest struct {
	UserID     string `json:"userId" binding:"required"`
	SenderID   string `json:"senderId" binding:"required"`
	SenderName string `json:"senderName" binding:"required"`
	Message    string `json:"message" binding:"required"`
	ChatID     string `json:"chatId" binding:"required"`
}

func publishNotification(ctx context.Context, notification *shared.Notification) error {
	notification.Timestamp = time.Now()

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: "notifications:stream",
		MaxLen: 10000,
		Approx: true,
		ID:     "*",
		Values: map[string]interface{}{
			"data": string(data),
		},
	}

	msgID, err := redisClient.XAdd(ctx, args).Result()
	if err != nil {
		return fmt.Errorf("failed to publish to stream: %w", err)
	}

	log.Printf("✅ Published notification to Redis (ID: %s) for user %s", msgID, notification.UserID)
	return nil
}

func sendMessageHandler(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("📨 Received message from %s to %s: %s", req.SenderName, req.UserID, req.Message)

	// Create notification
	notification := &shared.Notification{
		Type:       "NEW_MESSAGE",
		UserID:     req.UserID,
		SenderID:   req.SenderID,
		SenderName: req.SenderName,
		Message:    req.Message,
		ChatID:     req.ChatID,
		Metadata: map[string]interface{}{
			"priority": "high",
		},
	}

	// Publish to Redis Stream
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := publishNotification(ctx, notification); err != nil {
		log.Printf("❌ Failed to publish notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message sent and notification published",
	})
}

func main() {
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "chat"})
	})

	// Send message endpoint
	router.POST("/api/messages", sendMessageHandler)

	log.Println("🚀 Chat Service started on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start chat service: %v", err)
	}
}
