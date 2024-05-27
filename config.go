package prxmail

import (
	"context"
	"fmt"
	"strings"
)

type (
	// 設定
	Config struct {
		// コンテキスト
		Ctx context.Context
		// バージョン
		Version string
		// Gitリビジョン
		Revision string
		// ホスト
		Host string
		// ポート
		Port string
		// SASL Authユーザ
		Username string
		// SASL Authパスワード
		Password string
		// 送信元
		From string
		// 送信先
		Recipients []string
		// 件名
		Subject string
		// 本文
		Body string
		// ヘルプの表示が必要か？
		IsHelpRequested bool
		// バージョンの表示が必要か？
		IsVersionRequested bool
	}
)

var (
	// 設定のインスタンス
	configInstance = &Config{
		Ctx:     context.Background(),
		Version: "v0.0.1",
	}
)

// 設定の取得
func GetConfigInstance() *Config {
	return configInstance
}

// バージョン情報の取得
func (c *Config) VersionInfo() string {
	return fmt.Sprintf("%s (build:%s)", c.Version, c.Revision)
}

// ログ出力するパスワード
func (c *Config) LogPassword() string {
	size := len(c.Password)
	// 2文字以内の場合はすべてマスクする。
	if size <= 2 {
		return strings.Repeat("*", size)
	}
	// その他の場合、最初の2文字以外をマスクする。
	return c.Password[0:2] + strings.Repeat("*", (size-2))
}

// TLSサーバ
func (c *Config) TlsServer() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
