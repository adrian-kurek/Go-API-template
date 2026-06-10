package service

import (
	commonInterfaces "github.com/slodkiadrianek/Go-API-template/common/interfaces"
	gomail "gopkg.in/gomail.v2"
)

type EmailService struct {
	loggerService commonInterfaces.Logger
	hostEmail     string
	emailPassword string
}

func NewEmailService(hostEmail, emailPassword string, loggerService commonInterfaces.Logger) *EmailService {
	return &EmailService{
		loggerService: loggerService,
		hostEmail:     hostEmail,
		emailPassword: emailPassword,
	}
}

func (es *EmailService) SendEmail(to, subject, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", es.hostEmail)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)

	dialer := gomail.NewDialer("smtp.gmail.com", 587, es.hostEmail, es.emailPassword)
	if err := dialer.DialAndSend(message); err != nil {
		es.loggerService.Error("failed to send email", err)
		return err
	}

	es.loggerService.Info("email sent successfully", to)
	return nil
}
