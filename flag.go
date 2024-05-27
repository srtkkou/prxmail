package prxmail

import (
	"errors"
	"fmt"
	"os"

	"github.com/goark/errs"
	"github.com/spf13/pflag"
)

var (
	// ホスト名取得エラー
	ErrFlagHostname = errors.New("prxmail.flag.ErrFlagHostname")
)

// フラグの解析
func ParseFlags() error {
	config := GetConfigInstance()
	// 実行環境のホスト名の取得
	currentHost, err := os.Hostname()
	if err != nil {
		err = errs.Wrap(ErrFlagHostname, errs.WithCause(err))
		return err
	}
	// 送信元の設定
	defaultFrom := fmt.Sprintf("prxmail@%s", currentHost)
	pflag.StringVarP(
		&(config.From), "return-address", "r", defaultFrom,
		"use `ADDRESS` as the return address when sending mail",
	)
	// 送信先の設定
	pflag.StringSliceVarP(
		&(config.Recipients), "to", "t", []string{},
		"add recipient `ADDRESS` to send mail",
	)
	// 件名の設定
	pflag.StringVarP(
		&(config.Subject), "subject", "s", "empty",
		"send a message with the given `SUBJECT`",
	)
	// ヘルプ表示フラグの設定
	pflag.BoolVarP(
		&(config.IsHelpRequested), "help", "h", false,
		"give this help list",
	)
	// バージョン表示フラグの設定
	pflag.BoolVarP(
		&(config.IsVersionRequested), "version", "V", false,
		"print program version",
	)
	// フラグの解析
	pflag.Parse()
	return nil
}
