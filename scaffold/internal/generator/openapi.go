// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package generator

import (
	"path/filepath"
)

// generateOpenAPISpec 生成 OpenAPI 示例规范文件 docs/openapi.yaml.
// 项目可通过 core/docs 在线托管该文件与 Swagger UI（见 main.go 中的注释示例）。
func generateOpenAPISpec(dir, name string) error {
	content := `openapi: 3.0.3
info:
  title: ` + name + ` API
  description: |
    ` + name + ` 的 OpenAPI 规范（OAS 3.0）。
    在 main.go 的 RegisterRoutes 中调用 docs.Register 即可启用在线文档：
      - GET /docs            Swagger UI
      - GET /docs/openapi    本规范文件
  version: 1.0.0
servers:
  - url: http://localhost:8080
tags:
  - name: system
    description: 系统与运维接口
paths:
  /health:
    get:
      tags:
        - system
      summary: 业务健康检查
      responses:
        '200':
          description: 服务正常
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    example: 200
                  message:
                    type: string
                    example: ok
                  content:
                    type: object
  /healthz:
    get:
      tags:
        - system
      summary: 存活探针（liveness）
      responses:
        '200':
          description: 进程存活
`

	return writeFile(filepath.Join(dir, "docs", "openapi.yaml"), content)
}
