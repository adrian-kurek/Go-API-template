package request

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	commonErrors "github.com/slodkiadrianek/Go-API-template/common/errors"
	"github.com/slodkiadrianek/Go-API-template/common/response"
)

const (
	green = "\x1b[32m"
	reset = "\x1b[0m"
)

type HTTPFunc func(w http.ResponseWriter, r *http.Request) error

func Make(f HTTPFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		actualDate := time.Now()
		logTime := actualDate.Format("2006-01-02 15:04:05")
		if err := f(w, r); err != nil {
			if apiErr, ok := err.(*commonErrors.APIError); ok {
				response.Send(w, apiErr.StatusCode, map[string]string{"message": apiErr.Message})
			} else {
				fmt.Println(err.Error())
				response.Send(w, 500, map[string]string{"message": "Internal server error"})
			}
		}
		durationOfTheRoute := time.Since(start) / time.Millisecond
		formattedDurationOfTheRoute := strconv.FormatInt(int64(durationOfTheRoute), 10) + "ms"

		fmt.Println(green + "[INFO: " + logTime + "] " + r.Method + "-" + r.URL.Path + "-" + r.
			RemoteAddr + "-" + formattedDurationOfTheRoute + reset)
	}
}

func ReadAuthorizedUserIDFromToken(r *http.Request) (int, error) {
	userID, ok := r.Context().Value("id").(int)
	if !ok || userID == 0 {
		err := errors.New("failed to read user from context")
		return 0, err
	}

	return userID, nil
}

func ReadBody[T any](r *http.Request) (*T, error) {
	if r.Body == nil {
		return nil, errors.New("no request body provided")
	}
	var body T

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func ReadQueryParam(r *http.Request, QueryName string) string {
	name := r.URL.Query().Get(QueryName)
	return name
}

func SetContext(r *http.Request, key, data any) *http.Request {
	ctx := context.WithValue(r.Context(), key, data)
	return r.WithContext(ctx)
}
