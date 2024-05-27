package prxmail

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/goark/errs"
)

type (
	// メール
	Mail struct {
		Subject    string            // 件名
		Body       string            // 本文
		from       *mail.Address     // 送信元
		recipients [](*mail.Address) // 送信先
	}
)

var (
	// 不正なメールアドレス形式
	ErrMailAddressInvalid = errors.New("message.ErrMailAddressInvalid")
	// 送信元が空
	ErrMailFromEmpty = errors.New("message.ErrMailFromEmpty")
	// 送信先が空
	ErrMailRecipientsEmpty = errors.New("message.ErrMailRecipientsEmpty")
	// 本文が空
	ErrMailBodyEmpty = errors.New("message.ErrMailBodyEmpty")
)

// 新規メールの作成
func NewMail() *Mail {
	return &Mail{
		recipients: make([](*mail.Address), 0, 1),
	}
}

// 送信元の取得
func (m *Mail) From() *mail.Address {
	return m.from
}

// 送信元の設定
func (m *Mail) SetFrom(addr string) error {
	from, err := mail.ParseAddress(addr)
	if err != nil {
		return errs.Wrap(ErrMailAddressInvalid, errs.WithCause(err))
	}
	m.from = from
	return nil
}

// 送信先の取得
func (m *Mail) Recipients() [](*mail.Address) {
	return m.recipients
}

// 送信先の設定
func (m *Mail) SetRecipients(addrs ...string) error {
	for _, addr := range addrs {
		to, err := mail.ParseAddress(addr)
		if err != nil {
			return errs.Wrap(ErrMailAddressInvalid, errs.WithCause(err))
		}
		m.recipients = append(m.recipients, to)
	}
	return nil
}

// メッセージの作成
func (m *Mail) Message() (string, error) {
	// エラーの確認
	if m.from == nil {
		return "", ErrMailFromEmpty
	}
	if len(m.recipients) == 0 {
		return "", ErrMailRecipientsEmpty
	}
	if len(m.Body) == 0 {
		return "", ErrMailBodyEmpty
	}
	// 送信元の組み立て
	var sb strings.Builder
	sb.WriteString("From: ")
	sb.WriteString(m.from.String())
	sb.WriteString("\r\n")
	// 送信先の組み立て
	sb.WriteString("To: ")
	for i, recipient := range m.recipients {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(recipient.String())
	}
	sb.WriteString("\r\n")
	// 件名の組み立て
	sb.WriteString("Subject: ")
	if len(m.Subject) == 0 {
		sb.WriteString("empty")
	} else {
		sb.WriteString(m.Subject)
	}
	sb.WriteString("\r\n\r\n")
	// 本文の組み立て
	sb.WriteString(m.Body)
	// メッセージを文字列として出力する。
	return sb.String(), nil
}
