package util

import (
	"os"
	"os/user"
	"strings"
)

//whether windows develop environment
func EnvWinDevelopment() bool {

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	//if exPath contains \\AppData\\Local\\Temp we regard as dev.
	systemUser, err := user.Current()
	if systemUser != nil {

		return strings.HasPrefix(ex, systemUser.HomeDir+"\\AppData\\Local\\Temp")

	}

	return false

}

//whether mac develop environment
func EnvMacDevelopment() bool {

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	return strings.HasPrefix(ex, "/private/var/folders")

}

//whether develop environment (whether run in IDE)
func EnvDevelopment() bool {

	return EnvWinDevelopment() || EnvMacDevelopment()

}
