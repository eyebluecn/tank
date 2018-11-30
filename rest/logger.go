package rest

import (
	"fmt"
	"log"
	"os"
)

//日志系统必须高保
//全局唯一的日志对象(在main函数中初始化)
var LOGGER *Logger = &Logger{}

//在Logger的基础上包装一个全新的Logger.
type Logger struct {
	//继承logger
	goLogger *log.Logger
	//日志记录所在的文件
	file *os.File
}

//处理日志的统一方法。
func (this *Logger) log(prefix string, format string, v ...interface{}) {
	fmt.Printf(format+"\r\n", v...)

	this.goLogger.SetPrefix(prefix)
	this.goLogger.Printf(format, v...)
}

//处理日志的统一方法。
func (this *Logger) Debug(format string, v ...interface{}) {
	this.log("[debug]", format, v...)
}

func (this *Logger) Info(format string, v ...interface{}) {
	this.log("[info]", format, v...)
}

func (this *Logger) Warn(format string, v ...interface{}) {
	this.log("[warn]", format, v...)
}

func (this *Logger) Error(format string, v ...interface{}) {
	this.log("[error]", format, v...)
}

func (this *Logger) Panic(format string, v ...interface{}) {
	this.log("[panic]", format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (this *Logger) Init() {

	//日志输出到文件中 文件打开后暂时不关闭
	filePath := GetLogPath() + "/tank.log"
	fmt.Printf("使用日志文件 %s\r\n", filePath)
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("日志文件无法正常打开: " + err.Error())
	}

	this.goLogger = log.New(f, "", log.Ltime)
	this.file = f
}

func (this *Logger) Destroy() {
	if this.file != nil {
		err := this.file.Close()
		if err != nil {
			panic("尝试关闭日志时出错: " + err.Error())
		}
	}

}
