package pkglog

import "gopkg.in/natefinch/lumberjack.v2"

type Config struct {
	*lumberjack.Logger
	Format string `json:"format" yaml:"format"`
}
