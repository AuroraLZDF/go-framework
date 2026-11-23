// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

// 代码生成：scaffold generate model / crud.
// 必须在生成项目的根目录（go.mod 所在目录）执行，
// 以便读取 module 路径供 import 使用。

package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// modelNamePattern 限制模型名：小写字母开头，仅允许小写字母、数字、'_'.
var modelNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// fieldNamePattern 限制字段名：小写字母开头，仅允许小写字母、数字、'_'.
var fieldNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// goTypeMap CLI 字段类型到 Go 类型的映射.
var goTypeMap = map[string]string{
	"string":  "string",
	"text":    "string",
	"bool":    "bool",
	"int":     "int",
	"int8":    "int8",
	"int16":   "int16",
	"int32":   "int32",
	"int64":   "int64",
	"uint":    "uint",
	"uint8":   "uint8",
	"uint16":  "uint16",
	"uint32":  "uint32",
	"uint64":  "uint64",
	"float":   "float64",
	"float32": "float32",
	"float64": "float64",
	"time":    "time.Time",
}

// supportedTypeOrder 支持类型的展示顺序（供错误提示与帮助信息）.
var supportedTypeOrder = []string{"string", "text", "bool", "int", "int8", "int16", "int32", "int64",
	"uint", "uint8", "uint16", "uint32", "uint64", "float", "float32", "float64", "time"}

// fieldTypeGormTags 需要额外 gorm 类型标签的字段类型.
var fieldTypeGormTags = map[string]string{
	"text": `gorm:"type:text"`,
}

// Field 生成的模型字段.
type Field struct {
	// Name Go 结构体字段名（导出驼峰，如 UserName）
	Name string
	// Column 数据库列名与 JSON 键（snake_case）
	Column string
	// Type Go 类型（如 string、time.Time）
	Type string
	// GormTag 额外的 gorm 标签（可为空）
	GormTag string
}

// GenerateOptions 代码生成选项.
type GenerateOptions struct {
	// Name 模型名（snake_case，如 user_profile）
	Name string
	// Fields 字段列表
	Fields []Field
	// Force 覆盖已存在的文件
	Force bool
}

// StructName 返回模型名对应的 Go 结构体名（如 user_profile → UserProfile）.
func StructName(name string) string {
	parts := strings.Split(name, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}

	return strings.Join(parts, "")
}

// ParseFields 解析 CLI 传入的字段描述，如 "name:string,email:string,birth:time".
func ParseFields(spec string) ([]Field, error) {
	parts := strings.Split(spec, ",")
	fields := make([]Field, 0, len(parts))

	seen := make(map[string]struct{})
	for _, raw := range parts {
		p := strings.TrimSpace(raw)
		if p == "" {
			continue
		}

		name, fieldType, found := strings.Cut(p, ":")
		if !found {
			return nil, fmt.Errorf("invalid field %q: expected <name>:<type>", p)
		}
		name = strings.TrimSpace(name)
		fieldType = strings.TrimSpace(fieldType)

		if !fieldNamePattern.MatchString(name) {
			return nil, fmt.Errorf("invalid field name %q: must start with a lowercase letter and contain only lowercase letters, digits and '_'", name)
		}

		goType, ok := goTypeMap[fieldType]
		if !ok {
			return nil, fmt.Errorf("invalid field type %q for field %q (supported: %s)", fieldType, name, supportedTypes())
		}

		goName := StructName(name)
		if _, dup := seen[goName]; dup {
			return nil, fmt.Errorf("duplicate field %q", name)
		}
		seen[goName] = struct{}{}

		fields = append(fields, Field{
			Name:    goName,
			Column:  name,
			Type:    goType,
			GormTag: fieldTypeGormTags[fieldType],
		})
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields specified")
	}

	return fields, nil
}

// supportedTypes 返回支持的类型列表（按固定顺序）.
func supportedTypes() string {
	types := make([]string, 0, len(supportedTypeOrder))
	for _, t := range supportedTypeOrder {
		if _, ok := goTypeMap[t]; ok {
			types = append(types, t)
		}
	}

	return strings.Join(types, ", ")
}

// modulePath 读取项目根目录 go.mod 的 module 声明.
func modulePath(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod")) // #nosec G304 -- 固定读取用户在项目根目录下执行的 go.mod
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w (run this command in the project root)", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if mod, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(mod), nil
		}
	}

	return "", fmt.Errorf("no module declaration found in go.mod")
}

