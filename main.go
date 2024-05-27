package prxmail

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"strings"

	"github.com/goark/errs"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

var (
	// バージョン
	Version = "v0.0.1"
	// envファイルロードエラー
	ErrMainDotenvLoad = errors.New("prxmail.main.ErrMainDotenvLoad")
	// 引数解析エラー
	ErrMainArgs = errors.New("prxmail.main.ErrMainArgs")
	// ターミナルの標準入力は受け付けない
	ErrMainTermStdin = errors.New("prxmail.main.ErrMainTermStdin")
	// 標準入力読み込みエラー
	ErrMainStdinRead = errors.New("prxmail.main.ErrMainStdinRead")
	// TLSダイヤルエラー
	ErrMainTlsDial = errors.New("prxmail.main.ErrMainTlsDial")
	// SMTPクライアント初期化エラー
	ErrMainSmtpNewClient = errors.New("prxmail.main.ErrMainSmtpNewClient")
	// SMTPエラー
	ErrMainSmtpAuth = errors.New("prxmail.main.ErrMainSmtpAuth")
	// SMTP MAILコマンドエラー
	ErrMainSmtpMail = errors.New("prxmail.main.ErrMainSmtpMail")
	// SMTP RCPTコマンドエラー
	ErrMainSmtpRcpt = errors.New("prxmail.main.ErrMainSmtpRcpt")
	// SMTP DATAコマンドエラー
	ErrMainSmtpData = errors.New("prxmail.main.ErrMainSmtpData")
	// SMTP 書き込みエラー
	ErrMainSmtpWrite = errors.New("prxmail.main.ErrMainSmtpWrite")
)

func AppMain(args []string, revision string) (code int) {
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
	// 環境変数ファイルの読み込み
	if err = godotenv.Load(); err != nil {
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
func Send(mail *Mail) error {
	logMsg := "prxmail.main.Send()"
	config := GetConfigInstance()
	// TLS認証準備
	tlsServer := config.TlsServer()
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         config.Host,
	}
	// TLS通信開始
	tlsConn, err := tls.Dial("tcp", tlsServer, tlsConf)
	if err != nil {
		return errs.Wrap(ErrMainTlsDial, errs.WithCause(err),
			errs.WithContext("server", tlsServer))
	}
	defer tlsConn.Close()
	// SMTP通信開始
	client, err := smtp.NewClient(tlsConn, config.Host)
	if err != nil {
		return errs.Wrap(ErrMainSmtpNewClient, errs.WithCause(err),
			errs.WithContext("host", config.Host))
	}
	defer client.Quit()
	// 認証の実行
	auth := smtp.PlainAuth(
		"", config.Username, config.Password, config.Host,
	)
	if err = client.Auth(auth); err != nil {
		return errs.Wrap(ErrMainSmtpAuth, errs.WithCause(err),
			errs.WithContext("username", config.Username),
			errs.WithContext("password", config.Password),
			errs.WithContext("host", config.Host),
		)
	}
	// MAILコマンドの実行
	if err = client.Mail(mail.From().String()); err != nil {
		return errs.Wrap(ErrMainSmtpMail, errs.WithCause(err),
			errs.WithContext("from", mail.From().String()))
	}
	// RCPTコマンドの実行
	for _, recipient := range mail.Recipients() {
		if err = client.Rcpt(recipient.String()); err != nil {
			return errs.Wrap(ErrMainSmtpRcpt, errs.WithCause(err),
				errs.WithContext("recipient", recipient.String()))
		}
	}
	// DATAコマンドの実行
	w, err := client.Data()
	if err != nil {
		return errs.Wrap(ErrMainSmtpData, errs.WithCause(err))
	}
	defer w.Close()
	writer := io.MultiWriter(w, os.Stdout)
	// メッセージの取得
	message, err := mail.Message()
	if err != nil {
		return err
	}
	// メッセージの書き込み
	if _, err = writer.Write([]byte(message)); err != nil {
		return errs.Wrap(ErrMainSmtpWrite, errs.WithCause(err),
			errs.WithContext("message", message))
	}
	Logger.Info().Str("Message", message).Msg(logMsg)
	return nil
}
