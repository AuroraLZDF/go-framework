// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/auroralzdf/biddingAI.

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

// NewTreeService returns a new TreeService.
func NewTreeService(config TreeService) *TreeService {
	return &TreeService{
		PrimaryKey:  config.PrimaryKey,
		ParentKey:   config.ParentKey,
		ExpandedKey: "expanded",
		LeafKey:     "leaf",
		ChildrenKey: "children",
		Expanded:    false,
		Result:      make([]map[string]interface{}, 0),
		Level:       make(map[interface{}]int),
	}
}

// MakeTree is a function for make tree data.
//
//	@param data is a data for make tree.
//	@return a tree data.
func (t *TreeService) MakeTree(data []map[string]interface{}) []map[string]interface{} {
	dataset := t.buildData(data)
	t.Result = t.makeTreeCore(0, dataset, "normal")
	return t.Result
}

// MakeTreeForHTML 生成线性结构, 便于HTML输出, 参数同上.
func (t *TreeService) MakeTreeForHTML(data []map[string]interface{}) []map[string]interface{} {
	dataset := t.buildData(data)
	t.Result = t.makeTreeCore(0, dataset, "linear")
	return t.Result
}

// 构建树形结构
func (t *TreeService) buildData(data []map[string]interface{}) map[interface{}]map[interface{}]map[string]interface{} {
	r := make(map[interface{}]map[interface{}]map[string]interface{})
	for _, item := range data {
		id := item[t.PrimaryKey]
		parentID := item[t.ParentKey]
		if r[parentID] == nil {
			r[parentID] = make(map[interface{}]map[string]interface{})
		}
		r[parentID][id] = item
	}
	return r
}

// 递归生成树
func (t *TreeService) makeTreeCore(index interface{}, data map[interface{}]map[interface{}]map[string]interface{}, tType string) []map[string]interface{} {
	r := make([]map[string]interface{}, 0)
	for id, item := range data[index] {
		if tType == "normal" {
			if _, ok := data[id]; ok {
				item[t.ExpandedKey] = t.Expanded
				item[t.ChildrenKey] = t.makeTreeCore(id, data, tType)
			} else {
				item[t.LeafKey] = true
			}
			r = append(r, item)
		} else if tType == "linear" {
			parentID := item[t.ParentKey]
			t.Level[id] = t.Level[parentID] + 1
			item["level"] = t.Level[id]
			t.Result = append(t.Result, item)
			if _, ok := data[id]; ok {
				t.makeTreeCore(id, data, tType)
			}
			r = t.Result
		}
	}
	return r
}
