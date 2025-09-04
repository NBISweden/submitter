package mail

import (
	"gopkg.in/gomail.v2"
)

type Config struct {
	SMTPHost string
	SMTPport int
	Username string
	Password string
	From     string
}

var cfg Config

func Init(host string, port int, username, password, from string) {
	cfg = Config{
		SMTPHost: host,
		SMTPport: port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func Send(subject string, message string, reciever string, attachements []string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", reciever)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	for _, file := range attachements {
		m.Attach(file)
	}

	d := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPport, cfg.Username, cfg.Password)
	return d.DialAndSend(m)
}
