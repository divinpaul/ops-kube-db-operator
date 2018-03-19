package worker

//go:generate mockgen -source=$GOFILE -destination=../mocks/mock_logger.go -package=mocks

import "github.com/golang/glog"

type Logger interface {
	Error(message string)
	Info(message string)
	Fatal(message string)
}

type CustomLogger struct{}

func NewLogger() *CustomLogger {
	return &CustomLogger{}
}
func (c *CustomLogger) Error(message string) {
	glog.Error(message)
}

func (c *CustomLogger) Fatal(message string) {
	glog.Fatal(message)
}

func (c *CustomLogger) Info(message string) {
	glog.Info(message)
}
