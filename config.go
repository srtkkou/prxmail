package prxmail

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/goark/errs"
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
		// 実行ファイルのパス
		ExePath string
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

var (
	// 実行ファイルパスが取得できない。
	ErrConfigExePath = errors.New("prxmail.config.ErrConfigExePath")
)

// 設定の取得
func GetConfigInstance() *Config {
	return configInstance
}

// 実行ファイルパスの設定
func (c *Config) LoadExePath() (err error) {
	// 実行ファイルのパスの取得
	var exePath string
	exePath, err = os.Executable()
	if err != nil {
		return errs.Wrap(ErrConfigExePath, errs.WithCause(err))
	}
	// 絶対パスの取得
	c.ExePath, err = filepath.Abs(exePath)
	if err != nil {
		return errs.Wrap(ErrConfigExePath, errs.WithCause(err))
	}
	return nil
}

// 実行ファイルのディレクトリ
func (c *Config) ExeDir() string {
	return filepath.Dir(c.ExePath)
}

// ログファイルのパス
func (c *Config) LogPath() string {
	return filepath.Join(c.ExeDir(), "prxmail.log")
}

// 環境変数ファイルのパス
func (c *Config) EnvPath() string {
	return filepath.Join(c.ExeDir(), "prxmail.env")
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

// ポート番号付きホスト名
func (c *Config) HostWithPort() string {
	return net.JoinHostPort(c.Host, c.Port)
}
