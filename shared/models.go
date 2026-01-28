package shared

import "time"

type Notification struct {
	Type       string                 `json:"type"`
	UserID     string                 `json:"userId"`
	SenderID   string                 `json:"senderId"`
	SenderName string                 `json:"senderName"`
	Message    string                 `json:"message"`
	ChatID     string                 `json:"chatId"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
}
