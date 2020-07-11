package test

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/robfig/cron/v3"
	"log"
	"testing"
	"time"
)

func TestEveryOneSecondCron(t *testing.T) {

	i := 0
	customMethod := true

	var c *cron.Cron
	var spec string
	if customMethod {
		//use custom cron. every 1 second.
		c = cron.New(cron.WithSeconds())
		spec = "*/1 * * * * ?"
	} else {
		c = cron.New()
		spec = "@every 1s"
	}

	entryId, err := c.AddFunc(spec, func() {
		i++
		log.Println("cron running:", i)
		if i == 5 {
			return
		}
	})
	fmt.Printf("entryId = %d\n", entryId)
	core.PanicError(err)

	c.Start()

	time.Sleep(3500 * time.Millisecond)
	if i != 3 {
		t.Errorf("should be 3\n")
	}
	fmt.Printf("i = %d", i)

}

func TestSimpleCron(t *testing.T) {

	i := 0

	var c *cron.Cron
	var spec string

	//新标准。
	c = cron.New()
	spec = "@every 2s"

	_, _ = c.AddFunc(spec, func() {
		i++
		log.Println("cron running:", i)
	})
	c.Start()

	time.Sleep(70 * time.Second)

}

func TestValidateCron(t *testing.T) {

	spec := "@every 1s"
	_, err := cron.ParseStandard(spec)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("%s passed\n", spec)
	}

}
