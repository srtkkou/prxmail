package prxmail

import (
	"bytes"
	"errors"
	"net/smtp"
	"slices"

	"github.com/goark/errs"
)

type (
	// PLAIN/LOGIN認証
	// * Go言語のnet/smtpでLOGIN認証が無いため自作。
	plainOrLoginAuth struct {
		username string
		password string
		host     string
		port     string
		method   string
	}
)

var (
	// ホスト名が一致していない。
	ErrAuthHostMatch = errors.New("prxmail.auth.ErrAuthHostMatch")
	// PLAIN認証で認証情報を送信済みなのにさらに情報を求められた。
	ErrAuthPlainChallenge = errors.New("prxmail.auth.ErrAuthPlainChallenge")
	// LOGIN認証のサーバからの要求が不正
	ErrAuthLoginChallenge = errors.New("prxmail.auth.ErrAuthLoginChallenge")
)

// PLAIN/LOGIN認証の初期化
func NewPlainOrLoginAuth() smtp.Auth {
	config := GetConfigInstance()
	return &plainOrLoginAuth{
		username: config.Username,
		password: config.Password,
		host:     config.Host,
		port:     config.Port,
	}
}

// 認証の開始
func (a *plainOrLoginAuth) Start(
	server *smtp.ServerInfo,
) (proto string, toServer []byte, err error) {
	// ホスト名が一致していることを確認する。
	if server.Name != a.host {
		err = errs.Wrap(ErrAuthHostMatch,
			errs.WithContext("server", server.Name),
			errs.WithContext("host", a.host))
		return "", nil, err
	}
	// サーバの認証方式がPLAIN認証かを確認する。
	if slices.Contains(server.Auth, "PLAIN") {
		a.method = "PLAIN"
		resp := []byte("\x00" + a.username + "\x00" + a.password)
		return a.method, resp, nil
	}
	// LOGIN認証を返す。
	a.method = "LOGIN"
	return a.method, []byte{}, nil
}

// 認証の継続
func (a *plainOrLoginAuth) Next(
	fromServer []byte, more bool,
) (toServer []byte, err error) {
	if !more {
		return nil, nil
	}
	// PLAIN認証の検証
	if a.method == "PLAIN" {
		// すでに認証情報を送信済みなのでエラーを送る。
		err = errs.Wrap(ErrAuthPlainChallenge,
			errs.WithContext("method", a.method),
			errs.WithContext("more", more))
		return nil, err
	}
	// LOGIN認証の検証
	switch {
	case bytes.Equal(fromServer, []byte("Username:")):
		return []byte(a.username), nil
	case bytes.Equal(fromServer, []byte("Password:")):
		return []byte(a.password), nil
	default:
		err = errs.Wrap(ErrAuthLoginChallenge,
			errs.WithContext("method", a.method),
			errs.WithContext("more", more),
			errs.WithContext("fromServer", fromServer))
		return nil, err
	}
}
