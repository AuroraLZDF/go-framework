// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package util

// TreeService is a service for tree data.
type TreeService struct {
	PrimaryKey  string
	ParentKey   string
	ExpandedKey string
	LeafKey     string
	ChildrenKey string
	Expanded    bool
	Result      []map[string]interface{}
	Level       map[interface{}]int
}

const (
	defaultExpandedKey = "expanded"
	defaultLeafKey     = "leaf"
	defaultChildrenKey = "children"
)

// NewTreeService returns a new TreeService.
// 未显式配置的 Key/Expanded 使用默认值；Result/Level 每次实例化都是独立的.
func NewTreeService(config TreeService) *TreeService {
	expandedKey := config.ExpandedKey
	if expandedKey == "" {
		expandedKey = defaultExpandedKey
	}
	leafKey := config.LeafKey
	if leafKey == "" {
		leafKey = defaultLeafKey
	}
	childrenKey := config.ChildrenKey
	if childrenKey == "" {
		childrenKey = defaultChildrenKey
	}

	return &TreeService{
		PrimaryKey:  config.PrimaryKey,
		ParentKey:   config.ParentKey,
		ExpandedKey: expandedKey,
		LeafKey:     leafKey,
		ChildrenKey: childrenKey,
		Expanded:    config.Expanded,
		Result:      make([]map[string]interface{}, 0),
		Level:       make(map[interface{}]int),
	}
}

// MakeTree is a function for make tree data.
//
//	@param data is a data for make tree.
//	@return a tree data.
func (t *TreeService) MakeTree(data []map[string]interface{}) []map[string]interface{} {
	t.reset()
	dataset := t.buildData(data)
	t.Result = t.makeTreeCore(0, dataset, "normal")

	return t.Result
}

// MakeTreeForHTML 生成线性结构, 便于HTML输出, 参数同上.
func (t *TreeService) MakeTreeForHTML(data []map[string]interface{}) []map[string]interface{} {
	t.reset()
	dataset := t.buildData(data)
	t.Result = t.makeTreeCore(0, dataset, "linear")

	return t.Result
}

// reset 清空上一次构建结果，保证重复调用不会重复追加/串味.
func (t *TreeService) reset() {
	t.Result = make([]map[string]interface{}, 0)
	t.Level = make(map[interface{}]int)
}

// 构建树形结构；每个 item 做浅拷贝，避免修改调用方传入的 map.
func (t *TreeService) buildData(data []map[string]interface{}) map[interface{}]map[interface{}]map[string]interface{} {
	r := make(map[interface{}]map[interface{}]map[string]interface{}, len(data))
	for _, item := range data {
		if item == nil {
			continue
		}

		id := item[t.PrimaryKey]
		parentID := item[t.ParentKey]
		if r[parentID] == nil {
			r[parentID] = make(map[interface{}]map[string]interface{})
		}

		clone := make(map[string]interface{}, len(item)+3)
		for k, v := range item {
			clone[k] = v
		}
		r[parentID][id] = clone
	}

	return r
}

// 递归生成树
func (t *TreeService) makeTreeCore(index interface{}, data map[interface{}]map[interface{}]map[string]interface{}, tType string) []map[string]interface{} {
	r := make([]map[string]interface{}, 0)

	for id, item := range data[index] {
		switch tType {
		case "normal":
			if _, ok := data[id]; ok {
				item[t.ExpandedKey] = t.Expanded
				item[t.ChildrenKey] = t.makeTreeCore(id, data, tType)
			} else {
				item[t.LeafKey] = true
			}
			r = append(r, item)
		case "linear":
			parentID := item[t.ParentKey]
			t.Level[id] = t.Level[parentID] + 1
			item["level"] = t.Level[id]
			r = append(r, item)
			if _, ok := data[id]; ok {
				r = append(r, t.makeTreeCore(id, data, tType)...)
			}
		}
	}

	return r
}
