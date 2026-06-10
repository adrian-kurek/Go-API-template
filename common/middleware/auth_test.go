package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/slodkiadrianek/Go-API-template/common/interfaces"
	"github.com/slodkiadrianek/Go-API-template/common/log"
	config "github.com/slodkiadrianek/Go-API-template/config"
	"github.com/slodkiadrianek/Go-API-template/internal/user/model"
	userModel "github.com/slodkiadrianek/Go-API-template/internal/user/model"
	"github.com/slodkiadrianek/Go-API-template/test/mocks"
	"github.com/stretchr/testify/mock"
)

func setupAuthControllerDependencies() (*log.Logger, string, string) {
	loggerService := log.NewLogger("./logs", "2006-01-02", "15:04:05")
	defer func() {
		if closeErr := loggerService.Close(); closeErr != nil {
			fmt.Errorf("failed to properly close file with logs:%s", closeErr.Error())
		}
	}()

	err := config.SetupEnvVariables("../../.env")
	if err != nil {
		panic(err)
	}

	accessTokenSecret, ok := os.LookupEnv("ACCESS_TOKEN_SECRET")
	if !ok {
		err := errors.New("ACCESS_TOKEN_SECRET variable has not been initialized")
		loggerService.Error(err.Error(), map[string]string{
			"variable": "ACCESS_TOKEN_SECRET",
		})
		panic(err)
	}

	refreshTokenSecret, ok := os.LookupEnv("REFRESH_TOKEN_SECRET")
	if !ok {
		err := errors.New("REFRESH_TOKEN_SECRET variable has not been initialized")
		loggerService.Error(err.Error(), map[string]string{
			"variable": "REFRESH_TOKEN_SECRET",
		})
		panic(err)
	}

	return loggerService, accessTokenSecret, refreshTokenSecret
}

func TestGenerateRefreshToken(t *testing.T) {
	type args struct {
		title   string
		wantErr bool
		err     error
	}

	testsScenarios := []args{
		{
			title:   "with proper data",
			wantErr: false,
			err:     nil,
		},
	}

	for _, testScenario := range testsScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			loggerService, accessTokenSecret, refreshTokenSecret := setupAuthControllerDependencies()
			cacheService := new(mocks.MockCacheService)
			authorizationMiddleware := NewAuthenticationMiddleware(accessTokenSecret, refreshTokenSecret, loggerService, cacheService)

			token, err := authorizationMiddleware.GenerateRefreshToken()

			if (err != nil) != testScenario.wantErr {
				t.Errorf("GenerateRefreshToken() error = %v, wantErr %v", err, testScenario.wantErr)
				return
			}

			if !testScenario.wantErr && len(token) == 0 {
				t.Errorf("GenerateRefreshToken() did not generate a token")
			}
		})
	}
}

func TestHashToken(t *testing.T) {
	type args struct {
		title string
		token []byte
		want  string
	}

	testsScenarios := []args{
		{
			title: "with proper data",
			token: []byte("testToken"),
			want:  "4b4a2dd847324503f0febd6955148a7737ca1c9a1ceef7690e0c2b827577ec5f",
		},
	}

	for _, testScenario := range testsScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			loggerService, accessTokenSecret, refreshTokenSecret := setupAuthControllerDependencies()
			cacheService := new(mocks.MockCacheService)
			authorizationMiddleware := NewAuthenticationMiddleware(accessTokenSecret, refreshTokenSecret, loggerService, cacheService)

			if got := authorizationMiddleware.HashToken(testScenario.token); got != testScenario.want {
				t.Errorf("HashToken() = %v, want %v", got, testScenario.want)
			}
		})
	}
}

func TestGenerateAccessToken(t *testing.T) {
	type args struct {
		title   string
		wantErr bool
	}

	testsScenarios := []args{
		{
			title:   "with proper data",
			wantErr: false,
		},
	}

	for _, testScenario := range testsScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			loggerService, accessTokenSecret, refreshTokenSecret := setupAuthControllerDependencies()
			cacheService := new(mocks.MockCacheService)
			authorizationMiddleware := NewAuthenticationMiddleware(accessTokenSecret, refreshTokenSecret, loggerService, cacheService)

			_, err := authorizationMiddleware.GenerateAccessToken(model.User{ID: 1, Email: "jode@gmail.com", Username: "jode1"})
			if (err != nil) != testScenario.wantErr {
				t.Errorf("GenerateAccessToken() err = %v, wantErr = %v", err, testScenario.wantErr)
			}
		})
	}
}

