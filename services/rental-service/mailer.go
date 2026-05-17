package main

import (
	"bytes"
	_ "embed"
	"html/template"
	"time"

	gomail "github.com/go-mail/mail/v2"
)

//go:embed templates/reminder.tmpl
var reminderTmpl string

func (app *application) sendReminderEmail(to, itemTitle string, endAt time.Time) error {
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("02 Jan 2006 15:04")
		},
	}
	tmpl, err := template.New("reminder").Funcs(funcMap).Parse(reminderTmpl)
	if err != nil {
		return err
	}

	loc, _ := time.LoadLocation("Asia/Almaty")
	data := map[string]any{"ItemTitle": itemTitle, "EndAt": endAt.In(loc)}

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
