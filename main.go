package prxmail

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	"github.com/goark/errs"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

var (
	// バージョン
	Version = "v0.0.1"
	// 実行ファイルパスロードエラー
	ErrMainExePath = errors.New("prxmail.main.ErrMainExePath")
	// envファイルロードエラー
	ErrMainDotenvLoad = errors.New("prxmail.main.ErrMainDotenvLoad")
	// 引数解析エラー
	ErrMainArgs = errors.New("prxmail.main.ErrMainArgs")
	// ターミナルの標準入力は受け付けない
	ErrMainTermStdin = errors.New("prxmail.main.ErrMainTermStdin")
	// 標準入力読み込みエラー
	ErrMainStdinRead = errors.New("prxmail.main.ErrMainStdinRead")
	// SMTPメール送信エラー
	ErrMainSmtpSendMail = errors.New("prxmail.main.ErrMainSmtpSendMail")
)

func AppMain(revision string) (code int) {
	var err error
	// 設定の初期化
	config := GetConfigInstance()
	config.Revision = revision
	// ロガーの初期化
	SetupLogger()
	// フラグの解析
	if err = ParseFlags(); err != nil {
		err = errs.Wrap(ErrMainArgs, errs.WithCause(err))
		ErrorLog(err)
		return -1
	}
	// ヘルプの表示
	if config.IsHelpRequested {
		pflag.PrintDefaults()
		return 0
	}
	// バージョンの表示
	if config.IsVersionRequested {
		fmt.Printf("prxmail %s\n", config.VersionInfo())
		return 0
	}
	// 環境変数ファイルパスの取得
	envPath, err := envFilePath()
	if err != nil {
		ErrorLog(err)
		return -1
	}
	// 環境変数ファイルの読み込み
	if err = godotenv.Load(envPath); err != nil {
		err = errs.Wrap(ErrMainDotenvLoad, errs.WithCause(err))
		ErrorLog(err)
		return -1
	}
	// 環境変数の読み込み
	config.Host = os.Getenv("HOST")
	config.Port = os.Getenv("PORT")
	config.Username = os.Getenv("USERNAME")
	config.Password = os.Getenv("PASSWORD")
	// パイプの読み込み
	config.Body, err = ReadPipe()
	if err != nil {
		ErrorLog(err)
		return -1
	}
	if len(config.Body) == 0 {
		fmt.Println("パイプからメール本文を設定してください。")
		return 0
	}
	// アプリケーション情報の記録
	Logger.Info().Str("version", config.Version).
		Str("revision", config.Revision).
		Str("host", config.Host).
		Str("port", config.Port).
		Str("username", config.Username).
		Str("password", config.LogPassword()).
		Str("from", config.From).
		Strs("to", config.Recipients).
		Str("subject", config.Subject).
		Msg("start")
	// メールの組み立て
	mail, err := BuildMail()
	if err != nil {
		ErrorLog(err)
		return -1
	}
	// メールの送信
	if err := Send(mail); err != nil {
		ErrorLog(err)
		return -1
	}
	return 0
}

// パイプの読み込み
func ReadPipe() (string, error) {
	// ターミナル入力の場合は処理しない。
	isTerm := term.IsTerminal(int(os.Stdin.Fd()))
	if isTerm {
		return "", errs.Wrap(ErrMainTermStdin)
	}
	// パイプの標準入力を文字列化する。
	sb := new(strings.Builder)
	stdin := bufio.NewReader(os.Stdin)
	if _, err := io.Copy(sb, stdin); err != nil {
		err = errs.Wrap(ErrMainStdinRead, errs.WithCause(err))
		return "", err
	}
	return sb.String(), nil
}

// メールの組み立て
func BuildMail() (*Mail, error) {
	config := GetConfigInstance()
	mail := NewMail()
	mail.Subject = config.Subject
	mail.Body = config.Body
	if err := mail.SetFrom(config.From); err != nil {
		return nil, err
	}
	if err := mail.SetRecipients(config.Recipients...); err != nil {
		return nil, err
	}
	return mail, nil
}

// メールの送信
func Send(mail *Mail) (err error) {
	logMsg := "prxmail.main.Send()"
	config := GetConfigInstance()
	// メッセージの取得
	message, err := mail.Message()
	if err != nil {
		return err
	}
	fmt.Println(message)
	// メールの送信
	err = smtp.SendMail(
		config.HostWithPort(),
		NewPlainOrLoginAuth(),
		mail.From(),
		mail.Recipients(),
		[]byte(message),
	)
	if err != nil {
		return errs.Wrap(ErrMainSmtpSendMail, errs.WithCause(err),
			errs.WithContext("host", config.HostWithPort()),
			errs.WithContext("from", mail.From()),
			errs.WithContext("to", mail.Recipients()),
			errs.WithContext("message", message),
		)
	}
	Logger.Info().Str("Message", message).Msg(logMsg)
	return nil
}

// 環境変数ファイルパスの取得
func envFilePath() (string, error) {
	// 実行ファイルのパスの取得
	exePath, err := os.Executable()
	if err != nil {
		err = errs.Wrap(ErrMainExePath, errs.WithCause(err))
		return "", err
	}
	// 絶対パスの取得
	exeAbsPath, err := filepath.Abs(exePath)
	if err != nil {
		err = errs.Wrap(ErrMainExePath, errs.WithCause(err))
		return "", err
	}
	// 環境変数ファイルパスの組み立て
	envPath := filepath.Join(
		filepath.Dir(exeAbsPath), "prxmail.env")
	return envPath, nil
}
