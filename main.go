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
	// ロガーの初期化
	SetupLogger()
	// リビジョンの記録
	Logger.Info().Str("Revision", gitRevision).Msg("start")
	// 環境変数の読み込み
	err := godotenv.Load()
	if err != nil {
		err = errs.Wrap(ErrMainDotenvLoad, errs.WithCause(err))
		Logger.Error().Err(err).Msg("prxmail.AppMain()")
		return -1
	}
	// 環境変数の読み込み
	Logger.Info().
		Str("host", os.Getenv("HOST")).
		Str("port", os.Getenv("PORT")).
		Str("from", os.Getenv("FROM")).
		Str("password", os.Getenv("PASSWORD")).
		Msg("")
	return 0
}
