package service

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	commonErrors "github.com/slodkiadrianek/Go-API-template/common/errors"
	commonInterfaces "github.com/slodkiadrianek/Go-API-template/common/interfaces"
	authDTO "github.com/slodkiadrianek/Go-API-template/internal/auth/DTO"
	authModel "github.com/slodkiadrianek/Go-API-template/internal/auth/model"
	userModel "github.com/slodkiadrianek/Go-API-template/internal/user/model"
	"golang.org/x/crypto/bcrypt"
)

type authRepository interface {
	InsertRefreshToken(ctx context.Context, ipAddress, deviceInfo, refreshToken string, userID int) error
	GetRefreshTokenByTokenHash(ctx context.Context, refreshToken string) (authModel.TokenWithUserEmailToRefreshToken, error)
	UpdateLastTimeUsedToken(ctx context.Context, refreshToken string) error
	RemoveTokenFromDB(ctx context.Context, refreshToken string) error
	RemoveTokensFromDBByUserID(ctx context.Context, userID int) error
	ActivateAccount(ctx context.Context, userID int) error
}

type emailService interface {
	SendEmail(to, subject, body string) error
}

type userRepository interface {
	Create(ctx context.Context, user authDTO.CreateUser, hashedPassword []byte) error
	FindByEmail(ctx context.Context, email string) (userModel.User, error)
}

type AuthService struct {
	loggerService            commonInterfaces.Logger
	userRepository           userRepository
	authRepository           authRepository
	authenticationMiddleware commonInterfaces.AuthenticationMiddleware
	emailService             emailService
}

func NewAuthService(loggerService commonInterfaces.Logger, userRepository userRepository, authRepository authRepository, authenticationMiddleware commonInterfaces.AuthenticationMiddleware, emailService emailService) *AuthService {
	return &AuthService{
		loggerService:            loggerService,
		userRepository:           userRepository,
		authRepository:           authRepository,
		authenticationMiddleware: authenticationMiddleware,
		emailService:             emailService,
	}
}

func (as *AuthService) Register(ctx context.Context, user authDTO.CreateUser) error {
	userFromDB, err := as.userRepository.FindByEmail(ctx, user.Email)
	if err != nil {
		return err
	}

	if userFromDB.ID != 0 {
		err := errors.New("user with provided email already exists")
		as.loggerService.Info(err.Error(), user.Email)
		return commonErrors.NewAPIError(http.StatusBadRequest, err.Error())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		as.loggerService.Info(err.Error(), nil)
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	err = as.userRepository.Create(ctx, user, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

func (as *AuthService) Login(ctx context.Context, loginData authDTO.LoginUser, ipAddress, deviceInfo string) (string, []byte, error) {
	userFromDB, err := as.userRepository.FindByEmail(ctx, loginData.Email)
	if err != nil {
		return "", nil, err
	}

	if userFromDB.ID == 0 {
		err := errors.New("user with provided email not found")
		as.loggerService.Info(err.Error(), loginData.Email)
		return "", nil, commonErrors.NewAPIError(http.StatusNotFound, err.Error())
	}

	if !userFromDB.EmailVerified {
		userAuthData := userModel.User{
			ID:       userFromDB.ID,
			Email:    userFromDB.Email,
			Username: userFromDB.Username,
		}

		accessToken, err := as.authenticationMiddleware.GenerateAccessToken(userAuthData)
		if err != nil {
			return "", nil, err
		}

		host := os.Getenv("HOST_LINK")
		err = as.emailService.SendEmail(loginData.Email, "activation link", host+"/auth/activateAccount?token="+accessToken)
		if err != nil {
			return "", nil, err
		}

		err = errors.New("user with provided email is not verified, we sent to you mail with activation link")
		as.loggerService.Info(err.Error(), loginData.Email)
		return "", nil, commonErrors.NewAPIError(http.StatusUnauthorized, err.Error())
	}

	err = bcrypt.CompareHashAndPassword([]byte(userFromDB.Password), []byte(loginData.Password))
	if err != nil {
		err := errors.New("provided incorrect password")
		as.loggerService.Info(err.Error(), loginData.Email)
		return "", nil, commonErrors.NewAPIError(http.StatusUnauthorized, err.Error())

	}

	if err := ctx.Err(); err != nil {
		return "", nil, err
	}

	accessToken, err := as.authenticationMiddleware.GenerateAccessToken(userFromDB)
	if err != nil {
		return "", nil, err
	}

	refreshToken, err := as.authenticationMiddleware.GenerateRefreshToken()
	if err != nil {
		return "", nil, err
	}

	hashedRefreshToken := as.authenticationMiddleware.HashToken(refreshToken)

	if err := ctx.Err(); err != nil {
		return "", nil, err
	}

	err = as.authRepository.InsertRefreshToken(ctx, ipAddress, deviceInfo, hashedRefreshToken, userFromDB.ID)
	if err != nil {
		return "", nil, err
	}

	return accessToken, refreshToken, nil
}

func (as *AuthService) ActivateAccount(ctx context.Context, userID int) error {
	err := as.authRepository.ActivateAccount(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (as *AuthService) RefreshToken(ctx context.Context, refreshToken []byte) (string, error) {
	hashedRefreshToken := as.authenticationMiddleware.HashToken(refreshToken)

	tokenWithUserEmailToRefreshToken, err := as.authRepository.GetRefreshTokenByTokenHash(ctx, hashedRefreshToken)
	if err != nil {
		return "", err
	}

	if tokenWithUserEmailToRefreshToken.ID == 0 {
		err := errors.New("token not found")
		as.loggerService.Info(err.Error(), nil)
		return "", commonErrors.NewAPIError(http.StatusUnauthorized, err.Error())
	}

	if tokenWithUserEmailToRefreshToken.ExpiresAt.Before(time.Now()) {
		err := errors.New("refresh token expired")
		as.loggerService.Info(err.Error(), nil)
		return "", commonErrors.NewAPIError(http.StatusUnauthorized, err.Error())
	}

	err = as.authRepository.UpdateLastTimeUsedToken(ctx, hashedRefreshToken)
	if err != nil {
		return "", err
	}

	user := userModel.User{ID: tokenWithUserEmailToRefreshToken.ID, Email: tokenWithUserEmailToRefreshToken.Email, Username: tokenWithUserEmailToRefreshToken.Username}

	accessToken, err := as.authenticationMiddleware.GenerateAccessToken(user)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (as *AuthService) LogoutUser(ctx context.Context, refreshToken []byte) error {
	hashedRefreshToken := as.authenticationMiddleware.HashToken(refreshToken)

	err := as.authRepository.RemoveTokenFromDB(ctx, hashedRefreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (as *AuthService) LogoutUserFromAllDevices(ctx context.Context, userID int) error {
	err := as.authRepository.RemoveTokensFromDBByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}
