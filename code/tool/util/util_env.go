package util

import (
	"os"
	"os/user"
	"strings"
)

//是否为win开发环境
func EnvWinDevelopment() bool {

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	//如果exPath中包含了 \\AppData\\Local\\Temp 我们认为是在Win的开发环境中
	systemUser, err := user.Current()
	if systemUser != nil {

		return strings.HasPrefix(ex, systemUser.HomeDir+"\\AppData\\Local\\Temp")

	}

	return false

}

//是否为mac开发环境
func EnvMacDevelopment() bool {

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	return strings.HasPrefix(ex, "/private/var/folders")

}

//是否为开发环境 (即是否在IDE中运行)
func EnvDevelopment() bool {

	return EnvWinDevelopment() || EnvMacDevelopment()

}
