package prxmail

import (
	"crypto/tls"
	"errors"
	"io"
	"net/smtp"
	"os"

	"github.com/goark/errs"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

var (
	// バージョン
	Version = "v0.0.1"
	// envファイルロードエラー
	ErrMainDotenvLoad = errors.New("main.ErrMainDotendLoad")
	// 引数解析エラー
	ErrMainArgs = errors.New("main.ErrMainArgs")
	// TLSダイヤルエラー
	ErrMainTlsDial = errors.New("main.ErrMainTlsDial")
	// SMTPクライアント初期化エラー
	ErrMainSmtpNewClient = errors.New("main.ErrMainSmtpNewClient")
	// SMTPエラー
	ErrMainSmtpAuth = errors.New("main.ErrMainSmtpAuth")
	// SMTP MAILコマンドエラー
	ErrMainSmtpMail = errors.New("main.ErrMainSmtpMail")
	// SMTP RCPTコマンドエラー
	ErrMainSmtpRcpt = errors.New("main.ErrMainSmtpRcpt")
	// SMTP DATAコマンドエラー
	ErrMainSmtpData = errors.New("main.ErrMainSmtpData")
	// SMTP 書き込みエラー
	ErrMainSmtpWrite = errors.New("main.ErrMainSmtpWrite")
)

func AppMain(args []string, revision string) (code int) {
	logMsg := "prxmail.main.AppMain()"
	// 設定の初期化
	config := GetConfig()
	config.Revision = revision
	// ロガーの初期化
	SetupLogger()
	// 環境変数の読み込み
	err := godotenv.Load()
	if err != nil {
		err = errs.Wrap(ErrMainDotenvLoad, errs.WithCause(err))
		Logger.Error().Err(err).Msg(logMsg)
		return -1
	}
	// 引数の解析
	app := &cli.App{
		Version: config.VersionInfo(),
		Action: func(cCtx *cli.Context) error {
			return CliAction()
		},
	}
	if err = app.Run(args); err != nil {
		err = errs.Wrap(ErrMainArgs, errs.WithCause(err))
		Logger.Error().Err(err).Msg(logMsg)
		return -1
	}
	return 0
}

func CliAction() error {
	config := GetConfig()
	// リビジョンの記録
	Logger.Info().Str("Version", config.VersionInfo()).Msg("start")
	// 環境変数の読み込み
	config.Host = os.Getenv("HOST")
	config.Port = os.Getenv("PORT")
	config.Username = os.Getenv("USERNAME")
	config.Password = os.Getenv("PASSWORD")
	Logger.Info().
		Str("host", config.Host).
		Str("port", config.Port).
		Str("username", config.Username).
		Str("password", config.Password).
		Msg("")
	// メッセージの組み立て
	mail := NewMail()
	mail.Subject = "Test Subject"
	mail.Body = "Body Body Body\nBody Body Body"
	if err := mail.SetFrom("admin@example.com"); err != nil {
		return err
	}
	if err := mail.SetRecipients("alice@foo.bar", "bob@foo.bar"); err != nil {
		return err
	}
	// メールの送信
	if err := Send(mail); err != nil {
		return err
	}
	return nil
}

func Send(mail *Mail) error {
	logMsg := "prxmail.main.Send()"
	// 設定の取得
	config := GetConfig()
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
