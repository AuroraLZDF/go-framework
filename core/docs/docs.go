// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

// Package docs 提供 OpenAPI/Swagger 在线文档能力：
// 把 OpenAPI 规范文件（YAML/JSON）挂到 HTTP 服务上，并托管 Swagger UI。
//
// 使用示例（在 RegisterRoutes 中调用）：
//
//	if err := docs.Register(engine, docs.Options{SpecPath: "docs/openapi.yaml"}); err != nil {
//	    return err
//	}
//
// 启用后可访问：
//   - /docs            Swagger UI 页面
//   - /docs/openapi    OpenAPI 规范文件
//
// 规范文件可手写，也可由注解工具生成（如 swag CLI：swag init 产出 swagger.json，
// 将 SpecPath 指向该文件即可）。
package docs

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// Options 文档注册选项.
type Options struct {
	// SpecPath OpenAPI 规范文件路径（.yaml/.yml/.json），必填
	SpecPath string
	// RoutePrefix 文档路由前缀，默认 "/docs"
	RoutePrefix string
}

// prefix 返回规范化后的路由前缀.
func (o *Options) prefix() string {
	if o.RoutePrefix == "" {
		return "/docs"
	}

	return o.RoutePrefix
}

// Register 在 engine 上注册 Swagger UI 与 OpenAPI 规范路由.
//   - GET <prefix>          Swagger UI 页面
//   - GET <prefix>/openapi  OpenAPI 规范文件（按文件扩展名返回对应 Content-Type）
//
// SpecPath 文件不存在时返回错误，保证启动即失败（fail fast），
// 避免线上部署后才发现文档缺失。
func Register(engine *gin.Engine, opts Options) error {
	if opts.SpecPath == "" {
		return fmt.Errorf("docs: spec-path is required")
	}

	specBytes, err := os.ReadFile(opts.SpecPath)
	if err != nil {
		return fmt.Errorf("docs: read spec %s: %w", opts.SpecPath, err)
	}

	contentType := "application/yaml"
	switch ext := filepath.Ext(opts.SpecPath); ext {
	case ".json":
		contentType = "application/json"
	case ".yaml", ".yml":
	default:
		return fmt.Errorf("docs: unsupported spec file extension %q (want .yaml/.yml/.json)", ext)
	}

	prefix := opts.prefix()

	engine.GET(prefix, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIHTML(prefix+"/openapi")))
	})

	engine.GET(prefix+"/openapi", func(c *gin.Context) {
		c.Data(http.StatusOK, contentType, specBytes)
	})

	return nil
}

// swaggerUIHTML 渲染 Swagger UI 页面（通过 CDN 加载静态资源）.
func swaggerUIHTML(specURL string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>API Documentation</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js" crossorigin></script>
<script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-standalone-preset.js" crossorigin></script>
<script>
  window.onload = function () {
    window.ui = SwaggerUIBundle({
      url: "` + specURL + `",
      dom_id: "#swagger-ui",
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
      layout: "StandaloneLayout",
    });
  };
</script>
</body>
</html>`
}
