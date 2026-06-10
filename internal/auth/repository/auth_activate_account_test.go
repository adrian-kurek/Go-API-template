package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestActivateAccount(t *testing.T) {
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
				mock.ExpectPrepare(regexp.QuoteMeta("UPDATE users SET email_verified = true WHERE id = $1")).
					ExpectExec().
					WithArgs(sqlmock.AnyArg()).
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
				mock.ExpectPrepare(regexp.QuoteMeta("UPDATE users SET email_verified = true WHERE id = $1")).
					WillReturnError(errors.New("failed to prepare sql query"))

				return db, ctx
			},
			wantErr: true,
			err:     errors.New("failed to prepare sql query"),
		},
		{
			title: "execute query failed",
			setupMock: func() (*sql.DB, context.Context) {
				db, mock, _ := sqlmock.New()
				ctx := context.Background()
				mock.ExpectPrepare(regexp.QuoteMeta("UPDATE users SET email_verified = true WHERE id = $1")).
					ExpectExec().
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(errors.New("failed to execute sql query"))

				return db, ctx
			},
			wantErr: true,
			err:     errors.New("failed to execute sql query"),
		},
	}

	for _, testScenario := range testScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			db, ctx := testScenario.setupMock()
			loggerService := setupAuthRepositoryDependencies()
			authRepository := NewAuthRepository(loggerService, db)
			err := authRepository.ActivateAccount(ctx, 1)

			if (err != nil) != testScenario.wantErr {
				t.Errorf("ActivateAccount() error = %v, wantErr = %v", err, testScenario.wantErr)
			}
			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("ActivateAccount() error = %v, scenarioError = %v", err, testScenario.err)
				}
			}
		})
	}
}
