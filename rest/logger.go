package rest

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

//日志系统必须高保
//全局唯一的日志对象(在main函数中初始化)
var LOGGER = &Logger{}

//在Logger的基础上包装一个全新的Logger.
type Logger struct {
	//加锁，在维护日志期间，禁止写入日志。
	sync.RWMutex

	//继承logger
	goLogger *log.Logger
	//日志记录所在的文件
	file *os.File
	//每天凌晨定时整理器
	maintainTimer *time.Timer
}

//处理日志的统一方法。
func (this *Logger) log(prefix string, format string, v ...interface{}) {

	//控制台中打印日志
	var consoleFormat = fmt.Sprintf("%s%s %s\r\n", prefix, ConvertTimeToTimeString(time.Now()), format)
	fmt.Printf(consoleFormat, v...)

	this.goLogger.SetPrefix(prefix)
	//每一行我们加上换行符
	var fileFormat = fmt.Sprintf("%s\r\n", format)
	this.goLogger.Printf(fileFormat, v...)
}

//处理日志的统一方法。
func (this *Logger) Debug(format string, v ...interface{}) {
	this.log("[DEBUG]", format, v...)
}

func (this *Logger) Info(format string, v ...interface{}) {
	this.log("[INFO ]", format, v...)
}

func (this *Logger) Warn(format string, v ...interface{}) {
	this.log("[WARN ]", format, v...)
}

func (this *Logger) Error(format string, v ...interface{}) {
	this.log("[ERROR]", format, v...)
}

func (this *Logger) Panic(format string, v ...interface{}) {
	this.log("[PANIC]", format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (this *Logger) Init() {

	this.openFile()

	//日志需要自我备份，自我维护。明天第一秒触发
	nextTime := FirstSecondOfDay(Tomorrow())
	duration := nextTime.Sub(time.Now())

	this.Info("下一次日志维护时间%s 距当前 %ds ", ConvertTimeToDateTimeString(nextTime), duration/time.Second)
	this.maintainTimer = time.AfterFunc(duration, func() {
		go SafeMethod(this.maintain)
	})

}

//将日志写入到今天的日期中(该方法内必须使用异步方法记录日志，否则会引发死锁)
func (this *Logger) maintain() {

	this.Info("每日维护日志")

	this.Lock()
	defer this.Unlock()

	//首先关闭文件。
	this.closeFile()

	//日志归类到昨天
	destPath := GetLogPath() + "/tank-" + Yesterday().Local().Format("2006-01-02") + ".log"

	//直接重命名文件
	err := os.Rename(this.fileName(), destPath)
	if err != nil {
		this.Error("重命名文件出错", err.Error())
	}

	//再次打开文件
	this.openFile()

	//准备好下次维护日志的时间。
	now := time.Now()
	nextTime := FirstSecondOfDay(Tomorrow())
	duration := nextTime.Sub(now)
	this.Info("下次维护时间：%s ", ConvertTimeToDateTimeString(nextTime))
	this.maintainTimer = time.AfterFunc(duration, func() {
		go SafeMethod(this.maintain)
	})
}

//日志名称
func (this *Logger) fileName() string {
	return GetLogPath() + "/tank.log"
}

//打开日志文件
func (this *Logger) openFile() {
	//日志输出到文件中 文件打开后暂时不关闭
	fmt.Printf("使用日志文件 %s\r\n", this.fileName())
	f, err := os.OpenFile(this.fileName(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("日志文件无法正常打开: " + err.Error())
	}

	this.goLogger = log.New(f, "", log.Ltime)
	this.file = f
}

//关闭日志文件
func (this *Logger) closeFile() {
	if this.file != nil {
		err := this.file.Close()
		if err != nil {
			panic("尝试关闭日志时出错: " + err.Error())
		}
	}
}

func (this *Logger) Destroy() {

	this.closeFile()

	if this.maintainTimer != nil {
		this.maintainTimer.Stop()
	}

}
