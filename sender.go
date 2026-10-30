package prxmail

import (
	"errors"
	"fmt"
	"net/smtp"

	"github.com/goark/errs"
)

var (
	// SMTPメール送信エラー
	ErrSMTPSendMail = errors.New("sender.ErrSMTPSendMail")
)

// メールの送信
func sendMail(mail *Mail) (err error) {
	logMsg := "prxmail.sendMail()"
	config := Config()
	// メッセージの取得
	message, err := mail.Message()
	if err != nil {
		return err
	}
	fmt.Println(message)
	// メールの送信
	err = smtp.SendMail(
		config.SMTPHostWithPort(),
		NewPlainOrLoginAuth(),
		mail.From(),       // config.From()
		mail.Recipients(), // config.Recipients()
		[]byte(message),
	)
	if err != nil {
		return errs.Wrap(ErrMainSmtpSendMail, errs.WithCause(err),
			errs.WithContext("host", config.SMTPHostWithPort()),
			errs.WithContext("from", mail.From()),
			errs.WithContext("to", mail.Recipients()),
			errs.WithContext("message", message),
		)
	}
	Logger.Info().Str("Message", message).Msg(logMsg)
	return nil
}