// GenerateModel 生成 GORM 模型文件 internal/model/<name>.go，返回生成的文件路径.
func GenerateModel(opts GenerateOptions) ([]string, error) {
	if err := validateGenerateOptions(opts); err != nil {
		return nil, err
	}

	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	target := filepath.Join(root, "internal", "model", opts.Name+".go")
	if err := checkTarget(target, opts.Force); err != nil {
		return nil, err
	}

	if err := writeFile(target, renderModel(opts.Name, opts.Fields)); err != nil {
		return nil, err
	}

	return []string{relPath(root, target)}, nil
}

// GenerateCRUD 生成 model + store（CRUD 仓储）+ controller 三件套，返回生成的文件路径.
func GenerateCRUD(opts GenerateOptions) ([]string, error) {
	if err := validateGenerateOptions(opts); err != nil {
		return nil, err
	}

	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	mod, err := modulePath(root)
	if err != nil {
		return nil, err
	}

	structName := StructName(opts.Name)

	targets := []string{
		filepath.Join(root, "internal", "model", opts.Name+".go"),
		filepath.Join(root, "internal", "store", opts.Name+"_store.go"),
		filepath.Join(root, "internal", "controller", opts.Name+"_controller.go"),
	}
	for _, target := range targets {
		if err := checkTarget(target, opts.Force); err != nil {
			return nil, err
		}
	}

	contents := []string{
		renderModel(opts.Name, opts.Fields),
		renderStore(structName, mod),
		renderController(opts.Name, structName, mod, opts.Fields),
	}

	paths := make([]string, 0, len(targets))
	for i, target := range targets {
		if err := writeFile(target, contents[i]); err != nil {
			return nil, err
		}

		paths = append(paths, relPath(root, target))
	}

	return paths, nil
}

// validateGenerateOptions 校验模型名与字段列表.
func validateGenerateOptions(opts GenerateOptions) error {
	if !modelNamePattern.MatchString(opts.Name) {
		return fmt.Errorf("invalid model name %q: must start with a lowercase letter and contain only lowercase letters, digits and '_'", opts.Name)
	}

	if len(opts.Fields) == 0 {
		return fmt.Errorf("no fields specified")
	}

	return nil
}

