package test

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
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

	time.Sleep(3 * time.Second)

}

func TestDayAgo(t *testing.T) {

	dayAgo := time.Now()
	dayAgo = dayAgo.AddDate(0, 0, -8)

	thenDay := util.FirstSecondOfDay(dayAgo)

	fmt.Printf("%s\n", util.ConvertTimeToDateTimeString(thenDay))

}
