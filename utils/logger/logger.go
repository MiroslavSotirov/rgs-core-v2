package logger

const (
	//Debug has verbose message
	Debug = "debug"
	//Info is default log level
	Info = "info"
	//Warn is for logging messages about possible issues
	Warn = "warn"
	//Error is for logging errors
	Error = "error"
	//Fatal is for logging fatal messages. The sytem shutsdown after logging the message.
	Fatal = "fatal"
)

var log Logger

//Fields Type to pass when we want to call WithFields for structured logging
type Fields map[string]interface{}

//Logger is our contract for the logger
type Logger interface {
	//Debug(format string)
	//
	//Info(format string)
	//
	//Warn(format string)
	//
	//Error(format string)
	//
	//Fatal(format string)
	//
	//Panic(format string)

	Debugf(format string, args ...interface{})

	Infof(format string, args ...interface{})

	Warnf(format string, args ...interface{})

	Errorf(format string, args ...interface{})

	Fatalf(format string, args ...interface{})

	Panicf(format string, args ...interface{})

	WithFields(keyValues Fields) Logger
}

// Configuration stores the config for the logger
// For some loggers there can only be one level across writers, for such the level of Console is picked by default
type Configuration struct {
	ConsoleJSONFormat bool
	ConsoleLevel      string
}

//NewLogger returns an instance of logger
func NewLogger(config Configuration) error {
	logger, err := newZapLogger(config)
	if err != nil {
		return err
	}
	log = logger
	return nil
}

//func Debug(format string) {
//	log.Debug(format)
//}
//
//func Info(format string) {
//	log.Info(format)
//}
//
//func Warn(format string) {
//	log.Warn(format)
//}
//
//func Error(format string) {
//	log.Errorf(format)
//}
//
//func Fatal(format string) {
//	log.Fatal(format)
//}
//
//func Panic(format string) {
//	log.Panicf(format)
//}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	log.Panicf(format, args...)
}

func WithFields(keyValues Fields) Logger {
	return log.WithFields(keyValues)
}
