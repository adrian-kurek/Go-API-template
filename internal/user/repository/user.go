package repository

import (
	"context"
	"database/sql"
	"errors"

	commonErrors "github.com/slodkiadrianek/Go-API-template/common/errors"
	commonInterfaces "github.com/slodkiadrianek/Go-API-template/common/interfaces"
	authDTO "github.com/slodkiadrianek/Go-API-template/internal/auth/DTO"
	"github.com/slodkiadrianek/Go-API-template/internal/user/model"
)

type UserRepository struct {
	loggerService commonInterfaces.Logger
	db            *sql.DB
}

func NewUserRepository(loggerService commonInterfaces.Logger, db *sql.DB) *UserRepository {
	return &UserRepository{
		loggerService: loggerService,
		db:            db,
	}
}

func (ur *UserRepository) Create(ctx context.Context, user authDTO.CreateUser, hashedPassword []byte) error {
	query := `INSERT INTO users (email,username,password,created_at,updated_at) VALUES ($1,$2,$3,$4,now(),now())`

	stmt, err := ur.db.PrepareContext(ctx, query)
	if err != nil {
		ur.loggerService.Error(commonErrors.FailedToPrepareQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ur.loggerService.Error(commonErrors.FailedToCloseStatement, closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, user.Email, user.Username, hashedPassword)
	if err != nil {
		ur.loggerService.Error(commonErrors.FailedToExecuteInsertQuery, map[string]any{
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

func (ur *UserRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	query := "SELECT id,email, username,password,email_verified,created_at,updated_at FROM USERS WHERE email = $1"
	stmt, err := ur.db.PrepareContext(ctx, query)
	if err != nil {
		ur.loggerService.Error(commonErrors.FailedToPrepareQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return model.User{}, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ur.loggerService.Error(commonErrors.FailedToCloseStatement, closeErr)
		}
	}()

	var user model.User
	err = stmt.QueryRowContext(ctx, email).Scan(&user.ID, &user.Email, &user.Username, &user.Password, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ur.loggerService.Info("user not found", map[string]any{
				"email": email,
			})
			return model.User{
				ID: 0,
			}, nil
		}
		ur.loggerService.Error(commonErrors.FailedToExecuteSelectQuery, map[string]any{
			"query": query,
			"args":  []any{email},
			"error": err,
		})
		return model.User{}, err
	}
	return user, nil
}
