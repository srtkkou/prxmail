package prxmail

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/goark/errs"
	"golang.org/x/term"
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
	logMsg := "AppMain()"
	// 設定ファイルの読み込み
	config := Config()
	err := config.LoadConfig()
	if err != nil {
		Log().Errorf(err)
		return -1
	}
	// パイプの読み込み
	body, err := readPipe()
	if err != nil {
		fmt.Println("パイプからメール本文を設定してください。")
		Log().Errorf(err)
		return -1
	}
	config.SetBody(body)
	// アプリケーション情報の記録
	Log().Info(logMsg, slog.String("config", config.String()))
	// メールの組み立て
	mail, err := buildMail()
	if err != nil {
		Log().Errorf(err)
		return -1
	}
	// メールの送信
	if err := sendMail(mail); err != nil {
		Log().Errorf(err)
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
func buildMail() (*Mail, error) {
	config := Config()
	mail := NewMail()
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
