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
	Logger.Info().
		Str("host", os.Getenv("HOST")).
		Str("port", os.Getenv("PORT")).
		Str("username", os.Getenv("USERNAME")).
		Str("password", os.Getenv("PASSWORD")).
		Msg("")
	// メッセージの組み立て
	builder := NewMessageBuilder()
	if err := builder.SetFromAddr("admin@example.com"); err != nil {
		Logger.Error().Err(err).Msg(logMsg)
		return -1
	}
	if err := builder.SetToAddrs("alice@foo.bar", "bob@foo.bar"); err != nil {
		Logger.Error().Err(err).Msg(logMsg)
		return -1
	}
	builder.SetSubject("Test Subject")
	builder.SetBody("Body Body Body\nBody Body Body")
	message, err := builder.Build()
	if err != nil {
		Logger.Error().Err(err).Msg(logMsg)
		return -1
	}
	Logger.Info().Str("Message", message).Msg(logMsg)
	return 0
}
