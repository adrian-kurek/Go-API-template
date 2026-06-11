package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	authDTO "github.com/slodkiadrianek/Go-API-template/internal/auth/DTO"
)

func TestRegister(t *testing.T) {
	type args struct {
		title     string
		setupMock func() (*sql.DB, context.Context)
		wantErr   bool
		err       error
	}

	testScenarios := []args{
		{
			title: "with proper data",
			setupMock: func() (*sql.DB, context.Context) {
				db, mock, _ := sqlmock.New()
				ctx := context.Background()
				mock.ExpectPrepare("INSERT INTO users").
					ExpectExec().
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return db, ctx
			},
			wantErr: false,
			err:     nil,
		},
		{
			title: "prepare query failed",
			setupMock: func() (*sql.DB, context.Context) {
				db, mock, _ := sqlmock.New()
				ctx := context.Background()
				mock.ExpectPrepare("INSERT INTO users").
					WillReturnError(errors.New("failed to prepare sql query"))
				return db, ctx
			},
			wantErr: true,
			err:     errors.New("failed to prepare sql query"),
		},
		{
			title: "duplicate email",
			setupMock: func() (*sql.DB, context.Context) {
				db, mock, _ := sqlmock.New()
				ctx := context.Background()
				mock.ExpectPrepare("INSERT INTO users").
					ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(&pq.Error{Code: "23505", Message: "duplicate key value"})
				return db, ctx
			},
			wantErr: true,
			err:     &pq.Error{Code: "23505", Message: "duplicate key value"},
		},
		{
			title: "context cancelled",
			setupMock: func() (*sql.DB, context.Context) {
				db, mock, _ := sqlmock.New()
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				mock.ExpectPrepare("INSERT INTO users").
					ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(context.Canceled)
				return db, ctx
			},
			wantErr: true,
			err:     context.Canceled,
		},
	}

	for _, testScenario := range testScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			loggerService := setupAuthRepositoryDependencies()
			user := authDTO.CreateUser{
				Username:        "joeDoe",
				Email:           "joedoe@gmail.com",
				Password:        "zaq1@#$rfvbgt5",
				ConfirmPassword: "zaq1@#$rfvbgt5",
			}
			db, ctx := testScenario.setupMock()
			authRepository := NewAuthRepository(loggerService, db)
			err := authRepository.RegisterUser(ctx, user, []byte("test"))

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
