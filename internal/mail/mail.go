package mail

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

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
	Email       string
	CC          []string
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
				Email:       c.UploaderEmail,
				Template:    "internal/mail/templates/notify-submitter.html",
				Subject:     "Successful Ingestion of Your Dataset Submission",
				Attachments: []string{fmt.Sprintf("data/%s-stableIDs.txt", c.DatasetFolder)},
			},
			"BigPicture": {
				Email:       "submit@bigpicture.eu",
				Template:    "internal/mail/templates/notify-bigpicture.html",
				Subject:     fmt.Sprintf("Dataset %s has been ingested", c.DatasetFolder),
				Attachments: []string{"data/dataset.txt", "data/policy.txt"},
			},
			"Minttu": {
				Email:       "minttu.sauramo@hus.fi",
				CC:          []string{"jarno.laitinen@csc.fi"},
				Template:    "internal/mail/templates/notify-minttu.html",
				Subject:     fmt.Sprintf("Dataset %s has been ingested", c.DatasetFolder),
				Attachments: []string{"data/dataset.txt", "data/rems.txt", "data/policy.txt"},
			},
		},
	}
	return m
}

func (mail *Mail) send(subject string, message string, reciever string, attachements []string, ccs []string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", mail.From)
	m.SetHeader("To", reciever)
	m.SetHeader("Subject", subject)

	if len(ccs) > 0 {
		addresses := make([]string, 0, len(ccs))
		for _, email := range ccs {
			addresses = append(addresses, m.FormatAddress(email, ""))
		}
		m.SetHeader("Cc", addresses...)
	}

	m.SetBody("text/html", message)

	// Enforce that the wanted attachements are files that exists
	if err := attachementsExists(attachements); err != nil {
		return err
	}
	for _, file := range attachements {
		m.Attach(file)
	}

	d := gomail.NewDialer(mail.SMTPHost, mail.SMTPport, mail.Email, mail.Password)
	fmt.Printf("[mail] notified <%s> about dataset completion\n", reciever)
	return d.DialAndSend(m)
}

func (mail *Mail) Notify(notifier string, dryRun bool) error {
	htmlBody, err := renderTemplate(mail.Lookup[notifier].Template, mail.Data)
	if err != nil {
		return fmt.Errorf("failed to render mail template: %v", err)
	}

	if dryRun {
		fmt.Printf("[mail] using <%s> instead of <%s> during dryrun\n", mail.Email, mail.Lookup[notifier].Email)
		err = mail.send(mail.Lookup[notifier].Subject, htmlBody, mail.Email, mail.Lookup[notifier].Attachments, nil) // no cc on dry run
	}

	if !dryRun {
		err = mail.send(mail.Lookup[notifier].Subject, htmlBody, mail.Lookup[notifier].Email, mail.Lookup[notifier].Attachments, mail.Lookup[notifier].CC)
	}
	if err != nil {
		return fmt.Errorf("failed to send mail notification %v", err)
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

func attachementsExists(attachements []string) error {
	for _, attachement := range attachements {
		info, err := os.Stat(attachement)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", attachement)
			}
			return fmt.Errorf("error checking file %s: %w", attachement, err)
		}

		if info.IsDir() {
			return fmt.Errorf("path is a directory, not a file: %s", attachement)
		}

		if info.Size() == 0 {
			return fmt.Errorf("file is empty: %s", attachement)
		}
	}

	return nil
}
