// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package biz

import (
	"context"
)

// UserBiz 用户业务逻辑接口
type UserBiz interface {
	// GetUserByID 根据ID获取用户
	GetUserByID(ctx context.Context, id int64) (interface{}, error)
	// ListUsers 获取用户列表
	ListUsers(ctx context.Context, page, pageSize int) ([]interface{}, int64, error)
}

// userBiz 用户业务逻辑实现
type userBiz struct {
	// 在这里注入需要的依赖
	// store store.Store
}

// NewUserBiz 创建用户业务逻辑实例
func NewUserBiz() UserBiz {
	return &userBiz{}
}

// GetUserByID 根据ID获取用户
func (b *userBiz) GetUserByID(ctx context.Context, id int64) (interface{}, error) {
	// TODO: 实现业务逻辑
	return nil, nil
}

// ListUsers 获取用户列表
func (b *userBiz) ListUsers(ctx context.Context, page, pageSize int) ([]interface{}, int64, error) {
	// TODO: 实现业务逻辑
	return nil, 0, nil
}