func TestParseClaimsToken(t *testing.T) {
	type args struct {
		title            string
		expectedUserData userModel.UserClaims
		wantErr          bool
	}

	testsScenarios := []args{
		{
			title: "with proper data",
			expectedUserData: userModel.UserClaims{
				ID:       1,
				Email:    "jode@gmail.com",
				Username: "jode1",
			},
			wantErr: false,
		},
	}

	for _, testScenario := range testsScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			loggerService, accessTokenSecret, refreshTokenSecret := setupAuthControllerDependencies()
			cacheService := new(mocks.MockCacheService)
			authorizationMiddleware := NewAuthenticationMiddleware(accessTokenSecret, refreshTokenSecret, loggerService, cacheService)

			token, err := authorizationMiddleware.GenerateAccessToken(userModel.User{
				ID:       1,
				Email:    "jode@gmail.com",
				Username: "jode1",
			})
			if err != nil {
				panic(err)
			}

			_, user, err := authorizationMiddleware.parseClaimsFromToken(token)

			if (err != nil) != testScenario.wantErr {
				t.Errorf("parseClaimsFromToken() err = %v, wantErr = %v", err, testScenario.wantErr)
			}

			if user.ID != testScenario.expectedUserData.ID || user.Email != testScenario.expectedUserData.Email || user.Username != testScenario.expectedUserData.Username {
				t.Errorf("parseClaimsFromToken() err = %v, wantErr = %v", err, testScenario.wantErr)
			}
		})
	}
}

func TestVerifyToken(t *testing.T) {
	type args struct {
		title     string
		token     string
		setupMock func() interfaces.CacheService
		wantErr   bool
		err       error
	}

	testsScenarios := []args{
		{
			title: "with proper data",
			token: "",
			setupMock: func() interfaces.CacheService {
				mCacheService := new(mocks.MockCacheService)
				mCacheService.On("Exists", mock.Anything, mock.Anything).Return(int64(0), nil)
				return mCacheService
			},
			wantErr: false,
			err:     nil,
		},
		{
			title: "lack of token",
			token: " eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvZGVAZ21haWwuY29tIiwiZXhwIjoxNzcxNzc4NjEzLCJpZCI6MSwidXNlcm5hbWUiOiJqb2RlMSJ9.lZj5y7h_WpYRR5J7bDxZiFyoXlARzMbnlHjLGFGRR3U",
			setupMock: func() interfaces.CacheService {
				mCacheService := new(mocks.MockCacheService)
				mCacheService.On("Exists", mock.Anything, mock.Anything).Return(int64(0), nil)
				return mCacheService
			},
			wantErr: true,
			err:     errors.New("api error: failed to authorize a user"),
		},
		{
			title: "invalid token",
			token: "Bearer yJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvZGVAZ21haWwuY29tIiwiZXhwIjoxNzcxNzc4NjEzLCJpZCI6MSwidXNlcm5hbWUiOiJqb2RlMSJ9.lZj5y7h_WpYRR5J7bDxZiFyoXlARzMbnlHjLGFGRR3U",
			setupMock: func() interfaces.CacheService {
				mCacheService := new(mocks.MockCacheService)
				mCacheService.On("Exists", mock.Anything, mock.Anything).Return(int64(0), nil)
				return mCacheService
			},
			wantErr: true,
			err:     errors.New("api error: provided token is invalid"),
		},
		{
			title: "cacheService.Exists() failed",
			token: "",
			setupMock: func() interfaces.CacheService {
				mCacheService := new(mocks.MockCacheService)
				mCacheService.On("Exists", mock.Anything, mock.Anything).Return(int64(0), errors.New("failed to check existance data in cache"))
				return mCacheService
			},
			wantErr: true,
			err:     errors.New("failed to check existance data in cache"),
		},
		{
			title: "cacheService.Exists() failed",
			token: "",
			setupMock: func() interfaces.CacheService {
				mCacheService := new(mocks.MockCacheService)
				mCacheService.On("Exists", mock.Anything, mock.Anything).Return(int64(1), nil)
				return mCacheService
			},
			wantErr: true,
			err:     errors.New("api error: token blacklisted"),
		},
	}

	for _, testScenario := range testsScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			loggerService, accessTokenSecret, refreshTokenSecret := setupAuthControllerDependencies()
			cacheService := testScenario.setupMock()
			authorizationMiddleware := NewAuthenticationMiddleware(accessTokenSecret, refreshTokenSecret, loggerService, cacheService)

			token, err := authorizationMiddleware.GenerateAccessToken(userModel.User{
				ID:       1,
				Email:    "jode1@gmail.com",
				Username: "jode1",
			})
			if err != nil {
				panic(err)
			}

			r, err := http.NewRequest("POST", "/auth/verify", nil)
			if err != nil {
				panic(err)
			}

			if len(testScenario.token) > 0 {
				r.Header.Set("Authorization", testScenario.token)
			} else {
				r.Header.Set("Authorization", "Bearer "+token)
			}

			_, err = authorizationMiddleware.VerifyToken(r)

			if (err != nil) != testScenario.wantErr {
				t.Errorf("parseClaimsFromToken() err = %v, wantErr = %v", err, testScenario.wantErr)
			}

			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("parseClaimsFromToken() err = %v, testScenario.err = %v", err, testScenario.err)
				}
			}
		})
	}
}

