package core

type Logger interface {

	//basic log method
	Log(prefix string, format string, v ...interface{})

	//log with different level.
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Panic(format string, v ...interface{})
}
