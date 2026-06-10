package handler

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"time"

	commonInterfaces "github.com/slodkiadrianek/Go-API-template/common/interfaces"
	"github.com/slodkiadrianek/Go-API-template/common/middleware"
	"github.com/slodkiadrianek/Go-API-template/internal/auth/DTO"

	commonErrors "github.com/slodkiadrianek/Go-API-template/common/errors"
	"github.com/slodkiadrianek/Go-API-template/common/request"
	"github.com/slodkiadrianek/Go-API-template/common/response"
)

const authTimeout = 2 * time.Second

type authService interface {
	Register(ctx context.Context, user dto.CreateUser) error
	Login(ctx context.Context, loginData dto.LoginUser, ipAddress, deviceInfo string) (string, []byte, error)
	RefreshToken(ctx context.Context, token []byte) (string, error)
	LogoutUser(ctx context.Context, refreshToken []byte) error
	LogoutUserFromAllDevices(ctx context.Context, userID int) error
	ActivateAccount(ctx context.Context, userID int) error
}

type AuthHandler struct {
	loggerService            commonInterfaces.Logger
	authService              authService
	authenticationMiddleware commonInterfaces.AuthenticationMiddleware
}

func NewAuthHandler(loggerService commonInterfaces.Logger, authService authService, authenticationMiddleware commonInterfaces.AuthenticationMiddleware) *AuthHandler {
	return &AuthHandler{
		loggerService:            loggerService,
		authService:              authService,
		authenticationMiddleware: authenticationMiddleware,
	}
}

func (ah *AuthHandler) Register(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	reqData, err := request.ReadBody[dto.CreateUser](r)
	if err != nil {
		return commonErrors.NewAPIError(http.StatusUnprocessableEntity, "provided invalid json format")
	}

	err = middleware.ValidateRequestData(reqData)
	if err != nil {
		return err
	}

	err = ah.authService.Register(ctx, *reqData)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			ah.loggerService.Info("request timed out", r.URL.Path)
			return commonErrors.NewAPIError(http.StatusRequestTimeout, "")
		}
		return err
	}

	response.Send(w, http.StatusOK, map[string]string{})
	return nil
}

func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	reqData, err := request.ReadBody[dto.LoginUser](r)
	if err != nil {
		return commonErrors.NewAPIError(http.StatusUnprocessableEntity, "provided invalid json format")
	}

	err = middleware.ValidateRequestData(reqData)
	if err != nil {
		return err
	}

	ipAddress := r.RemoteAddr
	deviceInfo := r.UserAgent()

	accessToken, refreshToken, err := ah.authService.Login(ctx, *reqData, ipAddress, deviceInfo)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			ah.loggerService.Info("request timed out", r.URL.Path)
			return commonErrors.NewAPIError(http.StatusRequestTimeout, "")
		}
		return err
	}

	expiration := time.Now().Add(7 * 24 * time.Hour)

	cookie := http.Cookie{
		Name:     "refreshToken",
		Value:    hex.EncodeToString(refreshToken),
		Expires:  expiration,
		Secure:   os.Getenv("GO_ENV") == "production",
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	response.Send(w, http.StatusOK, map[string]string{"token": accessToken})

	return nil
}

func (ah *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	refreshToken, err := r.Cookie("refreshToken")
	if err != nil {
		ah.loggerService.Error("failed to read cookie from request", err.Error())
		return err
	}

	tokenBytes, err := hex.DecodeString(refreshToken.Value)
	if err != nil {
		ah.loggerService.Error("failed to decode string into bytes", err.Error())
		return err
	}

	newAccessToken, err := ah.authService.RefreshToken(ctx, tokenBytes)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			ah.loggerService.Info("request timed out", r.URL.Path)
			return commonErrors.NewAPIError(http.StatusRequestTimeout, "")
		}
		return err
	}

	response.Send(w, http.StatusOK, map[string]string{"token": newAccessToken})

	return nil
}

func (ah *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	r = r.WithContext(ctx)

	r, err := ah.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			ah.loggerService.Info("request timed out", r.URL.Path)
			return commonErrors.NewAPIError(http.StatusRequestTimeout, "")
		}
		return err
	}

	return nil
}

func (ah *AuthHandler) ActivateAccount(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	r = r.WithContext(ctx)

	authToken := request.ReadQueryParam(r, "token")
	r.Header.Set("authenticationMiddleware", "Bearer "+authToken)

	r, err := ah.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			ah.loggerService.Info("request timed out", r.URL.Path)
			return commonErrors.NewAPIError(http.StatusRequestTimeout, "")
		}
		return err
	}

	userID, err := request.ReadAuthorizedUserIDFromToken(r)
	if err != nil {
		return err
	}

	err = ah.authService.ActivateAccount(ctx, userID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			ah.loggerService.Info("request timed out", r.URL.Path)
			return commonErrors.NewAPIError(http.StatusRequestTimeout, "")
		}
		return err
	}

	return nil
}

func (ah *AuthHandler) LogoutUser(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	err := ah.authenticationMiddleware.BlacklistUser(r)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			ah.loggerService.Info("request timed out", r.URL.Path)
			return commonErrors.NewAPIError(http.StatusRequestTimeout, "")
		}
		return err
	}

	refreshToken, err := r.Cookie("refreshToken")
	if err != nil {
		ah.loggerService.Error("failed to read cookie from request", r.URL.Path)
		return err
	}

	tokenBytes, err := hex.DecodeString(refreshToken.Value)
	if err != nil {
		ah.loggerService.Error("failed to decode string into bytes", r.URL.Path)
		return err
	}

	err = ah.authService.LogoutUser(ctx, tokenBytes)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			ah.loggerService.Info("request timed out", r.URL.Path)
			return commonErrors.NewAPIError(http.StatusRequestTimeout, "")
		}
		return err
	}

	return nil
}

func (ah *AuthHandler) LogoutUserFromAllDevices(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	r, err := ah.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			ah.loggerService.Info("request timed out", r.URL.Path)
			return commonErrors.NewAPIError(http.StatusRequestTimeout, "")
		}
		return err
	}

	userID, err := request.ReadAuthorizedUserIDFromToken(r)
	if err != nil {
		return err
	}

	err = ah.authService.LogoutUserFromAllDevices(ctx, userID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			ah.loggerService.Info("request timed out", r.URL.Path)
			return commonErrors.NewAPIError(http.StatusRequestTimeout, "")
		}
		return err
	}

	return nil
}
