// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/gin-gonic/gin"

	"go-framework/core/response"
)

// UserController 用户控制器
type UserController struct {
	// userBiz biz.UserBiz
}

// NewUserController 创建用户控制器实例
func NewUserController() *UserController {
	return &UserController{}
}

// GetUserByID 根据ID获取用户
// @Summary 获取用户信息
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id} [get]
func (ctrl *UserController) GetUserByID(c *gin.Context) {
	// TODO: 实现获取用户逻辑
	response.Success(c, gin.H{
		"id":   1,
		"name": "Alice",
	}, "ok")
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/users [get]
func (ctrl *UserController) ListUsers(c *gin.Context) {
	// TODO: 实现获取用户列表逻辑
	response.Success(c, []gin.H{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	}, "ok")
}
