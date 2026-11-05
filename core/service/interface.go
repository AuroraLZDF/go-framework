// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package service

import "context"

// SMSProvider 短信服务提供商接口.
// 远程调用应遵循 ctx 的超时/取消；实现方需在 ctx 结束时尽快返回.
type SMSProvider interface {
	// SendCode 发送验证码
	SendCode(ctx context.Context, phone string, code string) error
	// SendMessage 发送普通短信
	SendMessage(ctx context.Context, phone string, content string, templateCode string) error
}

// FileInfo 文件信息
type FileInfo struct {
	Name      string
	Size      int64
	Extension string
	MD5       string
	URL       string
}

// StorageProvider 文件存储提供商接口.
// 上传/下载/删除均可能访问远端存储，因此统一接收 ctx.
type StorageProvider interface {
	// UploadFile 上传文件
	UploadFile(ctx context.Context, file []byte, filename string) (*FileInfo, error)
	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, fileID string) error
	// GetFileURL 获取文件访问URL
	GetFileURL(ctx context.Context, fileID string) (string, error)
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