func TestBlacklistUser(t *testing.T) {
	type args struct {
		title     string
		token     string
		setupMock func() interfaces.CacheService
		wantErr   bool
		err       error
	}

	testsScenarios := []args{
		{
			title: "with proper data",
			token: "",
			setupMock: func() interfaces.CacheService {
				mCacheService := new(mocks.MockCacheService)
				mCacheService.On("Exists", mock.Anything, mock.Anything).Return(int64(0), nil)
				mCacheService.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				return mCacheService
			},
			wantErr: false,
			err:     nil,
		},
		{
			title: "lack of token",
			token: "1",
			setupMock: func() interfaces.CacheService {
				mCacheService := new(mocks.MockCacheService)
				return mCacheService
			},
			wantErr: true,
			err:     errors.New("api error: failed to authorize a user"),
		},
		{
			title: "invalid token",
			token: "Bearer 1",
			setupMock: func() interfaces.CacheService {
				mCacheService := new(mocks.MockCacheService)
				return mCacheService
			},
			wantErr: true,
			err:     errors.New("api error: failed to read token"),
		},
		{
			title: "Exists failed",
			token: "",
			setupMock: func() interfaces.CacheService {
				mCacheService := new(mocks.MockCacheService)
				mCacheService.On("Exists", mock.Anything, mock.Anything).Return(int64(0), errors.New("failed to check data in cache"))
				return mCacheService
			},
			wantErr: true,
			err:     errors.New("failed to check data in cache"),
		},
		{
			title: "token already blacklisted",
			token: "",
			setupMock: func() interfaces.CacheService {
				mCacheService := new(mocks.MockCacheService)
				mCacheService.On("Exists", mock.Anything, mock.Anything).Return(int64(1), nil)
				return mCacheService
			},
			wantErr: true,
			err:     errors.New("api error: token already blacklisted"),
		},
		{
			title: "Set Failed",
			token: "",
			setupMock: func() interfaces.CacheService {
				mCacheService := new(mocks.MockCacheService)
				mCacheService.On("Exists", mock.Anything, mock.Anything).Return(int64(0), nil)
				mCacheService.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to set data in cache"))
				return mCacheService
			},
			wantErr: true,
			err:     errors.New("failed to set data in cache"),
		},
	}

	for _, testScenario := range testsScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			loggerService, accessTokenSecret, refreshTokenSecret := setupAuthControllerDependencies()
			cacheService := testScenario.setupMock()
			authorizationMiddleware := NewAuthenticationMiddleware(accessTokenSecret, refreshTokenSecret, loggerService, cacheService)

			token, err := authorizationMiddleware.GenerateAccessToken(userModel.User{
				ID:       1,
				Email:    "jode1@gmail.com",
				Username: "jode1",
			})
			if err != nil {
				panic(err)
			}

			r, err := http.NewRequest("POST", "/auth/verify", nil)
			if err != nil {
				panic(err)
			}

			if len(testScenario.token) > 0 {
				r.Header.Set("Authorization", testScenario.token)
			} else {
				r.Header.Set("Authorization", "Bearer "+token)
			}

			err = authorizationMiddleware.BlacklistUser(r)
			if (err != nil) != testScenario.wantErr {
				t.Errorf("BlacklistUser() err = %v, wantErr = %v", err, testScenario.wantErr)
			}

			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("BlacklistUser() err = %v, testScenario.err = %v", err, testScenario.err)
				}
			}
		})
	}
}