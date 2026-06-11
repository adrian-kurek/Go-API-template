package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	commonInterfaces "github.com/slodkiadrianek/Go-API-template/common/interfaces"
	jsonutil "github.com/slodkiadrianek/Go-API-template/common/json_util"
	authDTO "github.com/slodkiadrianek/Go-API-template/internal/auth/DTO"
	authMocks "github.com/slodkiadrianek/Go-API-template/test/mocks/auth"
	"github.com/stretchr/testify/mock"
)

func TestLogin(t *testing.T) {
	type args struct {
		title           string
		bodyRequestData authDTO.LoginUser
		setupMocks      func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter)
		wantErr         bool
		err             error
	}

	testsScenarios := []args{
		{
			title: "with proper data",
			bodyRequestData: authDTO.LoginUser{
				Email:    "joeDoe1@gmail.com",
				Password: "zaqwerfdsafsa@!44",
			},
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				mAuthService.On("Login", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("12312", []byte("1233445"), nil)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: false,
			err:     nil,
		},
		{
			title: "incorrect email format",
			bodyRequestData: authDTO.LoginUser{
				Email:    "joeDoe1gmail.com",
				Password: "zaqwerfdsafsa@!44",
			},
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				mAuthService.On("Login", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("12312", []byte("1233445"), nil)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("api error: the Email field must be a valid email address"),
		},
		{
			title: "authService.Login failed",
			bodyRequestData: authDTO.LoginUser{
				Email:    "joeDoe1@gmail.com",
				Password: "zaqwerfdsafsa@!44",
			},
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				mAuthService.On("Login", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", []byte(""), errors.New("failed to process the data"))
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("failed to process the data"),
		},
		{
			title: "context.DeadlineExceeded",
			bodyRequestData: authDTO.LoginUser{
				Email:    "joeDoe1@gmail.com",
				Password: "zaqwerfdsafsa@!44",
			},
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				mAuthService.On("Login", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", []byte(""), context.DeadlineExceeded)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("api error: "),
		},
	}

	for _, testScenario := range testsScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			loggerService := setupAuthHandlerDependencies()
			authorizationMiddleware, authService, w := testScenario.setupMocks()
			authController := NewAuthHandler(loggerService, authService, authorizationMiddleware)

			bodyBytes, err := jsonutil.MarshalData(testScenario.bodyRequestData)
			if err != nil {
				panic(err)
			}

			bodyReader := bytes.NewReader(bodyBytes)
			r, err := http.NewRequest("POST", "/auth/login", bodyReader)
			if err != nil {
				panic(err)
			}

			err = authController.Login(w, r)

			if (err != nil) != testScenario.wantErr {
				t.Errorf("Login() err = %v, wantErr = %v", err, testScenario.wantErr)
			}

			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("Login() error = %v, scenarioError = %v", err, testScenario.err)
				}
			}
		})
	}
}
