package prxmail

import (
	"errors"
	"os"

	"github.com/goark/errs"
	"github.com/joho/godotenv"
)

var (
	// envファイルロードエラー
	ErrMainDotenvLoad = errors.New("main.ErrMainDotendLoad")
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
	message, err := mail.Message()
	if err != nil {
		Logger.Error().Err(err).Msg(logMsg)
		return -1
	}
	// メールの送信
	Logger.Info().Str("Message", message).Msg(logMsg)
	return 0
}
