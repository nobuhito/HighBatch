package highbatch

import (
	"bytes"
	"encoding/base64"
	"github.com/kr/pretty"
	"net/smtp"
	"strings"
	"net/http"
	"encoding/json"
)

// http://qiita.com/yamasaki-masahide/items/a9f8b43eeeaddbfb6b44
// 76バイト毎にCRLFを挿入する
func add76crlf(msg string) string {
	var buffer bytes.Buffer
	for k, c := range strings.Split(msg, "") {
		buffer.WriteString(c)
		if k%76 == 75 {
			buffer.WriteString("\r\n")
		}
	}
	return buffer.String()
}

// UTF8文字列を指定文字数で分割
func utf8Split(utf8string string, length int) []string {
	resultString := []string{}
	var buffer bytes.Buffer
	for k, c := range strings.Split(utf8string, "") {
		buffer.WriteString(c)
		if k%length == length-1 {
			resultString = append(resultString, buffer.String())
			buffer.Reset()
		}
	}
	if buffer.Len() > 0 {
		resultString = append(resultString, buffer.String())
	}
	return resultString
}

// サブジェクトをMIMEエンコードする
func encodeSubject(subject string) string {
	var buffer bytes.Buffer
	buffer.WriteString("Subject:")
	for _, line := range utf8Split(subject, 13) {
		buffer.WriteString(" =?utf-8?B?")
		buffer.WriteString(base64.StdEncoding.EncodeToString([]byte(line)))
		buffer.WriteString("?=\r\n")
	}
	return buffer.String()
}

func sendSmtp(n NotifyConfig, subject, body string) error {

	auth := func(n NotifyConfig) smtp.Auth {
		if n.SmtpAuth.User != "" && n.SmtpAuth.Pass != "" {
			auth := smtp.PlainAuth(
				"",
				n.SmtpAuth.User,
				n.SmtpAuth.Pass,
				n.MailInfo.Host,
			)
			return auth
		}

		return nil
	}

	var mail bytes.Buffer
	mail.WriteString("From: " + n.MailInfo.FromAddress + "\r\n")
	mail.WriteString("To: " + n.MailInfo.ToAddress[0] + "\r\n")
	for i := range n.MailInfo.ToAddress {
		mail.WriteString("Cc: " + n.MailInfo.ToAddress[i] + "\r\n")
	}
	mail.WriteString(encodeSubject(subject))
	mail.WriteString("MIME-Version: 1.0\r\n")
	mail.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	mail.WriteString("Content-Transfer-Encoding: base64\r\n")
	mail.WriteString("X-Sender: HighBatch\r\n")
	mail.WriteString("\r\n")
	mail.WriteString(add76crlf(base64.StdEncoding.EncodeToString([]byte(body))))

	err := smtp.SendMail(
		n.MailInfo.Host+":"+n.MailInfo.Port,
		auth(n),
		n.MailInfo.FromAddress,
		n.MailInfo.ToAddress,
		[]byte(mail.String()),
	)
	if err != nil {
		return err
	}

	return nil
}

func webhook (n NotifyConfig, spec Spec) {
	json, err := json.Marshal(spec)
	if err != nil {
		le(err)
	}
	_, err = http.Post(n.WebhookInfo.Url + "?room=" + n.WebhookInfo.Room, "application/json", bytes.NewBuffer(json))
	if err != nil {
		le(err)
	}
}

func notify(spec Spec) {
	n := Conf.Notify
	webhook(n, spec)

	if n.MailInfo.FromAddress != "" && len(n.MailInfo.ToAddress) > 0 {
		subject := "[HighBatch] Notify " + spec.Hostname + " / " + spec.Name
		body := spec.Output + "\r\n\r\n" + pretty.Sprintf("%# v", spec)

		if err := sendSmtp(n, subject, body); err != nil {
			le(err)
		}
	}

}
