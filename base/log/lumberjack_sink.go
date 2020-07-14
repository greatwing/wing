package logger

import (
	"fmt"
	"github.com/greatwing/wing/base/config"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/url"
	"os"
	"path"
	"strings"
)

var rotateLogger lumberjack.Logger

type lumberjackSink struct {
	*lumberjack.Logger
}

func (l *lumberjackSink) Sync() error {
	return nil
}

func registerLumberjack(u *url.URL) (zap.Sink, error) {
	return &lumberjackSink{&rotateLogger}, nil
}

func GetProcName() string {
	procName := strings.ReplaceAll(os.Args[0], "\\", "/")
	procName = path.Base(procName)
	ext := path.Ext(procName)
	if len(ext) > 0 {
		procName = strings.TrimSuffix(procName, ext)
	}
	return procName
}

func Rotate() {
	rotateLogger.Rotate()
}

func init() {
	logPath := fmt.Sprintf("log/%s/%d/%s", config.GetSvcGroup(), config.GetSvcIndex(), GetProcName()+".log")
	rotateLogger = lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    1024,
		MaxBackups: 0,
		MaxAge:     0,
		Compress:   false,
	}
}
