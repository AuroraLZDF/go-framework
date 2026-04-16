// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package service

import "context"

// SMSProvider 短信服务提供商接口
type SMSProvider interface {
	// SendCode 发送验证码
	SendCode(phone string, code string) error
	// SendMessage 发送普通短信
	SendMessage(phone string, content string, templateCode string) error
}

// FileInfo 文件信息
type FileInfo struct {
	Name      string
	Size      int64
	Extension string
	MD5       string
	URL       string
}

// StorageProvider 文件存储提供商接口
type StorageProvider interface {
	// UploadFile 上传文件
	UploadFile(file []byte, filename string) (*FileInfo, error)
	// DeleteFile 删除文件
	DeleteFile(fileID string) error
	// GetFileURL 获取文件访问URL
	GetFileURL(fileID string) (string, error)
}

// Message LLM 聊天消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage Token 使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ModelConfig LLM 模型配置
type ModelConfig struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID      string `json:"id"`
	OwnedBy string `json:"owned_by"`
}

// LLMProvider LLM 提供商接口
type LLMProvider interface {
	// ChatCompletion 普通对话
	ChatCompletion(ctx context.Context, messages []Message, config ModelConfig) (string, Usage, error)
	// StreamCompletion 流式对话
	StreamCompletion(ctx context.Context, messages []Message, config ModelConfig, callback func(chunk string, usage Usage, err error))
	// ValidateModel 验证模型配置
	ValidateModel(ctx context.Context, config ModelConfig) ([]ModelInfo, error)
}
