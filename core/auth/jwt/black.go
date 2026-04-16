// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package token

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"go-framework/core/util"
)

// Blacklist 黑名单中存储的数据结构
type Blacklist struct {
	ID        string `json:"id"`
	ExpiredAt int64  `json:"expired_at"`
}

// IsBlacklisted 检查 Token 是否在黑名单中
func IsBlacklisted(token *jwt.Token) bool {
	// 获取 Token 的 ID
	tokenID := token.Claims.(jwt.MapClaims)[config.TokenId].(string)
	expire, _ := token.Claims.GetExpirationTime()

	// 读取黑名单文件中的内容
	blacklist, err := readBlacklistFromFile()
	if err != nil {
		return false
	}

	// 遍历黑名单列表，检查 Token 是否在其中
	for _, blToken := range blacklist {
		if blToken.ID == tokenID {
			// token 过期时间 小于等于黑名单中时间
			if expire.Unix() <= blToken.ExpiredAt {
				return true
			}
		}
	}

	return false
}

// AddToBlacklist 添加 Token 到黑名单中
func AddToBlacklist(r *http.Request) error {
	token, err := ParseToken(r)
	if err != nil {
		return err
	}

	// 获取 Token 的 ID
	tokenID := token.Claims.(jwt.MapClaims)[config.TokenId].(string)
	// 获取 Token 的过期时间
	expirationTime := int64(token.Claims.(jwt.MapClaims)["exp"].(float64))

	// 创建 Blacklist 结构体实例
	_blacklist := Blacklist{ID: tokenID, ExpiredAt: expirationTime}

	// 读取黑名单文件中的内容
	blacklist, err := readBlacklistFromFile()
	if err != nil {
		return err
	}

	// 将 Token 添加到黑名单列表
	blacklist = append(blacklist, _blacklist)

	// 将黑名单转换为 JSON 格式
	data, err := json.MarshalIndent(blacklist, "", "  ")
	if err != nil {
		return err
	}

	// 将输入JSON字符串压缩为一个单行的JSON字符串，并写入到输出缓冲区中
	var out bytes.Buffer
	err = json.Compact(&out, data)
	if err != nil {
		return err
	}

	// 将黑名单信息写入文件
	err = writeBlacklistToFile(out.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// FailureAllClient 失效所有客户端的 token
func FailureAllClient(tokenId string) error {
	// 获取黑名单列表
	var path = config.BlackPath
	if err := util.AddFileIfNotExists(path); err != nil {
		return err
	}

	// 读取文件内容
	blacklistBytes, err := os.ReadFile(config.BlackPath)
	if err != nil {
		return err
	}

	// 将黑名单数据解析为切片
	var blacklist []Blacklist
	if len(blacklistBytes) > 0 {
		if err = json.Unmarshal(blacklistBytes, &blacklist); err != nil {
			return err
		}
	}

	// 将该用户之前的所有 token 从黑名单中移除
	for i := 0; i < len(blacklist); i++ {
		if blacklist[i].ID == tokenId {
			blacklist = append(blacklist[:i], blacklist[i+1:]...)
			i--
		}
	}

	// 将新的 token 加入黑名单
	blacklist = append(blacklist, Blacklist{
		ID:        tokenId,
		ExpiredAt: time.Now().Add(time.Duration(config.Expire) * time.Hour).Unix(),
	})

	// 将黑名单数据写入文件
	blacklistBytes, err = json.MarshalIndent(blacklist, "", "  ")
	if err != nil {
		return err
	}

	// 将输入JSON字符串压缩为一个单行的JSON字符串，并写入到输出缓冲区中
	var out bytes.Buffer
	err = json.Compact(&out, blacklistBytes)
	if err != nil {
		return err
	}

	// 将黑名单信息写入文件
	err = writeBlacklistToFile(out.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// readBlacklistFromFile 从文件中读取黑名单信息
func readBlacklistFromFile() ([]Blacklist, error) {
	var path = config.BlackPath
	if err := util.AddFileIfNotExists(path); err != nil {
		return nil, err
	}

	// 读取文件内容
	blacklistBytes, err := os.ReadFile(config.BlackPath)
	if err != nil {
		return nil, err
	}

	if len(blacklistBytes) <= 0 {
		return []Blacklist{}, nil
	}

	// 解析文件内容为 Blacklist 列表
	var blacklist []Blacklist
	err = json.Unmarshal(blacklistBytes, &blacklist)
	if err != nil {
		return nil, err
	}

	return blacklist, nil
}

// writeBlacklistToFile 将黑名单信息写入文件
func writeBlacklistToFile(bytes []byte) error {
	// 写入文件
	if err := os.WriteFile(config.BlackPath, bytes, fs.FileMode(0644)); err != nil {
		return err
	}

	return nil
}
