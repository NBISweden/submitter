package mail

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"os"

	"github.com/NBISweden/submitter/cmd"
	"github.com/spf13/cobra"
	"gopkg.in/gomail.v2"
)

//go:embed templates/*.html
var templateFS embed.FS
var dryRun bool
var configPath string

var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "Send mail notifications",
	Long:  "Send mail notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := NewConfig(configPath)
		if err != nil {
			return err
		}
		m := New(conf)
		for _, recipient := range []string{"BigPicture", "Minttu", "Submitter"} {
			if err := m.Notify(recipient, dryRun); err != nil {
				return fmt.Errorf("failed to notify %s: %w", recipient, err)
			}
		}

		return nil
	},
}

func init() {
	cmd.AddCommand(mailCmd)
	mailCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Toggles dry-run mode. Dry run will send all emails to the address in configuration.Email (env or yaml conf)")
	mailCmd.Flags().StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
}

type Mail struct {
	smtpHost string
	smtpPort int
	email    string
	password string
	from     string
	data     TemplateData
	lookup   map[string]Notifiers
}

type TemplateData struct {
	uploader      string
	datasetID     string
	datasetFolder string
}

type Notifiers struct {
	email       string
	cc          []string
	template    string
	subject     string
	attachments []string
}

func New(c *Config) *Mail {
	m := &Mail{
		smtpHost: c.smtpHost,
		smtpPort: c.smtpPort,
		email:    c.emailAddress,
		password: c.emailPassword,
		from:     c.emailAddress,
		data: TemplateData{
			uploader:      c.uploaderName,
			datasetID:     c.datasetID,
			datasetFolder: c.datasetFolder,
		},
		lookup: map[string]Notifiers{
			"Submitter": {
				email:       c.uploaderEmail,
				template:    "notify-submitter.html",
				subject:     "Successful Ingestion of Your Dataset Submission",
				attachments: []string{fmt.Sprintf("/data/%s-stableIDs.txt", c.datasetFolder)},
			},
			"BigPicture": {
				email:       "submit@bigpicture.eu",
				template:    "notify-bigpicture.html",
				subject:     fmt.Sprintf("Dataset %s has been ingested", c.datasetFolder),
				attachments: []string{"/data/dataset.txt", "data/policy.txt"},
			},
			"Minttu": {
				email:       "minttu.sauramo@hus.fi",
				cc:          []string{"jarno.laitinen@csc.fi"},
				template:    "notify-minttu.html",
				subject:     fmt.Sprintf("Dataset %s has been ingested", c.datasetFolder),
				attachments: []string{"/data/dataset.txt", "data/rems.txt", "data/policy.txt"},
			},
		},
	}
	return m
}

func (mail *Mail) send(subject string, message string, reciever string, attachements []string, ccs []string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", mail.from)
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

	d := gomail.NewDialer(mail.smtpHost, mail.smtpPort, mail.email, mail.password)
	slog.Info("[mail] notification sent about dataset completion", "reciever", reciever)
	return d.DialAndSend(m)
}

func (mail *Mail) Notify(notifier string, dryRun bool) error {
	htmlBody, err := renderTemplate(mail.lookup[notifier].template, mail.data)
	if err != nil {
		return fmt.Errorf("failed to render mail template: %v", err)
	}

	if dryRun {
		slog.Info(fmt.Sprintf("[mail] dry-run enabled, using <%s> instead of <%s>", mail.email, mail.lookup[notifier].email))
		err = mail.send(mail.lookup[notifier].subject, htmlBody, mail.email, mail.lookup[notifier].attachments, nil)
	}

	if !dryRun {
		err = mail.send(mail.lookup[notifier].subject, htmlBody, mail.lookup[notifier].email, mail.lookup[notifier].attachments, mail.lookup[notifier].cc)
	}
	if err != nil {
		return fmt.Errorf("failed to send mail notification %v", err)
	}

	return nil
}

func renderTemplate(filename string, data TemplateData) (string, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/"+filename)
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
