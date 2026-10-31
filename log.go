package prxmail

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/goark/errs"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ロガー
type AppLogger struct {
	slog.Logger
	level *slog.LevelVar // ログレベル
}

const (
	// ログファイルのローテーションサイズ(MB)
	LOG_MAX_SIZE = 100
	// ログファイルの保存個数
	LOG_MAX_BACKUPS = 3
)

var (
	// ロガー初期化排他
	logOnce sync.Once
	// ロガーへのポインタ
	appLoggerPtr *AppLogger
)

var (
	ErrLoggerFileOpen  = errors.New("lib.ErrLoggerFileOpen")
	ErrLoggerFileClose = errors.New("lib.ErrLoggerFileClose")
)

func Log() *AppLogger {
	// ロガーの初期化
	logOnce.Do(func() {
		var writer io.Writer
		config := Config()
		// ログファイルの新規作成
		logPath := config.LogPath()
		f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			err = errs.Wrap(ErrLoggerFileOpen, errs.WithCause(err),
				errs.WithContext("path", logPath))
			log.Fatalf("%+v\n", err)
		}
		if err = f.Close(); err != nil {
			err = errs.Wrap(ErrLoggerFileClose, errs.WithCause(err),
				errs.WithContext("path", logPath))
			log.Fatalf("%+v\n", err)
		}
		// ログファイルのローテーション設定
		fileWriter := &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    LOG_MAX_SIZE,
			MaxBackups: LOG_MAX_BACKUPS,
		}
		writer = io.MultiWriter(fileWriter, os.Stdout)
		// インスタンスの初期化
		appLogger := AppLogger{
			level: new(slog.LevelVar),
		}
		// レベルの設定
		appLogger.level.Set(slog.LevelInfo)
		// slogの初期化
		slogHandler := slog.NewJSONHandler(writer,
			&slog.HandlerOptions{Level: appLogger.level})
		slogLogger := slog.New(slogHandler)
		appLogger.Logger = *slogLogger
		appLoggerPtr = &appLogger
	})
	return appLoggerPtr
}

// エラーログの出力
func (l *AppLogger) Errorf(msg string, err error) {
	str := fmt.Sprintf("%+v", err)
	rawJSON := json.RawMessage([]byte(str))
	l.Error(msg, slog.Any("error", rawJSON))
}
