// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package verflag

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"

	"github.com/AuroraLZDF/go-framework/core/version"
)

// VersionValue 表示 --version 标志的状态值.
type VersionValue int

// 定义一些常量.
const (
	VersionFalse VersionValue = 0
	VersionTrue  VersionValue = 1
	VersionRaw   VersionValue = 2
)

const (
	strRawVersion   = "raw"
	versionFlagName = "version"
)

var versionFlag = Version(versionFlagName, VersionFalse, "Print version information and quit.")

func (v *VersionValue) IsBoolFlag() bool {
	return true
}

func (v *VersionValue) Get() interface{} {
	return v
}

// String 实现了 pflag.Value 接口中的 String 方法.
func (v *VersionValue) String() string {
	if *v == VersionRaw {
		return strRawVersion
	}

	return fmt.Sprintf("%v", bool(*v == VersionTrue))
}

// Set 实现了 pflag.Value 接口中的 Set 方法.
func (v *VersionValue) Set(s string) error {
	if s == strRawVersion {
		*v = VersionRaw

		return nil
	}
	boolVal, err := strconv.ParseBool(s)
	if boolVal {
		*v = VersionTrue
	} else {
		*v = VersionFalse
	}

	return err
}

// Type 实现了 pflag.Value 接口中的 Type 方法.
func (v *VersionValue) Type() string {
	return "version"
}

// VersionVar 定义了一个具有指定名称和用法的标志.
func VersionVar(p *VersionValue, name string, value VersionValue, usage string) {
	*p = value
	pflag.Var(p, name, usage)
	// `--version` 等价于 `--version=true`
	pflag.Lookup(name).NoOptDefVal = "true"
}

// Version 包装了 VersionVar 函数.
func Version(name string, value VersionValue, usage string) *VersionValue {
	p := new(VersionValue)
	VersionVar(p, name, value, usage)

	return p
}

// PrintAndExitIfRequested 将检查是否传递了 `--version` 标志，如果是，则打印版本并退出.
func PrintAndExitIfRequested() {
	if *versionFlag == VersionRaw {
		fmt.Printf("%#v\n", version.Get())
		os.Exit(0)
	} else if *versionFlag == VersionTrue {
		fmt.Printf("%s\n", version.Get())
		os.Exit(0)
	}
}
