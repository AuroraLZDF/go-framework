// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

package util

import (
	"net"
	"net/http"
	"strings"
)

// GetRealClientIP 获取真实的客户端IP地址
func GetRealClientIP(r *http.Request) string {
	// 优先级从高到低
	headers := []string{
		"X-Real-IP",
		"X-Forwarded-For",
		"CF-Connecting-IP", // Cloudflare
		"Fastly-Client-IP", // Fastly
	}

	for _, header := range headers {
		if ip := r.Header.Get(header); ip != "" {
			return parseIP(ip)
		}
	}

	// 最后回退到标准方法
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// 处理可能包含多个IP的情况（如X-Forwarded-For: ip1, ip2, ip3）
func parseIP(ipStr string) string {
	ips := strings.Split(ipStr, ",")
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if isValidIP(ip) {
			return ip
		}
	}
	return ""
}

// 验证IP有效性（排除内网IP）
func isValidIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}

	// 过滤私有IP段
	if parsed.IsPrivate() || parsed.IsLoopback() {
		return false
	}

	return true
}
