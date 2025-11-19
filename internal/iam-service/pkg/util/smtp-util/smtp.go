package smtp_util

import (
	"errors"

	"gopkg.in/gomail.v2"
)

// Config 邮件服务器配置 [EN] Config mail server configuration
type Config struct {
	Host     string `json:"host" mapstructure:"host"`
	Port     int    `json:"port" mapstructure:"port"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`

	DefaultFrom From `json:"default_from" mapstructure:"default_from"`
}

type From struct {
	Email string `json:"email" mapstructure:"email"`
	Name  string `json:"name" mapstructure:"name"`
}

var _mailer *mailer

type mailer struct {
	cfg    Config
	dialer *gomail.Dialer
}

func (m *mailer) SendMail(toEmails []string, subject, contentType, body string, from ...*From) error {
	if len(toEmails) == 0 {
		return errors.New("email recipient cannot be empty")
	}
	if subject == "" || body == "" {
		return errors.New("email subject or body cannot be empty")
	}
	msg := gomail.NewMessage()
	// 设置发件人 [EN] Set sender
	var fromEmail *From
	if len(from) > 0 {
		fromEmail = from[0]
	} else {
		fromEmail = &m.cfg.DefaultFrom
	}
	if fromEmail.Name != "" {
		msg.SetAddressHeader("From", fromEmail.Email, fromEmail.Name)
	} else {
		msg.SetHeader("From", fromEmail.Email)
	}
	// 设置收件人 [EN] Set recipients
	msg.SetHeader("To", toEmails...)
	// 设置主题 [EN] Set theme
	msg.SetHeader("Subject", subject)
	// 设置正文 [EN] Set text
	msg.SetBody(contentType, body)
	// 发送邮件 [EN] Send email
	return m.dialer.DialAndSend(msg)
}

// --- API ---

func Init(cfg Config) error {
	if _mailer != nil {
		return errors.New("mailer already init")
	}
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	_mailer = &mailer{
		cfg:    cfg,
		dialer: dialer,
	}
	return nil
}

func SendEmail(toEmails []string, subject, contentType, body string, from ...*From) error {
	if _mailer == nil {
		return errors.New("mailer not init")
	}
	return _mailer.SendMail(toEmails, subject, contentType, body, from...)
}
