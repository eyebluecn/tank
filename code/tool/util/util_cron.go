package util

import (
	"github.com/robfig/cron/v3"
)

//validate a cron
func ValidateCron(spec string) bool {

	_, err := cron.ParseStandard(spec)
	if err != nil {
		return false
	}

	return true
}
