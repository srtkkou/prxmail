/*
Copyright © 2024 srtkkou <srtkkou@outlook.jp>
*/
package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/goark/errs"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ロガー
type AppLogger struct {
	base  slog.Logger
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

// DEBUGログの出力
func DebugLog(msg string, args ...any) {
	printLog(slog.LevelDebug, msg, args...)
}

// INFOログの出力
func InfoLog(msg string, args ...any) {
	printLog(slog.LevelInfo, msg, args...)
}

// ERRORログの出力
func ErrorLog(msg string, err error) {
	str := fmt.Sprintf("%+v", err)
	jsonBlob := json.RawMessage([]byte(str))
	fmt.Fprintln(os.Stderr, str)
	printLog(slog.LevelError, msg, slog.Any("error", jsonBlob))
}

// ロガーの取得
func logger() *AppLogger {
	// ロガーの初期化
	logOnce.Do(func() {
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
		writer := &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    LOG_MAX_SIZE,
			MaxBackups: LOG_MAX_BACKUPS,
		}
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
		appLogger.base = *slogLogger
		appLoggerPtr = &appLogger
	})
	return appLoggerPtr
}

// ログの出力
func printLog(level slog.Level, msg string, args ...any) {
	ctx := Config().Context()
	logger().base.Log(ctx, level, msg, args...)
}
