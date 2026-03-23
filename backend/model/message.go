package model

import (
	"time"
)

// Message 消息模型
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SenderID  uint      `json:"sender_id"` // 发送者ID，系统消息为0
	ReceiverID uint     `json:"receiver_id"` // 接收者ID
	Type      string    `json:"type"` // 消息类型：comment, like, reply, review, role
	Title     string    `json:"title"` // 消息标题
	Content   string    `json:"content"` // 消息内容
	RelatedID string    `json:"related_id"` // 相关资源ID，如文章ID、评论ID
	RelatedType string  `json:"related_type"` // 相关资源类型：post, comment
	Status    string    `json:"status"` // 消息状态：sent, read
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}

// Notification 通知模型
type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"` // 接收用户ID
	Type      string    `json:"type"` // 通知类型：comment, like, reply, review, role
	Title     string    `json:"title"` // 通知标题
	Content   string    `json:"content"` // 通知内容
	RelatedID string    `json:"related_id"` // 相关资源ID
	RelatedType string  `json:"related_type"` // 相关资源类型
	IsRead    bool      `json:"is_read"` // 是否已读
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Notification) TableName() string {
	return "notifications"
}