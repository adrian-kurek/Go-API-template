package middleware

import (
	"errors"
	"testing"
)

type testStructRequired struct {
	Name            string `validate:"required"`
	Age             int    `validate:"min=18"`
	MaxAge          int    `validate:"max=65"`
	Email           string `validate:"email"`
	Password        string `validate:"eqfield=ConfirmPassword"`
	ConfirmPassword string
}

func TestValidateRequestData(t *testing.T) {
	type args struct {
		title   string
		schema  any
		wantErr bool
		err     error
	}

	testsScenarios := []args{
		{
			title: "with proper data",
			schema: testStructRequired{
				Name:            "jode de",
				Age:             18,
				MaxAge:          64,
				Email:           "joedoe@gmail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErr: false,
			err:     nil,
		},
		{
			title: "required field is missing",
			schema: testStructRequired{
				Age:             16,
				MaxAge:          64,
				Email:           "joedoe@gmail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErr: true,
			err:     errors.New("api error: the Name field is required"),
		},
		{
			title: "min is too low",
			schema: testStructRequired{
				Name:            "joe doe",
				Age:             16,
				MaxAge:          64,
				Email:           "joedoe@gmail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErr: true,
			err:     errors.New("api error: the Age field must be at least 18 characters long"),
		},
		{
			title: "max is too high",
			schema: testStructRequired{
				Name:            "joe doe",
				Age:             19,
				MaxAge:          69,
				Email:           "joedoe@gmail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErr: true,
			err:     errors.New("api error: the MaxAge field must be at most 65 characters long"),
		},
		{
			title: "incorrect email format",
			schema: testStructRequired{
				Name:            "joe doe",
				Age:             19,
				MaxAge:          65,
				Email:           "joedoegmail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErr: true,
			err:     errors.New("api error: the Email field must be a valid email address"),
		},
		{
			title: "password is not the same as confirm password",
			schema: testStructRequired{
				Name:            "joe doe",
				Age:             19,
				MaxAge:          65,
				Email:           "joedoe@gmail.com",
				Password:        "password123",
				ConfirmPassword: "password12",
			},
			wantErr: true,
			err:     errors.New("api error: the Password field must be the same as ConfirmPassword field"),
		},
	}

	for _, testScenario := range testsScenarios {
		t.Run(testScenario.title, func(t *testing.T) {
			err := ValidateRequestData(testScenario.schema)
			if (err != nil) != testScenario.wantErr {
				t.Errorf("ValidateRequestData() error = %v, wantErr %v", err, testScenario.wantErr)
			}
			if err != nil && testScenario.err != nil {
				if err.Error() != testScenario.err.Error() {
					t.Errorf("ValidateRequestData() error = %v, expected error %v", err, testScenario.err)
				}
			}
		})
	}
}
