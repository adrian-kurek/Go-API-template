package middleware

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	commonErrors "github.com/slodkiadrianek/Go-API-template/common/errors"
	commonInterfaces "github.com/slodkiadrianek/Go-API-template/common/interfaces"
	"github.com/slodkiadrianek/Go-API-template/common/request"
	userModel "github.com/slodkiadrianek/Go-API-template/internal/user/model"
)

const accessTokenExpiration = 5 * time.Minute

type AuthenticationMiddleware struct {
	accessTokenSecret  string
	refreshTokenSecret string
	loggerService      commonInterfaces.Logger
	cacheService       commonInterfaces.CacheService
}

func NewAuthenticationMiddleware(accessTokenSecret string, refreshTokenSecret string, loggerService commonInterfaces.Logger, cacheService commonInterfaces.CacheService) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		accessTokenSecret:  accessTokenSecret,
		refreshTokenSecret: refreshTokenSecret,
		loggerService:      loggerService,
		cacheService:       cacheService,
	}
}

func (am *AuthenticationMiddleware) GenerateRefreshToken() ([]byte, error) {
	bytes := make([]byte, 64)

	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (am *AuthenticationMiddleware) HashToken(token []byte) string {
	hasher := sha256.New()
	hasher.Write(token)
	return hex.EncodeToString(hasher.Sum(nil))
}

func (am *AuthenticationMiddleware) GenerateAccessToken(user userModel.User) (string, error) {
	am.loggerService.Info("started signing a new access token", nil)
	tokenWithData := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
		"exp":      time.Now().Add(accessTokenExpiration).Unix(),
	})

	tokenString, err := tokenWithData.SignedString([]byte(am.accessTokenSecret))
	if err != nil {
		errMsg := errors.New("failed to sign access token properly")
		am.loggerService.Error(errMsg.Error(), err.Error())
		return "", commonErrors.NewAPIError(http.StatusUnauthorized, errMsg.Error())
	}

	am.loggerService.Info("Successfully signed a new access token", nil)
	return tokenString, nil
}

func (am *AuthenticationMiddleware) parseClaimsFromToken(tokenString string) (*jwt.Token, userModel.UserClaims, error) {
	var user userModel.UserClaims
	token, err := jwt.ParseWithClaims(tokenString, &user, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(am.accessTokenSecret), nil
	})
	if err != nil {
		return nil, userModel.UserClaims{}, err
	}
	return token, user, nil
}

func (am *AuthenticationMiddleware) readTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		am.loggerService.Info("token is missing", authHeader)
		return "", commonErrors.NewAPIError(http.StatusUnauthorized, "failed to authorize a user")
	}

	return strings.Split(authHeader, " ")[1], nil
}

func (am *AuthenticationMiddleware) isTokenBlackListed(ctx context.Context, token string) error {
	cacheKey := "tokenBlackList-" + token
	result, err := am.cacheService.Exists(ctx, cacheKey)
	if err != nil {
		am.loggerService.Info("failed to check blacklist", err)
		return err
	}

	if result > 0 {
		err := errors.New("token blacklisted")
		am.loggerService.Info(err.Error(), token)
		return commonErrors.NewAPIError(http.StatusUnauthorized, err.Error())
	}
	return nil
}

func (am *AuthenticationMiddleware) checkIsTokenValid(token *jwt.Token) error {
	if !token.Valid {
		err := errors.New("provided token is invalid")
		am.loggerService.Info(err.Error(), token)
		return commonErrors.NewAPIError(http.StatusUnauthorized, err.Error())
	}
	return nil
}

func (am *AuthenticationMiddleware) VerifyToken(r *http.Request) (*http.Request, error) {
	ctx := r.Context()
	if err := ctx.Err(); err != nil {
		return r, err
	}

	token, err := am.readTokenFromRequest(r)
	if err != nil {
		return r, err
	}

	err = am.isTokenBlackListed(ctx, token)
	if err != nil {
		return r, err
	}

	if err := ctx.Err(); err != nil {
		return r, err
	}

	tokenWithData, user, err := am.parseClaimsFromToken(token)
	if err != nil {
		am.loggerService.Info("failed to read data properly", err.Error())
		return r, commonErrors.NewAPIError(http.StatusUnauthorized, "provided token is invalid")
	}

	if err := ctx.Err(); err != nil {
		return r, err
	}

	err = am.checkIsTokenValid(tokenWithData)
	if err != nil {
		return r, err
	}

	r = request.SetContext(r, "id", user.ID)
	r = request.SetContext(r, "email", user.Email)
	r = request.SetContext(r, "username", user.Username)
	return r, nil
}

func (am *AuthenticationMiddleware) BlacklistUser(r *http.Request) error {
	ctx := r.Context()

	token, err := am.readTokenFromRequest(r)
	if err != nil {
		return err
	}

	tokenWithData, user, err := am.parseClaimsFromToken(token)
	if err != nil {
		am.loggerService.Info("failed to read data properly", token)
		return commonErrors.NewAPIError(http.StatusUnauthorized, "failed to read token")
	}

	err = am.checkIsTokenValid(tokenWithData)
	if err != nil {
		return err
	}
	err = am.isTokenBlackListed(ctx, token)
	if err != nil {
		return err
	}

	expirationTime := time.Until(user.ExpiresAt.Time)

	cacheKey := "tokenBlackList-" + token
	err = am.cacheService.Set(ctx, cacheKey, "true", expirationTime)
	if err != nil {
		am.loggerService.Info("failed to set data in cache", err)
		return err
	}

	return nil
}
