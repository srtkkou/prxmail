package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	app "github.com/srtkkou/prxmail"
)

// コマンド
type appCmd struct {
	fs         *pflag.FlagSet // フラグ
	isHelp     bool           // ヘルプ表示
	from       string         // 送信元
	recipients []string       // 送信先
	subject    string         // 件名
}

const (
	// コマンド名
	COMMAND = "prxmail"
	// バージョン
	VERSION = "0.1.0"
)

var (
	// コミットハッシュ
	CommitHash string
)

func main() {
	// 設定の取得
	config := app.Config()
	// 設定の更新
	config.SetVersion(VERSION)
	config.SetCommitHash(CommitHash)
	// フラグの設定
	cmd := &appCmd{
		fs: pflag.NewFlagSet("app", pflag.ExitOnError),
	}
	cmd.fs.BoolVarP(
		&cmd.isHelp, "help", "h", false,
		`Show help message.`)
	cmd.fs.StringVarP(
		&cmd.from, "return-address", "r", "",
		"Use `ADDRESS` as the return address when sending mail.")
	cmd.fs.StringSliceVarP(
		&cmd.recipients, "to", "t", []string{},
		"Add recipient `ADDRESS` to send mail.")
	cmd.fs.StringVarP(
		&cmd.subject, "subject", "s", "empty",
		"Send a message with the given `SUBJECT`.",
	)
	// フラグの解析
	cmd.fs.Parse(os.Args)
	// ヘルプの場合はメッセージを出して終了する。
	if cmd.isHelp {
		cmd.usage()
		os.Exit(0)
	}
	// フラグ値で設定を更新する。
	config.SetFrom(cmd.from)
	config.SetRecipients(cmd.recipients)
	config.SetSubject(cmd.subject)
	// 処理の実行
	os.Exit(app.AppMain())
}

// 使用方法の表示
func (c *appCmd) usage() {
	config := app.Config()
	fmt.Printf("usage: %s <options>\n", COMMAND)
	fmt.Println()
	fmt.Printf("version: %s (build:%s)\n",
		config.Version(), config.CommitHash())
	fmt.Println()
	fmt.Println(`Use pipe to set mail body.`)
	fmt.Println()
	c.fs.PrintDefaults()
}
