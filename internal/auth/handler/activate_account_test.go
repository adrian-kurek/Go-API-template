package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	commonInterfaces "github.com/slodkiadrianek/Go-API-template/common/interfaces"
	"github.com/slodkiadrianek/Go-API-template/common/request"
	authMocks "github.com/slodkiadrianek/Go-API-template/test/mocks/auth"
	"github.com/stretchr/testify/mock"
)

func TestActivateAccount(t *testing.T) {
	type args struct {
		title      string
		token      string
		setupMocks func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter)
		wantErr    bool
		err        error
	}

	testsScenarios := []args{
		{
			title: "with proper data",
			token: "123123123123123123123123123123",
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				r, err := http.NewRequest("GET", "/auth/activate?token=123123123123123123123123123123", nil)
				if err != nil {
					panic(err)
				}

				r = request.SetContext(r, "id", 1)
				mAuthorizationMiddleware.On("VerifyToken", mock.Anything).Return(r, nil)
				mAuthService.On("ActivateAccount", mock.Anything, mock.Anything).Return(nil)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: false,
			err:     nil,
		},
		{
			title: "authorization.VerifyToken() context.DeadlineExceeded",
			token: "123123123123123123123123123123",
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				r, err := http.NewRequest("GET", "/auth/activate?token=123123123123123123123123123123", nil)
				if err != nil {
					panic(err)
				}

				r = request.SetContext(r, "id", 1)
				mAuthorizationMiddleware.On("VerifyToken", mock.Anything).Return(r, context.DeadlineExceeded)
				mAuthService.On("ActivateAccount", mock.Anything, mock.Anything).Return(nil)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("api error: "),
		},
		{
			title: "authorization.VerifyToken() failed",
			token: "123123123123123123123123123123",
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				r, err := http.NewRequest("GET", "/auth/activate?token=123123123123123123123123123123", nil)
				if err != nil {
					panic(err)
				}

				r = request.SetContext(r, "id", 1)
				mAuthorizationMiddleware.On("VerifyToken", mock.Anything).Return(r, errors.New("failed to process data"))
				mAuthService.On("ActivateAccount", mock.Anything, mock.Anything).Return(nil)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("failed to process data"),
		},
		{
			title: "authService.ActivateAccount() failed",
			token: "123123123123123123123123123123",
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				r, err := http.NewRequest("GET", "/auth/activate?token=123123123123123123123123123123", nil)
				if err != nil {
					panic(err)
				}

				r = request.SetContext(r, "id", 1)
				mAuthorizationMiddleware.On("VerifyToken", mock.Anything).Return(r, nil)
				mAuthService.On("ActivateAccount", mock.Anything, mock.Anything).Return(errors.New("failed to process data"))
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("failed to process data"),
		},
		{
			title: "authService.ActivateAccount() context.DeadlineExceeded",
			token: "123123123123123123123123123123",
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				r, err := http.NewRequest("GET", "/auth/activate?token=123123123123123123123123123123", nil)
				if err != nil {
					panic(err)
				}

				r = request.SetContext(r, "id", 1)
				mAuthorizationMiddleware.On("VerifyToken", mock.Anything).Return(r, nil)
				mAuthService.On("ActivateAccount", mock.Anything, mock.Anything).Return(context.DeadlineExceeded)
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

			r, err := http.NewRequest("GET", "/auth/activate?token="+testScenario.token, nil)
			if err != nil {
				panic(err)
			}

			err = authController.ActivateAccount(w, r)

			if (err != nil) != testScenario.wantErr {
				t.Errorf("ActivateAccount() err = %v, wantErr = %v", err, testScenario.wantErr)
			}

			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("ActivateAccount() error = %v, scenarioError = %v", err, testScenario.err)
				}
			}
		})
	}
}
