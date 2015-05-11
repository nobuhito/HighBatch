package highbatch

import (
	"net/smtp"
	"fmt"
	"github.com/kr/pretty"
	"strings"
)

func sendSmtpAuth(n NotifyConfig, subject, body string) error{
	auth := smtp.PlainAuth(
		"",
		n.SmtpAuth.User,
		n.SmtpAuth.Pass,
		n.MailInfo.Host,
	)

	err := smtp.SendMail(
		n.MailInfo.Host+":"+n.MailInfo.Port,
		auth,
		n.MailInfo.FromAddress,
		n.MailInfo.ToAddress,
		[]byte("test"),
	)

	if err != nil {
		return err
	}

	return nil
}

func sendSmtp(n NotifyConfig, subject, body string) error {
	c, err := smtp.Dial(n.MailInfo.Host+":"+n.MailInfo.Port)
	if err != nil {
		return err
	}
	defer c.Quit()

	if err := c.Mail(n.MailInfo.FromAddress); err != nil {
		return err
	}
	for i := range n.MailInfo.ToAddress {
		if err := c.Rcpt(n.MailInfo.ToAddress[i]); err != nil {
			return err
		}
	}

	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	if _, err := fmt.Fprintf(wc, "Subject:"+subject+"\r\n\r\n"+strings.TrimSpace(body)); err != nil {
		return err
	}

	return nil
}

func notify(spec Spec) {
	subject := "[HighBatch] Notify "+spec.Hostname+" / "+spec.Name
	body := spec.Output+"\r\n\r\n"+pretty.Sprintf("%# v", spec)

	n := Conf.Notify

	if n.SmtpAuth.User != "" && n.SmtpAuth.Pass != "" {
		if err := sendSmtpAuth(n, subject, body); err != nil {
			le(err)
		}
	} else if n.MailInfo.Host != "" && n.MailInfo.Port != "" {
		if err := sendSmtp(n, subject, body); err != nil {
			le(err)
		}
	}
}
