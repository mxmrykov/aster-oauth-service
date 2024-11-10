package logger

import (
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
	"strconv"
)

func NewLogger(useStackTrace bool) *zerolog.Logger {
	if useStackTrace {
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			return filepath.Base(file) + ":" + strconv.Itoa(line)
		}
	}

	l := zerolog.New(os.Stdout).With().Caller().Timestamp().Logger()

	return &l
}
