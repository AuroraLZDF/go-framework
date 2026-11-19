// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package store

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testUser 测试模型：包含 gorm.Model 以覆盖软删除行为.
type testUser struct {
	gorm.Model
	Name string
	Age  int
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&testUser{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return db
}

func newTestRepo(t *testing.T) (*Repository[testUser], *gorm.DB) {
	t.Helper()

	db := newTestDB(t)
	repo, err := NewRepository[testUser](db)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}

	return repo, db
}

func seedUsers(t *testing.T, repo *Repository[testUser]) {
	t.Helper()

	users := []testUser{
		{Name: "alice", Age: 30},
		{Name: "alice", Age: 20},
		{Name: "bob", Age: 25},
	}
	for i := range users {
		if err := repo.Create(context.Background(), &users[i]); err != nil {
			t.Fatalf("seed user %s: %v", users[i].Name, err)
		}
	}
}

func TestNewRepositoryNilDB(t *testing.T) {
	if _, err := NewRepository[testUser](nil); err == nil {
		t.Error("NewRepository with nil db should fail")
	}
}

func TestCreateAndGetByID(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	user := testUser{Name: "alice", Age: 30}
	if err := repo.Create(ctx, &user); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if user.ID == 0 {
		t.Error("Create should set primary key")
	}

	got, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "alice" || got.Age != 30 {
		t.Errorf("GetByID = %+v, want name=alice age=30", got)
	}

	if _, err := repo.GetByID(ctx, 9999); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetByID missing record should return ErrNotFound, got %v", err)
	}
}

func TestUpdate(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	user := testUser{Name: "alice"}
	if err := repo.Create(ctx, &user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	user.Name = "alice2"
	user.Age = 42
	if err := repo.Update(ctx, &user); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "alice2" || got.Age != 42 {
		t.Errorf("Update result = %+v, want name=alice2 age=42", got)
	}

	// 更新不存在的记录应返回 ErrNotFound
	missing := testUser{Model: gorm.Model{ID: 9999}, Name: "ghost"}
	if err := repo.Update(ctx, &missing); !errors.Is(err, ErrNotFound) {
		t.Errorf("Update missing record should return ErrNotFound, got %v", err)
	}
}

func TestSoftDelete(t *testing.T) {
	repo, db := newTestRepo(t)
	ctx := context.Background()

	user := testUser{Name: "alice"}
	if err := repo.Create(ctx, &user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, user.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// 软删除后默认查询不可见
	if _, err := repo.GetByID(ctx, user.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetByID after soft delete should return ErrNotFound, got %v", err)
	}

	// 物理记录仍存在，DeletedAt 已填充
	var raw testUser
	if err := db.Unscoped().First(&raw, user.ID).Error; err != nil {
		t.Fatalf("unscoped query: %v", err)
	}
	if raw.DeletedAt.Valid != true {
		t.Error("DeletedAt should be set after soft delete")
	}

	if err := repo.Delete(ctx, 9999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete missing record should return ErrNotFound, got %v", err)
	}
}

func TestListPagination(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	seedUsers(t, repo)

	page, err := repo.List(ctx, Query{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page.Total != 3 {
		t.Errorf("Total = %d, want 3", page.Total)
	}
	if page.Page != 2 || page.PageSize != 2 {
		t.Errorf("Page/PageSize = %d/%d, want 2/2", page.Page, page.PageSize)
	}
	if len(page.Items) != 1 {
		t.Errorf("items on page 2 = %d, want 1", len(page.Items))
	}

	// 越界页返回空切片
	empty, err := repo.List(ctx, Query{Page: 10, PageSize: 2})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if empty.Items == nil || len(empty.Items) != 0 {
		t.Errorf("out-of-range page should return empty slice, got %+v", empty.Items)
	}
}

func TestListFilters(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	seedUsers(t, repo)

	page, err := repo.List(ctx, Query{Filters: map[string]any{"name": "alice"}})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// 2 条 alice，PageSize 未指定时默认 20，全部返回
	if page.Total != 2 {
		t.Errorf("filtered Total = %d, want 2", page.Total)
	}
	if len(page.Items) != 2 {
		t.Errorf("filtered items = %d, want 2", len(page.Items))
	}

	page, err = repo.List(ctx, Query{Filters: map[string]any{"name": "alice", "age": 20}})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page.Total != 1 || page.Items[0].Age != 20 {
		t.Errorf("multi-filter result = %+v, want 1 record with age=20", page.Items)
	}
}

func TestListInvalidFilters(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	for _, filter := range []map[string]any{
		{"name = ''; DROP TABLE test_users; --": "x"},
		{"name OR 1=1": "x"},
		{"": "x"},
	} {
		if _, err := repo.List(ctx, Query{Filters: filter}); !errors.Is(err, ErrInvalidField) {
			t.Errorf("invalid filter %+v should return ErrInvalidField, got %v", filter, err)
		}
	}
}

func TestListSort(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	seedUsers(t, repo)

	page, err := repo.List(ctx, Query{Sort: "age asc"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Items) != 3 || page.Items[0].Age != 20 || page.Items[2].Age != 30 {
		t.Errorf("sort age asc result = %+v, want ages [20 25 30]", page.Items)
	}

	page, err = repo.List(ctx, Query{Sort: "age desc"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page.Items[0].Age != 30 {
		t.Errorf("sort age desc first = %d, want 30", page.Items[0].Age)
	}

	// 大写方向与多余空格
	if _, err := repo.List(ctx, Query{Sort: "  age   DESC "}); err != nil {
		t.Errorf("sort with extra spaces failed: %v", err)
	}

	// 未指定排序时默认 id 倒序
	page, err = repo.List(ctx, Query{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page.Items[0].ID < page.Items[1].ID {
		t.Error("default sort should be id DESC")
	}
}

func TestListInvalidSort(t *testing.T) {
	repo, _ := newTestRepo(t)

	cases := []struct {
		sort string
		want error
	}{
		{"age; DROP TABLE test_users", ErrInvalidField},
		{"age bogus", ErrInvalidSort},
		{"age asc;", ErrInvalidSort},
	}

	for _, c := range cases {
		_, err := repo.List(context.Background(), Query{Sort: c.sort})
		if !errors.Is(err, c.want) {
			t.Errorf("sort %q: got %v, want %v", c.sort, err, c.want)
		}
	}
}

func TestCount(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	seedUsers(t, repo)

	total, err := repo.Count(ctx, Query{Filters: map[string]any{"name": "alice"}})
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if total != 2 {
		t.Errorf("Count = %d, want 2", total)
	}
}

func TestListNormalizeDefaults(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	seedUsers(t, repo)

	// Page/PageSize 均非法时使用默认值
	page, err := repo.List(ctx, Query{Page: -5, PageSize: 9999})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page.Page != 1 {
		t.Errorf("Page = %d, want 1", page.Page)
	}
	if page.PageSize != MaxPageSize {
		t.Errorf("PageSize = %d, want %d", page.PageSize, MaxPageSize)
	}
}
