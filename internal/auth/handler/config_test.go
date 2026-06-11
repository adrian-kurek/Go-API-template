package handler

import (
	"os"
	"testing"

	"github.com/slodkiadrianek/Go-API-template/common/log"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.RemoveAll("logs")
	os.Exit(code)
}

func setupAuthHandlerDependencies() *log.Logger {
	loggerService := log.NewLogger("./logs", "2006-01-02", "15:04:05")
	
	return loggerService
}
