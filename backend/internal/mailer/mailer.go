package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *mail.Dialer
	sender string
}

type VerificationEmailData struct {
	Code      string
	ExpiresIn int
}

type RentalEndingSoonData struct {
	ItemTitle string    `json:"item_title"`
	EndAt     time.Time `json:"end_at"`
}

func New(host string, port int, username, password, sender string) Mailer {

	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

var kzLocation *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Almaty")
	if err != nil {
		panic(err)
	}
	kzLocation = loc
}

func (m Mailer) Send(recipient, templateFile string, data interface{}) error {

	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("02 Jan 2006 15:04")
		},
	}

	tmpl, err := template.New("email").Funcs(funcMap).ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}
	return nil
}

func (m Mailer) SendVerificationEmail(to string, code string) error {
	data := VerificationEmailData{
		Code:      code,
		ExpiresIn: 10,
	}

	return m.Send(to, "verify_email.tmpl", data)
}

func (m Mailer) SendRentalEndingSoon(itemTitle, email string, endAt time.Time) error {
	localEndAt := endAt.In(kzLocation)
	data := RentalEndingSoonData{
		ItemTitle: itemTitle,
		EndAt:     localEndAt,
	}

	return m.Send(email, "reminder.tmpl", data)
}
