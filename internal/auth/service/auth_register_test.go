package service

import (
	"context"
	"errors"
	"testing"
	"time"

	commonInterfaces "github.com/slodkiadrianek/Go-API-template/common/interfaces"
	authDTO "github.com/slodkiadrianek/Go-API-template/internal/auth/DTO"
	userModel "github.com/slodkiadrianek/Go-API-template/internal/user/model"
	"github.com/slodkiadrianek/Go-API-template/test/mocks"
	authMocks "github.com/slodkiadrianek/Go-API-template/test/mocks/auth"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	type args struct {
		title     string
		user      authDTO.CreateUser
		setupMock func() (authRepository, userRepository, commonInterfaces.AuthenticationMiddleware)
		wantErr   bool
		err       error
	}
	testsScenarios := []args{
		{
			title: "with proper data",
			user: authDTO.CreateUser{
				Username:        "joeDoe",
				Email:           "joedoe@gmail.com",
				Password:        "zaq1@#$rfvbgt5",
				ConfirmPassword: "zaq1@#$rfvbgt5",
			},
			setupMock: func() (authRepository, userRepository, commonInterfaces.AuthenticationMiddleware) {
				mAuthRepository := new(authMocks.MockAuthRepository)
				mUserRepository := new(mocks.MockUserRepository)
				mAuthenticationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mUserRepository.On("FindUserByEmail", mock.Anything, mock.Anything).Return(userModel.User{
					ID: 0,
				}, nil)
				mAuthRepository.On("RegisterUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				return mAuthRepository, mUserRepository, mAuthenticationMiddleware
			},
			wantErr: false,
			err:     nil,
		},
		{
			title: "user with provided email already exists",
			user: authDTO.CreateUser{
				Username:        "joeDoe",
				Email:           "joedoe@gmail.com",
				Password:        "zaq1@#$rfvbgt5",
				ConfirmPassword: "zaq1@#$rfvbgt5",
			},
			setupMock: func() (authRepository, userRepository, commonInterfaces.AuthenticationMiddleware) {
				mAuthRepository := new(authMocks.MockAuthRepository)
				mUserRepository := new(mocks.MockUserRepository)
				mAuthenticationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mUserRepository.On("FindUserByEmail", mock.Anything, mock.Anything).Return(userModel.User{
					ID: 1,
				}, nil)
				mAuthRepository.On("RegisterUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				return mAuthRepository, mUserRepository, mAuthenticationMiddleware
			},
			wantErr: true,
			err:     errors.New("api error: user with provided email already exists"),
		},
		{
			title: "FindUserByEmail failed",
			user: authDTO.CreateUser{
				Username:        "joeDoe",
				Email:           "joedoe@gmail.com",
				Password:        "zaq1@#$rfvbgt5",
				ConfirmPassword: "zaq1@#$rfvbgt5",
			},
			setupMock: func() (authRepository, userRepository, commonInterfaces.AuthenticationMiddleware) {
				mAuthRepository := new(authMocks.MockAuthRepository)
				mUserRepository := new(mocks.MockUserRepository)
				mAuthenticationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mUserRepository.On("FindUserByEmail", mock.Anything, mock.Anything).Return(userModel.User{}, errors.New("failed to find user by email"))
				mAuthRepository.On("RegisterUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				return mAuthRepository, mUserRepository, mAuthenticationMiddleware
			},
			wantErr: true,
			err:     errors.New("failed to find user by email"),
		},
		{
			title: "RegisterUser failed",
			user: authDTO.CreateUser{
				Username:        "joeDoe",
				Email:           "joedoe@gmail.com",
				Password:        "zaq1@#$rfvbgt5",
				ConfirmPassword: "zaq1@#$rfvbgt5",
			},
			setupMock: func() (authRepository, userRepository, commonInterfaces.AuthenticationMiddleware) {
				mAuthRepository := new(authMocks.MockAuthRepository)
				mUserRepository := new(mocks.MockUserRepository)
				mAuthenticationMiddleware := new(authMocks.MockAuthorizationMiddleware)
				mUserRepository.On("FindUserByEmail", mock.Anything, mock.Anything).Return(userModel.User{
					ID: 0,
				}, nil)
				mAuthRepository.On("RegisterUser", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to insert user to DB"))
				return mAuthRepository, mUserRepository, mAuthenticationMiddleware
			},
			wantErr: true,
			err:     errors.New("failed to insert user to DB"),
		},
	}

	for _, testScenario := range testsScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			authRepository, userRepository, authorizationMiddleware := testScenario.setupMock()
			loggerService := setupAuthServiceDependencies()
			emailService := new(authMocks.MockEmailService)
			authService := NewAuthService(loggerService, userRepository, authRepository, authorizationMiddleware, emailService)

			err := authService.Register(ctx, testScenario.user)

			if (err != nil) != testScenario.wantErr {
				t.Errorf("Register() error = %v, wantErr = %v", err, testScenario.wantErr)
			}

			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("Register() error = %v, scenarioError = %v", err, testScenario.err)
				}
			}
		})
	}
}
