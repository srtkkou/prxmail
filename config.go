package prxmail

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goark/errs"
	"github.com/pelletier/go-toml/v2"
)

// 設定
type AppConfig struct {
	mtx         sync.RWMutex    // Mutex
	ctx         context.Context // コンテキスト
	version     string          // バージョン
	commitHash  string          // コミットハッシュ
	commitAt    time.Time       // コミット日時
	exePath     string          // 実行ファイルのパス
	host        string          // ホスト名
	smtpHost    string          // SMTPホスト名
	smtpPort    int             // SMTPポート番号
	username    string          // SASL Authユーザ
	password    string          // SASL Authパスワード
	discardFrom []string        // 拒否Fromアドレス
	from        string          // 送信元
	recipients  []string        // 送信先
	subject     string          // 件名
	body        string          // 本文
}

// 設定tomlファイル構造
type configTomlStructure struct {
	Host        string   // ホスト名
	Port        int      // ポート番号
	Username    string   // SASL Authユーザ
	Password    string   // SASL Authパスワード
	DiscardFrom []string // 拒否Fromアドレス
}

const (
	// コミット日時の形式
	commitAtLayout = "20060102-150405"
)

var (
	// 設定初期化排他
	configOnce sync.Once
	// 設定のインスタンス
	configPtr *AppConfig
)

var (
	// 実行ファイルパスが取得できない。
	ErrConfigExePath = errors.New("config.ErrConfigExePath")
	// 絶対パスが取得できない。
	ErrConfigAbsPath = errors.New("config.ErrConfigAbsPath")
	// ホスト名が取得できない。
	ErrConfigHostname = errors.New("config.ErrConfigHostname")
	// 設定ファイルが読み込めない。
	ErrConfigLoadToml = errors.New("config.ErrConfigLoadToml")
)

// 設定の取得
func Config() *AppConfig {
	// 設定の初期化
	configOnce.Do(func() {
		configPtr = &AppConfig{
			ctx:         context.Background(),
			commitHash:  "",
			version:     "0.1.0",
			discardFrom: make([]string, 0),
			recipients:  make([]string, 0),
		}
	})
	return configPtr
}

// 設定ファイルの読み取り
func (c *AppConfig) LoadConfig() (err error) {
	// 実行ファイルパスの更新
	if err = c.loadExePath(); err != nil {
		return err
	}
	// ホスト名の更新
	if err = c.loadHost(); err != nil {
		return err
	}
	// 設定toml設定ファイルの読み込み
	if err := c.loadToml(); err != nil {
		return err
	}
	return nil
}

// バージョンの更新
func (c *AppConfig) SetVersion(str string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.version = str
}

// バージョンの取得
func (c *AppConfig) Version() string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.version
}

// コミットハッシュの更新
func (c *AppConfig) SetCommitHash(str string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.commitHash = str
}

// コミットハッシュの取得
func (c *AppConfig) CommitHash() string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.commitHash
}

// コミット日時の更新
func (c *AppConfig) SetCommitAt(str string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	tm, err := time.Parse(commitAtLayout, str)
	if err == nil {
		c.commitAt = tm
	}
}

// コミット日時の取得
func (c *AppConfig) CommitAt() time.Time {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.commitAt
}

// 実行ファイルパスの取得
func (c *AppConfig) ExePath() string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.exePath
}

// ホスト名の取得
func (c *AppConfig) Host() string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.host
}

// SMTPホスト名の更新
func (c *AppConfig) SetSMTPHost(str string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.smtpHost = str
}

// SMTPホスト名の取得
func (c *AppConfig) SMTPHost() string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.smtpHost
}

// SMTPポート番号の更新
func (c *AppConfig) SetSMTPPort(num int) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.smtpPort = num
}

// SMTPポート番号の取得
func (c *AppConfig) SMTPPort() int {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.smtpPort
}

// SASL Authユーザ名の更新
func (c *AppConfig) SetUsername(str string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.username = str
}

