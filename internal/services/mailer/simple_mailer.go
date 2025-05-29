package mailer

import (
	"fmt"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

type SimpleMailer struct {
	From     string
	SMTPConf SMTPConfig
	baseUrl  string
}

func NewSimpleMailer(conf SMTPConfig, baseUrl string) *SimpleMailer {
	return &SimpleMailer{
		From:     conf.User,
		SMTPConf: conf,
		baseUrl:  baseUrl,
	}
}

func (m *SimpleMailer) SendResetEmail(to, token string, ttl time.Duration) error {
	server := mail.NewSMTPClient()
	server.Host = m.SMTPConf.Host
	server.Port = m.SMTPConf.Port
	server.Username = m.SMTPConf.User
	server.Password = m.SMTPConf.Password
	server.Encryption = mail.EncryptionSTARTTLS
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return fmt.Errorf("SMTP connect failed: %w", err)
	}

	link := fmt.Sprintf("%s/reset.html?email=%s&token=%s", m.baseUrl, to, token)

	msg := mail.NewMSG()
	msg.SetFrom("系统通知 <"+m.From+">").
		AddTo(to).
		SetSubject("密码重置").
		SetBody(mail.TextPlain,
			fmt.Sprintf("请点击以下链接重置密码（%s 内有效）：\n%s", ttl.String(), link),
		)

	return msg.Send(smtpClient)
}

func (m *SimpleMailer) SendVerificationEmail(to, token string) error {
	server := mail.NewSMTPClient()
	server.Host = m.SMTPConf.Host
	server.Port = m.SMTPConf.Port
	server.Username = m.SMTPConf.User
	server.Password = m.SMTPConf.Password
	server.Encryption = mail.EncryptionSTARTTLS
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return fmt.Errorf("SMTP connect failed: %w", err)
	}

	message := token

	msg := mail.NewMSG()
	msg.SetFrom("系统通知 <"+m.From+">").
		AddTo(to).
		SetSubject("邮箱验证").
		SetBody(mail.TextPlain,
			fmt.Sprintf("请输入您的验证码\n%s", message),
		)

	return msg.Send(smtpClient)
}
