package prxmail

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/goark/errs"
)

type (
	MessageBuilder struct {
		fromAddr *mail.Address
		toAddrs  [](*mail.Address)
		subject  string
		body     string
	}
)

var (
	// 不正なメールアドレス形式
	ErrMessageAddressInvalid = errors.New("message.ErrMessageAddressInvalid")
	// 送信元が空
	ErrMessageFromEmpty = errors.New("message.ErrMessageFromEmpty")
	// 送信先が空
	ErrMessageToEmpty = errors.New("message.ErrMessageToEmpty")
	// 本文が空
	ErrMessageBodyEmpty = errors.New("message.ErrMessageBodyEmpty")
)

func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		toAddrs: make([](*mail.Address), 0, 1),
	}
}

func (b *MessageBuilder) SetFromAddr(addr string) error {
	from, err := mail.ParseAddress(addr)
	if err != nil {
		return errs.Wrap(ErrMessageAddressInvalid, errs.WithCause(err))
	}
	b.fromAddr = from
	return nil
}

func (b *MessageBuilder) SetToAddrs(addrs ...string) error {
	for _, addr := range addrs {
		to, err := mail.ParseAddress(addr)
		if err != nil {
			return errs.Wrap(ErrMessageAddressInvalid, errs.WithCause(err))
		}
		b.toAddrs = append(b.toAddrs, to)
	}
	return nil
}

func (b *MessageBuilder) SetSubject(subject string) {
	b.subject = subject
}

func (b *MessageBuilder) SetBody(body string) {
	b.body = body
}

func (b *MessageBuilder) Build() (string, error) {
	// エラーの確認
	if b.fromAddr == nil {
		return "", ErrMessageFromEmpty
	}
	if len(b.toAddrs) == 0 {
		return "", ErrMessageToEmpty
	}
	if len(b.body) == 0 {
		return "", ErrMessageBodyEmpty
	}
	// 送信元の組み立て
	var sb strings.Builder
	sb.WriteString("From: ")
	sb.WriteString(b.fromAddr.String())
	sb.WriteString("\r\n")
	// 送信先の組み立て
	sb.WriteString("To: ")
	toStrs := make([]string, len(b.toAddrs))
	for i, toAddr := range b.toAddrs {
		toStrs[i] = toAddr.String()
	}
	sb.WriteString(strings.Join(toStrs, ", "))
	sb.WriteString("\r\n")
	// 件名の組み立て
	sb.WriteString("Subject: ")
	if len(b.subject) == 0 {
		sb.WriteString("empty")
	} else {
		sb.WriteString(b.subject)
	}
	sb.WriteString("\r\n\r\n")
	// 本文の組み立て
	sb.WriteString(b.body)
	// メッセージを文字列として出力する。
	return sb.String(), nil
}
