package mail

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/NBISweden/submitter/internal/config"
	"gopkg.in/gomail.v2"
)

type Mail struct {
	SMTPHost string
	SMTPport int
	Email    string
	Password string
	From     string
	Data     TemplateData
	Lookup   map[string]Notifiers
}

type TemplateData struct {
	Uploader      string
	DatasetID     string
	DatasetFolder string
}

type Notifiers struct {
	Template    string
	Subject     string
	Attachments []string
}

func Configure(c *config.Config) *Mail {
	m := &Mail{
		SMTPHost: c.SMTPHost,
		SMTPport: c.SMTPPort,
		Email:    c.Email,
		Password: c.Password,
		From:     c.Email,
		Data: TemplateData{
			Uploader:      c.Uploader,
			DatasetID:     c.DatasetID,
			DatasetFolder: c.DatasetFolder,
		},
		Lookup: map[string]Notifiers{
			"Submitter": {
				Template:    "internal/mail/templates/notify-submitter.html",
				Subject:     "Successful Ingestion of Your Dataset Submission",
				Attachments: []string{"data/stableIDs.txt"},
			},
			"BigPicture": {
				Template:    "internal/mail/templates/notify-bigpicture.html",
				Subject:     fmt.Sprintf("Dataset %s has been ingested", c.DatasetFolder),
				Attachments: []string{"data/dataset.txt", "data/policy.txt"},
			},
			"Jarno": {
				Template:    "internal/mail/templates/notify-jarno.html",
				Subject:     fmt.Sprintf("Dataset %s has been ingested", c.DatasetFolder),
				Attachments: []string{"data/dataset.txt", "data/rems.txt", "data/policy.txt"},
			},
		},
	}
	return m
}

func (mail *Mail) send(subject string, message string, reciever string, attachements []string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", mail.From)
	m.SetHeader("To", reciever)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", message)

	for _, file := range attachements {
		m.Attach(file)
	}

	d := gomail.NewDialer(mail.SMTPHost, mail.SMTPport, mail.Email, mail.Password)
	return d.DialAndSend(m)
}

func (mail *Mail) Notify(notifier string) error {
	htmlBody, err := renderTemplate(mail.Lookup[notifier].Template, mail.Data)
	if err != nil {
		return fmt.Errorf("Failed to render mail template: %v", err)
	}

	err = mail.send(mail.Lookup[notifier].Subject, htmlBody, mail.Email, mail.Lookup[notifier].Attachments)
	if err != nil {
		return fmt.Errorf("Failed to send notification %v", err)
	}

	return nil
}

func renderTemplate(filename string, data TemplateData) (string, error) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
