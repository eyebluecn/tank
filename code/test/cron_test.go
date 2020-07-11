package test

import (
	"github.com/robfig/cron/v3"
	"testing"
)

func TestValidateCron(t *testing.T) {

	spec := "*/1 * * * * ?"
	_, err := cron.ParseStandard(spec)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("%s passed\n", spec)
	}

}
