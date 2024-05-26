package prxmail

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	// ロガー
	AppLogger struct {
		zerolog.Logger
	}
)

const (
	// ログファイルのローテーションサイズ(MB)
	LOG_MAX_SIZE = 100
	// ログファイルの保存個数
	LOG_MAX_BACKUPS = 3
)

var (
	// ロガーのインスタンス
	Logger *AppLogger
)

func init() {
	// ロガーの初期化
	Logger = &AppLogger{
		Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}
}

// ログ出力先の設定
func SetupLogger() {
	// ログファイルの設定
	fileWriter := &lumberjack.Logger{
		Filename:   "prxmail.log",
		MaxSize:    LOG_MAX_SIZE,
		MaxBackups: LOG_MAX_BACKUPS,
	}
	writer := io.MultiWriter(fileWriter, os.Stdout)
	// ロガーの出力先の変更
	Logger = &AppLogger{
		Logger: Logger.Output(writer),
	}
}