// SASL Authユーザ名の取得
func (c *AppConfig) Username() string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.username
}

// SASL Authパスワードの更新
func (c *AppConfig) SetPassword(str string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.password = str
}

// SASL Authパスワードの取得
func (c *AppConfig) Password() string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.password
}

// 拒否アドレスの更新
func (c *AppConfig) SetDiscardFrom(strs []string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.discardFrom = strs
}

// 拒否アドレスの取得
func (c *AppConfig) DiscardFrom() []string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.discardFrom
}

// 送信元の更新
func (c *AppConfig) SetFrom(str string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.from = str
}

// 送信元の取得
func (c *AppConfig) From() string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.from
}

// 送信先の更新
func (c *AppConfig) SetRecipients(strs []string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.recipients = strs
}

// 送信先の取得
func (c *AppConfig) Recipients() []string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.recipients
}

// 件名の更新
func (c *AppConfig) SetSubject(str string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.subject = str
}

// 件名の取得
func (c *AppConfig) Subject() string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.subject
}

// 本文の更新
func (c *AppConfig) SetBody(str string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.body = str
}

// 本文の取得
func (c *AppConfig) Body() string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.body
}

// 実行ファイルのディレクトリ
func (c *AppConfig) ExeDir() string {
	return filepath.Dir(c.ExePath())
}

// ログファイルのパス
func (c *AppConfig) LogPath() string {
	return filepath.Join(c.ExeDir(), "prxmail.log")
}

// 設定ファイルのパス
func (c *AppConfig) ConfigFilePath() string {
	return filepath.Join(c.ExeDir(), "prxmail.toml")
}

// ログ用にマスクしたパスワード
func (c *AppConfig) MaskedPassword() string {
	pw := c.Password()
	size := len(pw)
	// 2文字以内の場合はすべてマスクする。
	if size <= 2 {
		return strings.Repeat("*", size)
	}
	// その他の場合、最初の2文字以外をマスクする。
	return pw[0:2] + strings.Repeat("*", (size-2))
}

// ポート番号付きホスト名
func (c *AppConfig) SMTPHostWithPort() string {
	return net.JoinHostPort(c.SMTPHost(), strconv.Itoa(c.SMTPPort()))
}

// 実行ファイルパスの設定
func (c *AppConfig) loadExePath() (err error) {
	// 実行ファイルのパスの取得
	var exePath string
	exePath, err = os.Executable()
	if err != nil {
		return errs.Wrap(ErrConfigExePath, errs.WithCause(err))
	}
	// 絶対パスの取得
	absPath, err := filepath.Abs(exePath)
	if err != nil {
		return errs.Wrap(ErrConfigAbsPath, errs.WithCause(err))
	}
	// 実行パスの更新
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.exePath = absPath
	return nil
}

// 実行ホスト名の設定
func (c *AppConfig) loadHost() (err error) {
	// 実行環境のホスト名の取得
	host, err := os.Hostname()
	if err != nil {
		err = errs.Wrap(ErrConfigHostname, errs.WithCause(err))
		return err
	}
	// ホスト名の更新
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.host = host
	return nil
}

// 設定tomlファイルの読み込み
func (c *AppConfig) loadToml() error {
	// 設定ファイルパスの読み込み
	tomlPath := filepath.Join(c.ExeDir(), "prxmail.toml")
	blob, err := os.ReadFile(tomlPath)
	if err != nil {
		return errs.Wrap(ErrConfigLoadToml, errs.WithCause(err),
			errs.WithContext("path", tomlPath))
	}
	// TOMLの読み込み
	values := configTomlStructure{}
	if err := toml.Unmarshal(blob, &values); err != nil {
		return errs.Wrap(ErrConfigLoadToml, errs.WithCause(err))
	}
	// 値の設定
	c.SetSMTPHost(values.Host)
	c.SetSMTPPort(values.Port)
	c.SetUsername(values.Username)
	c.SetPassword(values.Password)
	c.SetDiscardFrom(values.DiscardFrom)
	return nil
}
