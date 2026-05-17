package main

import (
	"bytes"
	_ "embed"
	"html/template"
	"time"

	gomail "github.com/go-mail/mail/v2"
)

//go:embed templates/verify_email.tmpl
var verifyEmailTmpl string

func (app *application) sendVerificationEmail(to, code string) error {
	tmpl, err := template.New("verify").Parse(verifyEmailTmpl)
	if err != nil {
		return err
	}
	data := map[string]any{"Code": code, "ExpiresIn": 10}

	subj, plain, html := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	_ = tmpl.ExecuteTemplate(subj, "subject", data)
	_ = tmpl.ExecuteTemplate(plain, "plainBody", data)
	_ = tmpl.ExecuteTemplate(html, "htmlBody", data)

	msg := gomail.NewMessage()
	msg.SetHeader("From", app.cfg.smtp.sender)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subj.String())
	msg.SetBody("text/plain", plain.String())
	msg.AddAlternative("text/html", html.String())

	d := gomail.NewDialer(app.cfg.smtp.host, app.cfg.smtp.port, app.cfg.smtp.username, app.cfg.smtp.password)
	d.Timeout = 5 * time.Second
	return d.DialAndSend(msg)
}
