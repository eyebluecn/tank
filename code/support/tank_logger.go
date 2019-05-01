package support

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/robfig/cron"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

//在Logger的基础上包装一个全新的Logger.
type TankLogger struct {
	//加锁，在维护日志期间，禁止写入日志。
	sync.RWMutex

	//继承logger
	goLogger *log.Logger
	//日志记录所在的文件
	file *os.File
}

func (this *TankLogger) Init() {

	this.openFile()

	//每日00:00整理日志。
	expression := "0 0 0 * * ?"
	cronJob := cron.New()
	err := cronJob.AddFunc(expression, this.maintain)
	core.PanicError(err)
	cronJob.Start()
	this.Info("[cron job] 每日00:00维护日志")

}

func (this *TankLogger) Destroy() {
	this.closeFile()
}

//处理日志的统一方法。
func (this *TankLogger) Log(prefix string, format string, v ...interface{}) {

	content := fmt.Sprintf(format+"\r\n", v...)

	//控制台中打印日志，记录行号。
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}

	var consoleFormat = fmt.Sprintf("%s%s %s:%d %s", prefix, util.ConvertTimeToTimeString(time.Now()), util.GetFilenameOfPath(file), line, content)
	fmt.Printf(consoleFormat)

	this.goLogger.SetPrefix(prefix)

	//每一行我们加上换行符
	err := this.goLogger.Output(3, content)
	if err != nil {
		fmt.Printf("occur error while logging %s \r\n", err.Error())
	}
}

//处理日志的统一方法。
func (this *TankLogger) Debug(format string, v ...interface{}) {
	this.Log("[DEBUG]", format, v...)
}

func (this *TankLogger) Info(format string, v ...interface{}) {
	this.Log("[INFO ]", format, v...)
}

func (this *TankLogger) Warn(format string, v ...interface{}) {
	this.Log("[WARN ]", format, v...)
}

func (this *TankLogger) Error(format string, v ...interface{}) {
	this.Log("[ERROR]", format, v...)
}

func (this *TankLogger) Panic(format string, v ...interface{}) {
	this.Log("[PANIC]", format, v...)
	panic(fmt.Sprintf(format, v...))
}

//将日志写入到今天的日期中(该方法内必须使用异步方法记录日志，否则会引发死锁)
func (this *TankLogger) maintain() {

	this.Info("每日维护日志")

	this.Lock()
	defer this.Unlock()

	//首先关闭文件。
	this.closeFile()

	//日志归类到昨天
	destPath := util.GetLogPath() + "/tank-" + util.ConvertTimeToDateString(util.Yesterday()) + ".log"

	//直接重命名文件
	err := os.Rename(this.fileName(), destPath)
	if err != nil {
		this.Error("重命名文件出错", err.Error())
	}

	//再次打开文件
	this.openFile()

	//删除一个月之前的日志文件。
	monthAgo := time.Now()
	monthAgo = monthAgo.AddDate(0, -1, 0)
	oldDestPath := util.GetLogPath() + "/tank-" + util.ConvertTimeToDateString(monthAgo) + ".log"
	this.Log("删除日志文件 %s", oldDestPath)

	//删除文件
	exists := util.PathExists(oldDestPath)
	if exists {
		err = os.Remove(oldDestPath)
		if err != nil {
			this.Error("删除磁盘上的文件%s 出错 %s", oldDestPath, err.Error())
		}
	} else {
		this.Error("日志文件 %s 不存在，无需删除", oldDestPath)
	}

}

//日志名称
func (this *TankLogger) fileName() string {
	return util.GetLogPath() + "/tank.log"
}

//打开日志文件
func (this *TankLogger) openFile() {
	//日志输出到文件中 文件打开后暂时不关闭
	fmt.Printf("使用日志文件 %s\r\n", this.fileName())
	f, err := os.OpenFile(this.fileName(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("日志文件无法正常打开: " + err.Error())
	}

	this.goLogger = log.New(f, "", log.Ltime|log.Lshortfile)

	if this.goLogger == nil {
		fmt.Printf("Error: cannot create goLogger \r\n")
	}

	this.file = f
}

//关闭日志文件
func (this *TankLogger) closeFile() {
	if this.file != nil {
		err := this.file.Close()
		if err != nil {
			panic("尝试关闭日志时出错: " + err.Error())
		}
	}
}
