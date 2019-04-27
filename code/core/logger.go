package core

type Logger interface {
	//处理日志的统一方法。
	Log(prefix string, format string, v ...interface{})

	//不同级别的日志处理
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Panic(format string, v ...interface{})
}
