package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	commonErrors "github.com/slodkiadrianek/Go-API-template/common/errors"
	commonInterfaces "github.com/slodkiadrianek/Go-API-template/common/interfaces"
	authDTO "github.com/slodkiadrianek/Go-API-template/internal/auth/DTO"
	authModel "github.com/slodkiadrianek/Go-API-template/internal/auth/model"
)

const refreshTokenExpiration = 7 * 24 * time.Hour

type AuthRepository struct {
	loggerService commonInterfaces.Logger
	db            *sql.DB
}

func NewAuthRepository(loggerService commonInterfaces.Logger, db *sql.DB) *AuthRepository {
	return &AuthRepository{
		loggerService: loggerService,
		db:            db,
	}
}

func (ar *AuthRepository) RegisterUser(ctx context.Context, user authDTO.CreateUser, hashedPassword []byte) error {
	query := `INSERT INTO users (email,username,password,created_at) VALUES ($1,$2,$3,$4)`

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToPrepareQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.loggerService.Error(commonErrors.FailedToCloseStatement, closeErr)
		}
	}()

	timestamp := time.Now()

	_, err = stmt.ExecContext(ctx, user.Email, user.Username, hashedPassword, timestamp)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToExecuteInsertQuery, map[string]any{
			"query": query,
			"args": map[string]string{
				"username": user.Username,
				"email":    user.Email,
			},
			"error": err,
		})
		return err
	}

	return nil
}

func (ar *AuthRepository) InsertRefreshToken(ctx context.Context, ipAddress, deviceInfo, refreshToken string, userID int) error {
	query := `INSERT INTO refresh_tokens(user_id,token_hash,device_info,ip_address, expires_at, last_used_at) VALUES($1,$2,$3,$4,$5,$6)`

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToPrepareQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.loggerService.Error(commonErrors.FailedToCloseStatement, closeErr)
		}
	}()

	expiration := time.Now().Add(refreshTokenExpiration)
	lastUsedAt := time.Now()

	_, err = stmt.ExecContext(ctx, userID, refreshToken, deviceInfo, ipAddress, expiration, lastUsedAt)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToExecuteInsertQuery, map[string]any{
			"query": query,
			"args": map[string]any{
				"userId":     userID,
				"deviceInfo": deviceInfo,
				"ipAddress":  ipAddress,
			},
			"error": err,
		})
		return err
	}

	return nil
}

func (ar *AuthRepository) GetRefreshTokenByTokenHash(ctx context.Context, refreshToken string) (authModel.TokenWithUserEmailToRefreshToken, error) {
	query := `
	SELECT 
		rt.id,
		rt.user_id,
		u.email,
		u.username,
		rt.token_hash,
		rt.expires_at 
	FROM refresh_tokens rt
		INNER JOIN users u ON u.id = rt.user_id
	WHERE
		rt.token_hash = $1`

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToPrepareQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return authModel.TokenWithUserEmailToRefreshToken{}, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.loggerService.Error(commonErrors.FailedToCloseStatement, closeErr)
		}
	}()

	row := stmt.QueryRowContext(ctx, refreshToken)

	var tokenWithUserEmailToRefreshToken authModel.TokenWithUserEmailToRefreshToken

	err = row.Scan(&tokenWithUserEmailToRefreshToken.ID, &tokenWithUserEmailToRefreshToken.UserID, &tokenWithUserEmailToRefreshToken.Email, &tokenWithUserEmailToRefreshToken.Username, &tokenWithUserEmailToRefreshToken.TokenHash, &tokenWithUserEmailToRefreshToken.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ar.loggerService.Info("token not found", nil)
			return authModel.TokenWithUserEmailToRefreshToken{ID: 0}, nil
		}
		ar.loggerService.Error(commonErrors.FailedToExecuteSelectQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return authModel.TokenWithUserEmailToRefreshToken{}, err
	}
	return tokenWithUserEmailToRefreshToken, nil
}

func (ar *AuthRepository) UpdateLastTimeUsedToken(ctx context.Context, refreshToken string) error {
	query := `
	UPDATE refresh_tokens SET last_used_at = $1 WHERE token_hash = $2	
	`

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToPrepareQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.loggerService.Error(commonErrors.FailedToCloseStatement, closeErr)
		}
	}()

	lastUsedAt := time.Now()

	_, err = stmt.ExecContext(ctx, lastUsedAt, refreshToken)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToExecuteUpdateQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}

	return nil
}

func (ar *AuthRepository) RemoveTokenFromDB(ctx context.Context, refreshToken string) error {
	query := "DELETE FROM refresh_tokens WHERE token_hash = $1"

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToPrepareQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.loggerService.Error(commonErrors.FailedToCloseStatement, closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, refreshToken)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToExecuteDeleteQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}

	return nil
}

func (ar *AuthRepository) ActivateAccount(ctx context.Context, userID int) error {
	query := "UPDATE users SET email_verified = true WHERE id = $1"

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToPrepareQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.loggerService.Error(commonErrors.FailedToCloseStatement, closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, userID)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToExecuteUpdateQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}

	return nil
}

func (ar *AuthRepository) RemoveTokensFromDBByUserID(ctx context.Context, userID int) error {
	query := "DELETE FROM refresh_tokens WHERE user_id = $1"

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToPrepareQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.loggerService.Error(commonErrors.FailedToCloseStatement, closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, userID)
	if err != nil {
		ar.loggerService.Error(commonErrors.FailedToExecuteDeleteQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}

	return nil
}
