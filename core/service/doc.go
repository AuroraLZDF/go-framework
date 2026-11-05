// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

// Package service 声明框架预留的外部服务扩展点接口（短信、对象存储、LLM）.
//
// 该包当前为实验性（Experimental）API：框架本身没有引用它，也不附带实现，
// 接口签名可能随实际需求在 1.x 的 minor 版本中调整或删除，不承诺源码兼容。
// 如果需要在业务中集成这些能力，建议先在自己的包内定义接口，待本包稳定后再迁移。
package service
