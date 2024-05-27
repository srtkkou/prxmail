package prxmail

import (
	"fmt"
)

type (
	Config struct {
		Version  string
		Revision string
		Host     string
		Port     string
		Username string
		Password string
	}
)

// 設定の初期化
var configInstance = &Config{
	Version: "v0.0.1",
}

// 設定の取得
func GetConfig() *Config {
	return configInstance
}

// バージョン情報の取得
func (c *Config) VersionInfo() string {
	return fmt.Sprintf("%s (build:%s)", c.Version, c.Revision)
}

// TLSサーバ
func (c *Config) TlsServer() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
