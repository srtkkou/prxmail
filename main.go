package prxmail

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/smtp"
	"os"

	"github.com/goark/errs"
	"github.com/joho/godotenv"
)

var (
	// envファイルロードエラー
	ErrMainDotenvLoad = errors.New("main.ErrMainDotendLoad")
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

func AppMain(gitRevision string) (code int) {
	logMsg := "prxmail.main.AppMain()"
	// ロガーの初期化
	SetupLogger()
	// リビジョンの記録
	Logger.Info().Str("Revision", gitRevision).Msg("start")
	// 環境変数の読み込み
	err := godotenv.Load()
	if err != nil {
		err = errs.Wrap(ErrMainDotenvLoad, errs.WithCause(err))
		Logger.Error().Err(err).Msg(logMsg)
		return -1
	}
	// 環境変数の読み込み
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	Logger.Info().
		Str("host", host).Str("port", port).
		Str("username", username).
		Str("password", password).
		Msg("")
	// メッセージの組み立て
	mail := NewMail()
	mail.Subject = "Test Subject"
	mail.Body = "Body Body Body\nBody Body Body"
	if err := mail.SetFrom("admin@example.com"); err != nil {
		Logger.Error().Err(err).Msg(logMsg)
		return -1
	}
	if err := mail.SetRecipients("alice@foo.bar", "bob@foo.bar"); err != nil {
		Logger.Error().Err(err).Msg(logMsg)
		return -1
	}
	// メールの送信
	if err := Send(host, port, username, password, mail); err != nil {
		Logger.Error().Err(err).Msg(logMsg)
		return -1
	}
	return 0
}

func Send(
	host string, port string,
	username string, password string,
	mail *Mail,
) error {
	logMsg := "prxmail.main.Send()"
	// TLS認証準備
	tlsServer := fmt.Sprintf("%s:%s", host, port)
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	// TLS通信開始
	tlsConn, err := tls.Dial("tcp", tlsServer, tlsConf)
	if err != nil {
		return errs.Wrap(ErrMainTlsDial, errs.WithCause(err),
			errs.WithContext("server", tlsServer))
	}
	defer tlsConn.Close()
	// SMTP通信開始
	client, err := smtp.NewClient(tlsConn, host)
	if err != nil {
		return errs.Wrap(ErrMainSmtpNewClient, errs.WithCause(err))
	}
	defer client.Quit()
	// 認証の実行
	auth := smtp.PlainAuth("", username, password, host)
	if err = client.Auth(auth); err != nil {
		return errs.Wrap(ErrMainSmtpAuth, errs.WithCause(err))
	}
	// MAILコマンドの実行
	if err = client.Mail(mail.From().String()); err != nil {
		return errs.Wrap(ErrMainSmtpMail, errs.WithCause(err))
	}
	// RCPTコマンドの実行
	for _, recipient := range mail.Recipients() {
		if err = client.Rcpt(recipient.String()); err != nil {
			return errs.Wrap(ErrMainSmtpRcpt, errs.WithCause(err))
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
	Logger.Info().Str("Message", message).Msg(logMsg)
	// メッセージの書き込み
	if _, err = writer.Write([]byte(message)); err != nil {
		return errs.Wrap(ErrMainSmtpWrite, errs.WithCause(err))
	}
	return nil
}
