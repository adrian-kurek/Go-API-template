package handler

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	commonInterfaces "github.com/slodkiadrianek/Go-API-template/common/interfaces"
	"github.com/slodkiadrianek/Go-API-template/common/request"
	authMocks "github.com/slodkiadrianek/Go-API-template/test/mocks/auth"
	"github.com/stretchr/testify/mock"
)

func TestLogoutUser(t *testing.T) {
	type args struct {
		title      string
		setCookie  bool
		setupMocks func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter)
		wantErr    bool
		err        error
	}

	testsScenarios := []args{
		{
			title:     "with proper data",
			setCookie: true,
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				mAuthorizationMiddleware.On("BlacklistUser", mock.Anything, mock.Anything).Return(nil)
				mAuthService.On("LogoutUser", mock.Anything, mock.Anything).Return(nil)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: false,
			err:     nil,
		},
		{
			title:     "authorization.BlacklistUser failed",
			setCookie: true,
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				mAuthorizationMiddleware.On("BlacklistUser", mock.Anything, mock.Anything).Return(errors.New("failed to process data"))
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("failed to process data"),
		},
		{
			title:     "authorization.BlacklistUser context.DeadlineExceeded",
			setCookie: true,
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				mAuthorizationMiddleware.On("BlacklistUser", mock.Anything, mock.Anything).Return(context.DeadlineExceeded)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("api error: "),
		},
		{
			title:     "failed to read cookie",
			setCookie: false,
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				mAuthorizationMiddleware.On("BlacklistUser", mock.Anything, mock.Anything).Return(nil)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("http: named cookie not present"),
		},
		{
			title:     "authService.LogoutUser failed",
			setCookie: true,
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				mAuthorizationMiddleware.On("BlacklistUser", mock.Anything, mock.Anything).Return(nil)
				mAuthService.On("LogoutUser", mock.Anything, mock.Anything).Return(errors.New("failed to process data"))
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("failed to process data"),
		},
		{
			title:     "authService.LogoutUser context.DeadlineExceeded",
			setCookie: true,
			setupMocks: func() (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				mAuthorizationMiddleware.On("BlacklistUser", mock.Anything, mock.Anything).Return(nil)
				mAuthService.On("LogoutUser", mock.Anything, mock.Anything).Return(context.DeadlineExceeded)
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

			r, err := http.NewRequest("DELETE", "/auth/logout", nil)
			if err != nil {
				panic(err)
			}

			if testScenario.setCookie {
				cookie := http.Cookie{
					Name:  "refreshToken",
					Value: hex.EncodeToString([]byte("1")),
				}
				r.AddCookie(&cookie)
			}

			err = authController.LogoutUser(w, r)

			if (err != nil) != testScenario.wantErr {
				t.Errorf("LogoutUser() err = %v, wantErr = %v", err, testScenario.wantErr)
			}

			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("LogoutUser() error = %v, scenarioError = %v", err, testScenario.err)
				}
			}
		})
	}
}

func TestLogoutUserFromAllDevices(t *testing.T) {
	type args struct {
		title          string
		setIDInContext bool
		setupMocks     func(setInContext bool) (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter)
		wantErr        bool
		err            error
	}

	testsScenarios := []args{
		{
			title:          "with proper data",
			setIDInContext: true,
			setupMocks: func(setInContext bool) (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				r, err := http.NewRequest("DELETE", "/auth/logoutAll", nil)
				if err != nil {
					panic(err)
				}

				if setInContext {
					r = request.SetContext(r, "id", 12)
				}

				mAuthorizationMiddleware.On("VerifyToken", mock.Anything).Return(r, nil)
				mAuthService.On("LogoutUserFromAllDevices", mock.Anything, mock.Anything).Return(nil)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: false,
			err:     nil,
		},
		{
			title:          "authorization.VerifyToken failed",
			setIDInContext: true,
			setupMocks: func(setInContext bool) (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				r, err := http.NewRequest("GET", "/auth/verify", nil)
				if err != nil {
					panic(err)
				}

				if setInContext {
					r = request.SetContext(r, "id", 12)
				}

				mAuthorizationMiddleware.On("VerifyToken", mock.Anything).Return(r, errors.New("failed to process data"))
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("failed to process data"),
		},
		{
			title:          "authorization.VerifyToken context.DeadlineExceeded",
			setIDInContext: true,
			setupMocks: func(setInContext bool) (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				r, err := http.NewRequest("DELETE", "/auth/logoutAll", nil)
				if err != nil {
					panic(err)
				}

				if setInContext {
					r = request.SetContext(r, "id", 12)
				}

				mAuthorizationMiddleware.On("VerifyToken", mock.Anything).Return(r, context.DeadlineExceeded)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("api error: "),
		},
		{
			title:          "failed to read id from token",
			setIDInContext: false,
			setupMocks: func(setInContext bool) (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				r, err := http.NewRequest("DELETE", "/auth/logoutAll", nil)
				if err != nil {
					panic(err)
				}

				if setInContext {
					r = request.SetContext(r, "id", 12)
				}

				mAuthorizationMiddleware.On("VerifyToken", mock.Anything).Return(r, nil)
				mAuthService.On("LogoutUserFromAllDevices", mock.Anything, mock.Anything).Return(nil)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("failed to read user from context"),
		},
		{
			title:          "authService.LogoutUserFromAllDevices failed",
			setIDInContext: true,
			setupMocks: func(setInContext bool) (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				r, err := http.NewRequest("DELETE", "/auth/logoutAll", nil)
				if err != nil {
					panic(err)
				}

				if setInContext {
					r = request.SetContext(r, "id", 12)
				}

				mAuthorizationMiddleware.On("VerifyToken", mock.Anything).Return(r, nil)
				mAuthService.On("LogoutUserFromAllDevices", mock.Anything, mock.Anything).Return(errors.New("failed to process data"))
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("failed to process data"),
		},
		{
			title:          "authService.LogoutUserFromAllDevices context.DeadlineExceeded",
			setIDInContext: true,
			setupMocks: func(setInContext bool) (commonInterfaces.AuthenticationMiddleware, authService, http.ResponseWriter) {
				mAuthorizationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mAuthService := new(authMocks.MockAuthService)
				r, err := http.NewRequest("DELETE", "/auth/logoutAll", nil)
				if err != nil {
					panic(err)
				}

				if setInContext {
					r = request.SetContext(r, "id", 12)
				}

				mAuthorizationMiddleware.On("VerifyToken", mock.Anything).Return(r, nil)
				mAuthService.On("LogoutUserFromAllDevices", mock.Anything, mock.Anything).Return(context.DeadlineExceeded)
				return mAuthorizationMiddleware, mAuthService, httptest.NewRecorder()
			},
			wantErr: true,
			err:     errors.New("api error: "),
		},
	}

	for _, testScenario := range testsScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			loggerService := setupAuthHandlerDependencies()
			authorizationMiddleware, authService, w := testScenario.setupMocks(testScenario.setIDInContext)
			authController := NewAuthHandler(loggerService, authService, authorizationMiddleware)

			r, err := http.NewRequest("DELETE", "/auth/logoutAll", nil)
			if err != nil {
				panic(err)
			}

			err = authController.LogoutUserFromAllDevices(w, r)

			if (err != nil) != testScenario.wantErr {
				t.Errorf("LogoutUserFromAllDevices() err = %v, wantErr = %v", err, testScenario.wantErr)
			}

			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("LogoutUserFromAllDevices() error = %v, scenarioError = %v", err, testScenario.err)
				}
			}
		})
	}
}
