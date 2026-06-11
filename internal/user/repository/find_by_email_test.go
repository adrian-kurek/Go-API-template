package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)
func TestFindByEmail(t *testing.T) {
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
				mock.ExpectPrepare(regexp.QuoteMeta("SELECT id,email, username,password,email_verified,created_at,updated_at FROM USERS WHERE email = $1")).
				ExpectQuery().
				WithArgs(sqlmock.AnyArg()).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "email", "username", "password", "email_verified", "created_at", "updated_at"}).
						AddRow(1, "test@example.com", "testuser", "hashedpass", true, time.Now(), time.Now()),
				)
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
				mock.ExpectPrepare(regexp.QuoteMeta("SELECT id,email, username,password,email_verified,created_at,updated_at FROM USERS WHERE email = $1")).
					WillReturnError(errors.New("failed to prepare sql query"))
				return db, ctx
			},
			wantErr: true,
			err:     errors.New("failed to prepare sql query"),
		},
		{
			title: "failed to execute",
			setupMock: func() (*sql.DB, context.Context) {
				db, mock, _ := sqlmock.New()
				ctx := context.Background()
				mock.ExpectPrepare(regexp.QuoteMeta("SELECT id,email, username,password,email_verified,created_at,updated_at FROM USERS WHERE email = $1")).
					ExpectQuery().WithArgs(sqlmock.AnyArg()).
					WillReturnError(&pq.Error{Code: "23505", Message: "failed to execute the query"})
				return db, ctx
			},
			wantErr: true,
			err:     &pq.Error{Code: "23505", Message: "failed to execute the query"},
		},
		{
			title: "user not found",
			setupMock: func() (*sql.DB, context.Context) {
				db, mock, _ := sqlmock.New()
				ctx := context.Background()
				mock.ExpectPrepare(regexp.QuoteMeta("SELECT id,email, username,password,email_verified,created_at,updated_at FROM USERS WHERE email = $1")).
					ExpectQuery().WithArgs(sqlmock.AnyArg()).
				WillReturnError(sql.ErrNoRows)				
				return db, ctx
			},
			wantErr: false,
			err:     nil,
		},
		{
			title: "context cancelled",
			setupMock: func() (*sql.DB, context.Context) {
				db, mock, _ := sqlmock.New()
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				mock.ExpectPrepare(regexp.QuoteMeta("SELECT id,email, username,password,email_verified,created_at,updated_at FROM USERS WHERE email = $1")).
					ExpectQuery().WithArgs(sqlmock.AnyArg()).
					WillReturnError(context.Canceled)
				return db, ctx
			},
			wantErr: true,
			err:     context.Canceled,
		},
	}

	for _, testScenario := range testScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			loggerService := setupUserRepositoryDependencies()
			db, ctx := testScenario.setupMock()
			authRepository := NewUserRepository(loggerService, db)
			_,err := authRepository.FindByEmail(ctx, "joedoe@gmail.com")

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