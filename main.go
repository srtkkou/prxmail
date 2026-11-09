/*
Copyright © 2024 srtkkou <srtkkou@outlook.jp>
*/
package prxmail

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/goark/errs"
	"golang.org/x/term"

	"github.com/srtkkou/prxmail/lib"
)

var (
	// ターミナルの標準入力は受け付けない
	ErrMainTermStdin = errors.New("main.ErrMainTermStdin")
	// 標準入力読み込みエラー
	ErrMainStdinRead = errors.New("main.ErrMainStdinRead")
	// 本文が空
	ErrMainBodyEmpty = errors.New("main.ErrMainBodyEmpty")
)

// 処理の実行
func AppMain() (code int) {
	// 設定ファイルの読み込み
	config := lib.Config()
	err := config.LoadConfig()
	if err != nil {
		lib.Log().Errorf("config.LoadConfig()", err)
		return -1
	}
	// パイプの読み込み
	body, err := readPipe()
	if err != nil {
		fmt.Println("パイプからメール本文を設定してください。")
		lib.Log().Errorf("main.readPipe()", err)
		return -1
	}
	config.SetBody(body)
	// アプリケーション情報の記録
	configBlob, _ := config.MarshalJSON()
	jsonBlob := json.RawMessage(configBlob)
	lib.Log().Info("main.AppMain()", slog.Any("config", jsonBlob))
	// メールの組み立て
	mail, err := buildMail()
	if err != nil {
		lib.Log().Errorf("main.buildMail", err)
		return -1
	}
	// メールの送信
	if err := lib.SendMail(mail); err != nil {
		lib.Log().Errorf("sender.sendMail", err)
		return -1
	}
	return 0
}

// パイプの読み込み
func readPipe() (string, error) {
	var err error
	// ターミナル入力の場合は処理しない。
	isTerm := term.IsTerminal(int(os.Stdin.Fd()))
	if isTerm {
		return "", errs.Wrap(ErrMainTermStdin)
	}
	// パイプの標準入力を文字列化する。
	sb := new(strings.Builder)
	stdin := bufio.NewReader(os.Stdin)
	if _, err = io.Copy(sb, stdin); err != nil {
		err = errs.Wrap(ErrMainStdinRead, errs.WithCause(err))
		return "", err
	}
	// 文字列が空か確認する。
	if sb.Len() == 0 {
		err = errs.Wrap(ErrMainBodyEmpty)
		return "", err
	}
	return sb.String(), err
}

// メールの組み立て
func buildMail() (*lib.Mail, error) {
	config := lib.Config()
	mail := lib.NewMail()
	mail.Subject = config.Subject()
	mail.Body = config.Body()
	if err := mail.SetFrom(config.From()); err != nil {
		return nil, err
	}
	addrs := config.Recipients()
	if err := mail.SetRecipients(addrs...); err != nil {
		return nil, err
	}
	return mail, nil
}
