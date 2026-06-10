package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestInsertRefreshToken(t *testing.T) {
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
				mock.ExpectPrepare("INSERT INTO refresh_tokens").
					ExpectExec().
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
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
				mock.ExpectPrepare("INSERT INTO refresh_tokens").
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
				mock.ExpectPrepare("INSERT INTO refresh_tokens").
					ExpectExec().
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
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
			err := authRepository.InsertRefreshToken(ctx, "192.168.0.1", "chrome", "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpbnN0cnVjdG9yTlIiOiItIiwiZXhwIjoxNzc5MDI2ODg4fQ.EMN_WVoAKPq0ocqd9AsAKcXE3RRCpk6erPZhBuiNP68", 1)

			if (err != nil) != testScenario.wantErr {
				t.Errorf("InsertRefreshToken() error = %v, wantErr = %v", err, testScenario.wantErr)
			}

			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("InsertRefreshToken() error = %v, scenarioError = %v", err, testScenario.err)
				}
			}
		})
	}
}

func TestGetRefreshTokenByTokenHash(t *testing.T) {
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
				mock.ExpectPrepare(
					regexp.QuoteMeta(`
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
						rt.token_hash = $1`,
					),
				).
					ExpectQuery().
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(
						sqlmock.NewRows(
							[]string{"id", "user_id", "email", "username", "token_hash", "expires_at"},
						).AddRow(
							1,
							1,
							"",
							"",
							"",
							time.Now().Add(24*time.Hour),
						),
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
				mock.ExpectPrepare(
					regexp.QuoteMeta(`
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
						rt.token_hash = $1`,
					),
				).WillReturnError(errors.New("failed to prepare sql query"))

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
				mock.ExpectPrepare(
					regexp.QuoteMeta(`
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
						rt.token_hash = $1`,
					),
				).ExpectQuery().
					WithArgs(sqlmock.AnyArg()).WillReturnError(errors.New("failed to execute sql query"))

				return db, ctx
			},
			wantErr: true,
			err:     errors.New("failed to execute sql query"),
		},
		{
			title: "token not found",
			setupMock: func() (*sql.DB, context.Context) {
				db, mock, _ := sqlmock.New()
				ctx := context.Background()
				mock.ExpectPrepare(
					regexp.QuoteMeta(`
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
						rt.token_hash = $1`,
					),
				).ExpectQuery().
					WithArgs(sqlmock.AnyArg()).WillReturnError(sql.ErrNoRows)

				return db, ctx
			},
			wantErr: false,
			err:     nil,
		},
	}

	for _, testScenario := range testScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			db, ctx := testScenario.setupMock()
			loggerService := setupAuthRepositoryDependencies()
			authRepository := NewAuthRepository(loggerService, db)
			_, err := authRepository.GetRefreshTokenByTokenHash(ctx, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpbnN0cnVjdG9yTlIiOiItIiwiZXhwIjoxNzc5MDI2ODg4fQ.EMN_WVoAKPq0ocqd9AsAKcXE3RRCpk6erPZhBuiNP68")

			if (err != nil) != testScenario.wantErr {
				t.Errorf("GetRefreshTokenByTokenHash() error = %v, wantErr = %v", err, testScenario.wantErr)
			}

			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("GetRefreshTokenByTokenHash() error = %v, scenarioError = %v", err, testScenario.err)
				}
			}
		})
	}
}

func TestUpdateLastTimeUsedToken(t *testing.T) {
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
				mock.ExpectPrepare(regexp.QuoteMeta("UPDATE refresh_tokens SET last_used_at = $1 WHERE token_hash = $2")).
					ExpectExec().
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
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
				mock.ExpectPrepare(regexp.QuoteMeta("UPDATE refresh_tokens SET last_used_at = $1 WHERE token_hash = $2")).
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
				mock.ExpectPrepare(regexp.QuoteMeta("UPDATE refresh_tokens SET last_used_at = $1 WHERE token_hash = $2")).
					ExpectExec().
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
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
			err := authRepository.UpdateLastTimeUsedToken(ctx, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpbnN0cnVjdG9yTlIiOiItIiwiZXhwIjoxNzc5MDI2ODg4fQ.EMN_WVoAKPq0ocqd9AsAKcXE3RRCpk6erPZhBuiNP68")

			if (err != nil) != testScenario.wantErr {
				t.Errorf("UpdateLastTimeUsedToken() error = %v, wantErr = %v", err, testScenario.wantErr)
			}
			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("UpdateLastTimeUsedToken() error = %v, scenarioError = %v", err, testScenario.err)
				}
			}
		})
	}
}

func TestRemoveTokenFromDB(t *testing.T) {
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
				mock.ExpectPrepare(regexp.QuoteMeta("DELETE FROM refresh_tokens WHERE token_hash = $1")).
					ExpectExec().WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
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
				mock.ExpectPrepare(regexp.QuoteMeta("DELETE FROM refresh_tokens WHERE token_hash = $1")).
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
				mock.ExpectPrepare(regexp.QuoteMeta("DELETE FROM refresh_tokens WHERE token_hash = $1")).
					ExpectExec().WithArgs(sqlmock.AnyArg()).WillReturnError(errors.New("failed to execute sql query"))

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
			err := authRepository.RemoveTokenFromDB(ctx, "fjkenfnewfehuioewoffweof")

			if (err != nil) != testScenario.wantErr {
				t.Errorf("RemoveTokenFromDB() error = %v, wantErr = %v", err, testScenario.wantErr)
			}
			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("RemoveTokenFromDB() error = %v, scenarioError = %v", err, testScenario.err)
				}
			}
		})
	}
}

func TestRemoveTokensFromDBByUserID(t *testing.T) {
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
				mock.ExpectPrepare(regexp.QuoteMeta("DELETE FROM refresh_tokens WHERE user_id = $1")).
					ExpectExec().WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
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
				mock.ExpectPrepare(regexp.QuoteMeta("DELETE FROM refresh_tokens WHERE user_id = $1")).
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
				mock.ExpectPrepare(regexp.QuoteMeta("DELETE FROM refresh_tokens WHERE user_id = $1")).
					ExpectExec().WithArgs(sqlmock.AnyArg()).WillReturnError(errors.New("failed to execute sql query"))

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
			err := authRepository.RemoveTokensFromDBByUserID(ctx, 1)

			if (err != nil) != testScenario.wantErr {
				t.Errorf("RemoveTokensFromDBByUserID() error = %v, wantErr = %v", err, testScenario.wantErr)
			}
			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("RemoveTokensFromDBByUserID() error = %v, scenarioError = %v", err, testScenario.err)
				}
			}
		})
	}
}
