package service

import (
	"os"
	"testing"

	config "github.com/slodkiadrianek/Go-API-template/config"
	mocks "github.com/slodkiadrianek/Go-API-template/test/mocks/auth"
	"github.com/stretchr/testify/mock"
)

func TestSendEmail(t *testing.T) {
	type args struct {
		title   string
		to      string
		wantErr bool
	}

	testsScenarios := []args{
		{
			title:   "with proper data",
			to:      "test@gmail.com",
			wantErr: false,
		},
		{
			title:   "with empty email",
			to:      "",
			wantErr: true,
		},
	}

	for _, testScenario := range testsScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			mEmailService := new(mocks.MockEmailService)
			mEmailService.On("SendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)

			err := config.SetupEnvVariables("../../../.env")
			if err != nil {
				panic(err)
			}
			loggerService := setupAuthServiceDependencies()
			hostEmail := os.Getenv("HOST_EMAIL")
			emailPassword := os.Getenv("PASSWORD_EMAIL")
			emailService := NewEmailService(hostEmail, emailPassword, loggerService)
			err = emailService.SendEmail(testScenario.to, "subject", "body")

			if (err != nil) != testScenario.wantErr {
				t.Errorf("SendEmail() error = %v, wantErr %v", err, testScenario.wantErr)
			}
		})
	}
}
