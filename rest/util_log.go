package rest

import (
	"fmt"
	"log"
	"os"
	"time"
)

func Log(prefix string, content string) {

	//日志输出到文件中
	filePath := GetLogPath() + "/tank-" + time.Now().Local().Format("2006-01-02") + ".log"
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.SetPrefix(prefix)
	log.Println(content)

	//如果需要输出到控制台。
	if CONFIG.LogToConsole {
		fmt.Println(content)
	}

}

func LogDebug(content string) {
	go Log("[Debug]", content)
}

func LogInfo(content string) {
	go Log("[Info]", content)
}

func LogWarning(content string) {
	go Log("[Warning]", content)
}

func LogError(content string) {
	go Log("[Error]", content)
}

func LogPanic(content interface{}) {
	Log("[Panic]", fmt.Sprintf("%v", content))
	panic(content)
}
