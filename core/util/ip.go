// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package util

import (
	"net"
	"net/http"
	"strings"
)

// GetRealClientIP 获取真实的客户端IP地址.
// 依次检查常见代理头，跳过无效/内网地址，全部无效时回退到 RemoteAddr.
func GetRealClientIP(r *http.Request) string {
	// 优先级从高到低
	headers := []string{
		"X-Real-IP",
		"X-Forwarded-For",
		"CF-Connecting-IP", // Cloudflare
		"Fastly-Client-IP", // Fastly
	}

	for _, header := range headers {
		if ip := parseIP(r.Header.Get(header)); ip != "" {
			return ip
		}
	}

	// 回退到连接的远端地址（在开发环境/未经过代理时通常就是客户端地址）
	return remoteIP(r.RemoteAddr)
}

// remoteIP 从 RemoteAddr（host:port）中提取 IP，不含端口时原样返回.
func remoteIP(remoteAddr string) string {
	remoteAddr = strings.TrimSpace(remoteAddr)
	if remoteAddr == "" {
		return ""
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}

	return strings.TrimSpace(host)
}

// parseIP 处理可能包含多个IP的情况（如X-Forwarded-For: ip1, ip2, ip3），
// 返回第一个公网有效IP；没有则返回 "".
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

// isValidIP 验证IP有效性（排除内网IP）
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
