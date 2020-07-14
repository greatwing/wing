package logger

import (
	_ "github.com/davyxu/cellnet/peer/tcp"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/davyxu/golog"
	"github.com/greatwing/wing/base/config"
	"strings"
)

type gologAdapter struct{}

func (g gologAdapter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSuffix(string(p), "\n")
	logger.Info(msg)
	return 0, nil
}

func init() {
	golog.VisitLogger(`\S*`, func(l *golog.Logger) bool {
		//fmt.Println(l.Name())
		l.SetOutptut(gologAdapter{})
		l.SetParts()

		if config.Debug() {
			l.SetLevel(golog.Level_Debug)
		} else {
			l.SetLevel(golog.Level_Info)
		}

		return true
	})
}
