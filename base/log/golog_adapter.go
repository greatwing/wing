package log

import (
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/golog"
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
		return true
	})
}
