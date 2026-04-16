// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	if configPath != "" {
		// 从命令行选项指定的配置文件中读取
		viper.SetConfigFile(configPath)
	} else {
		// 查找用户主目录
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home dir: %w", err)
		}

		// 将用 `$HOME/.framework` 目录加入到配置文件的搜索路径中
		viper.AddConfigPath(filepath.Join(home, ".framework"))

		// 把当前目录加入到配置文件的搜索路径中
		viper.AddConfigPath(".")

		// 设置配置文件格式为 YAML
		viper.SetConfigType("yaml")

		// 配置文件名称（没有文件扩展名）
		viper.SetConfigName("config")
	}

	// 读取匹配的环境变量
	viper.AutomaticEnv()

	// 读取环境变量的前缀为 FRAMEWORK
	viper.SetEnvPrefix("FRAMEWORK")

	// 将 viper.Get(key) key 字符串中 '.' 和 '-' 替换为 '_'
	replacer := strings.NewReplacer(".", "_", "-", "_")
	viper.SetEnvKeyReplacer(replacer)

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 验证数据库配置
	if c.Database.MySQL.Host == "" {
		return fmt.Errorf("mysql host is required")
	}
	if c.Database.MySQL.Username == "" {
		return fmt.Errorf("mysql username is required")
	}
	if c.Database.MySQL.Database == "" {
		return fmt.Errorf("mysql database is required")
	}

	// 验证 JWT 配置
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}

	// 验证服务器配置
	if c.Server.HTTPAddr == "" {
		c.Server.HTTPAddr = ":8080" // 默认值
	}

	return nil
}
