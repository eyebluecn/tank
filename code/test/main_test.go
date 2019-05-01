package test

import (
	"fmt"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/robfig/cron"
	"log"
	"strings"
	"testing"
	"time"
)

func TestHello(t *testing.T) {

	split := strings.Split("good", "/")
	fmt.Printf("%v", split)

}

//测试cron表达式
func TestCron(t *testing.T) {

	i := 0
	c := cron.New()
	spec := "*/1 * * * * ?"
	err := c.AddFunc(spec, func() {
		i++
		log.Println("cron running:", i)
		if i == 2 {
			panic("intent to panic.")
		}
	})
	core.PanicError(err)

	c.Start()

	//当前线程阻塞 20s
	time.Sleep(3 * time.Second)

}

//测试 时间
func TestDayAgo(t *testing.T) {

	dayAgo := time.Now()
	dayAgo = dayAgo.AddDate(0, 0, -8)

	thenDay := util.FirstSecondOfDay(dayAgo)

	fmt.Printf("%s\n", util.ConvertTimeToDateTimeString(thenDay))

}

//测试 打包
func TestZip(t *testing.T) {

	util.Zip("/Users/fusu/d/group/eyeblue/tank/tmp/matter/admin/root/morning", "/Users/fusu/d/group/eyeblue/tank/tmp/log/morning.zip")

}