// checkTarget 目标文件已存在且未指定 --force 时返回错误.
func checkTarget(path string, force bool) error {
	if force {
		return nil
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists (use --force to overwrite)", relPath("", path))
	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}

// relPath 返回相对于 root 的路径；root 为空时返回原路径.
func relPath(root, path string) string {
	if root == "" {
		return path
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}

	return rel
}

// renderModel 渲染 GORM 模型文件内容.
func renderModel(name string, fields []Field) string {
	structName := StructName(name)

	var imports []string
	imports = append(imports, `"gorm.io/gorm"`)
	if hasTimeField(fields) {
		imports = append(imports, `"time"`)
	}

	var b strings.Builder
	b.WriteString("package model\n\nimport (\n")
	for _, imp := range imports {
		b.WriteString("\t" + imp + "\n")
	}
	b.WriteString(")\n\n")

	fmt.Fprintf(&b, "// %s %s 模型.\ntype %s struct {\n\tgorm.Model\n\n", structName, name, structName)
	for _, f := range fields {
		tag := fmt.Sprintf("json:%q", f.Column)
		if f.GormTag != "" {
			tag = f.GormTag + " " + tag
		}
		fmt.Fprintf(&b, "\t%s %s `%s`\n", f.Name, f.Type, tag)
	}
	b.WriteString("}\n")

	return b.String()
}

// hasTimeField 判断字段中是否包含 time.Time 类型.
func hasTimeField(fields []Field) bool {
	for _, f := range fields {
		if f.Type == "time.Time" {
			return true
		}
	}

	return false
}

// renderStore 渲染 store 文件内容（基于框架 store.Repository 基类）.
func renderStore(structName, module string) string {
	var b strings.Builder
	b.WriteString(`package store

import (
	"gorm.io/gorm"

	"github.com/AuroraLZDF/go-framework/core/database/store"
	"` + module + `/internal/model"
)

// ` + structName + `Store ` + structName + ` 仓储：基于框架的泛型 CRUD 基类，开箱即用分页/过滤/软删除.
type ` + structName + `Store struct {
	*store.Repository[model.` + structName + `]
}

// New` + structName + `Store 创建 ` + structName + ` 仓储.
func New` + structName + `Store(db *gorm.DB) (*` + structName + `Store, error) {
	repo, err := store.NewRepository[model.` + structName + `](db)
	if err != nil {
		return nil, err
	}

	return &` + structName + `Store{Repository: repo}, nil
}
`)

	return b.String()
}

// renderController 渲染 controller 文件内容（CRUD 接口）.
func renderController(name, structName, module string, fields []Field) string {
	var b strings.Builder

	b.WriteString(`package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/AuroraLZDF/go-framework/core/errno"
	"github.com/AuroraLZDF/go-framework/core/response"
	"github.com/AuroraLZDF/go-framework/core/util"
	fwstore "github.com/AuroraLZDF/go-framework/core/database/store"
	"` + module + `/internal/model"
	"` + module + `/internal/store"
)

// ` + structName + `Controller ` + name + ` CRUD 接口.
type ` + structName + `Controller struct {
	store *store.` + structName + `Store
}

// New` + structName + `Controller 创建 ` + name + ` 控制器.
func New` + structName + `Controller(s *store.` + structName + `Store) *` + structName + `Controller {
	return &` + structName + `Controller{store: s}
}

// RegisterRoutes 注册路由：/` + name + ` 增删改查.
func (h *` + structName + `Controller) RegisterRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/` + name + `")
	{
		r.POST("", h.Create)
		r.GET("", h.List)
		r.GET("/:id", h.Get)
		r.PUT("/:id", h.Update)
		r.DELETE("/:id", h.Delete)
	}
}

// Create` + structName + `Request 创建请求体.
type Create` + structName + `Request struct {
`)

	writeRequestFields(&b, fields)
	b.WriteString(`}

// Update` + structName + `Request 更新请求体.
type Update` + structName + `Request struct {
`)
	writeRequestFields(&b, fields)

	b.WriteString(`}

// Create 创建 ` + name + `.
func (h *` + structName + `Controller) Create(c *gin.Context) {
	var req Create` + structName + `Request
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errno.ErrBind)
		return
	}

	item := model.` + structName + `{
`)

	for _, f := range fields {
		fmt.Fprintf(&b, "\t\t%s: req.%s,\n", f.Name, f.Name)
	}

	b.WriteString(`	}
	if err := h.store.Create(c.Request.Context(), &item); err != nil {
		response.Error(c, errno.ErrDBQuery)
		return
	}

	response.Success(c, item, "ok")
}

// Get 按 ID 查询 ` + name + `.
func (h *` + structName + `Controller) Get(c *gin.Context) {
	id, err := util.ParamToInt64(c, "id")
	if err != nil || id <= 0 {
		response.Error(c, errno.ErrInvalidParameter)
		return
	}

	item, err := h.store.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, errno.ErrNotFound)
		return
	}

	response.Success(c, item, "ok")
}

// List 分页查询 ` + name + `（page/page-size 查询参数）.
func (h *` + structName + `Controller) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page-size", strconv.Itoa(fwstore.DefaultPageSize)))
	if err != nil || pageSize < 1 {
		pageSize = fwstore.DefaultPageSize
	}

	result, err := h.store.List(c.Request.Context(), fwstore.Query{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Error(c, errno.ErrDBQuery)
		return
	}

	response.Success(c, result, "ok")
}

// Update 更新 ` + name + `.
func (h *` + structName + `Controller) Update(c *gin.Context) {
	id, err := util.ParamToInt64(c, "id")
	if err != nil || id <= 0 {
		response.Error(c, errno.ErrInvalidParameter)
		return
	}

	var req Update` + structName + `Request
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errno.ErrBind)
		return
	}

	item, err := h.store.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, errno.ErrNotFound)
		return
	}

`)

	for _, f := range fields {
		fmt.Fprintf(&b, "\titem.%s = req.%s\n", f.Name, f.Name)
	}

	b.WriteString(`	if err := h.store.Update(c.Request.Context(), item); err != nil {
		response.Error(c, errno.ErrDBQuery)
		return
	}

	response.Success(c, item, "ok")
}

// Delete 删除 ` + name + `（软删除）.
func (h *` + structName + `Controller) Delete(c *gin.Context) {
	id, err := util.ParamToInt64(c, "id")
	if err != nil || id <= 0 {
		response.Error(c, errno.ErrInvalidParameter)
		return
	}

	if err := h.store.Delete(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, errno.ErrNotFound)
		return
	}

	response.Success(c, nil, "ok")
}
`)

	return b.String()
}

// writeRequestFields 写入请求体结构体字段（string 字段标记 binding:required）.
func writeRequestFields(b *strings.Builder, fields []Field) {
	for _, f := range fields {
		tag := fmt.Sprintf("`json:%q", f.Column)
		if f.Type == "string" {
			tag += " binding:\"required\""
		}
		tag += "`"

		fmt.Fprintf(b, "\t%s %s %s\n", f.Name, f.Type, tag)
	}
}
