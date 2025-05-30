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
	"github.com/pelletier/go-toml/v2"
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
		// 拒否Fromアドレス
		DiscardFrom []string
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
	// 設定ファイルが読み込めない。
	ErrConfigLoadToml = errors.New("prxmail.config.ErrConfigLoadToml")
)

// 設定の初期化
func LoadConfig(revision string) (*Config, error) {
	// 実行ファイル位置の取得
	if err := configInstance.loadExePath(); err != nil {
		return nil, err
	}
	// toml設定ファイルの読み込み
	if err := configInstance.loadToml(); err != nil {
		return nil, err
	}
	return configInstance, nil
}

// 設定の取得
func GetConfigInstance() *Config {
	return configInstance
}

// 実行ファイルのディレクトリ
func (c *Config) ExeDir() string {
	return filepath.Dir(c.ExePath)
}

// ログファイルのパス
func (c *Config) LogPath() string {
	return filepath.Join(c.ExeDir(), "prxmail.log")
}

// 設定ファイルのパス
func (c *Config) ConfigFilePath() string {
	return filepath.Join(c.ExeDir(), "prxmail.toml")
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

// 実行ファイルパスの設定
func (c *Config) loadExePath() (err error) {
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

// tomlファイルのロード
func (c *Config) loadToml() error {
	// 設定ファイルパスの読み込み
	tomlPath := filepath.Join(c.ExeDir(), "prxmail.toml")
	blob, err := os.ReadFile(tomlPath)
	if err != nil {
		return errs.Wrap(ErrConfigLoadToml, errs.WithCause(err),
			errs.WithContext("path", tomlPath))
	}
	// TOMLの読み込み
	if err := toml.Unmarshal(blob, c); err != nil {
		return errs.Wrap(ErrConfigLoadToml, errs.WithCause(err))
	}
	return nil
}
