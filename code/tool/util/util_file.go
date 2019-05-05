package util

import (
	"fmt"
	"github.com/eyebluecn/tank/code/tool/result"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else {
		if os.IsNotExist(err) {
			return false
		} else {
			panic(result.BadRequest(err.Error()))
		}
	}
}

func GetGoPath() string {

	return build.Default.GOPATH

}

//get development home path.
func GetDevHomePath() string {

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot get dev home path.")
	}

	//$DevHomePath/code/tool/util/util_file.go
	dir := GetDirOfPath(file)
	dir = GetDirOfPath(dir)
	dir = GetDirOfPath(dir)
	dir = GetDirOfPath(dir)

	return dir
}

//get home path for application.
func GetHomePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	if EnvMacDevelopment() {
		exPath = GetDevHomePath() + "/tmp"
	}

	if EnvWinDevelopment() {
		exPath = GetDevHomePath() + "/tmp"
	}

	return UniformPath(exPath)
}

//get html path
//dev: return $project/build/html
//prod: return $application/html
func GetHtmlPath() string {

	//开发环境直接使用 build/html 下面的文件
	if EnvWinDevelopment() || EnvMacDevelopment() {
		return GetDevHomePath() + "/build/html"
	}
	return GetHomePath() + "/html"
}

//if directory not exit, create it.
func MakeDirAll(dirPath string) string {

	exists := PathExists(dirPath)

	if !exists {

		err := os.MkdirAll(dirPath, 0777)
		if err != nil {
			panic("error while creating directory")
		}
	}

	return dirPath
}

//eg /var/www/xx.log -> /var/www
func GetDirOfPath(fullPath string) string {

	index1 := strings.LastIndex(fullPath, "/")
	//maybe windows environment
	index2 := strings.LastIndex(fullPath, "\\")
	index := index1
	if index2 > index1 {
		index = index2
	}

	return fullPath[:index]
}

//get filename from path. eg /var/www/xx.log -> xx.log
func GetFilenameOfPath(fullPath string) string {

	index1 := strings.LastIndex(fullPath, "/")
	//maybe windows env
	index2 := strings.LastIndex(fullPath, "\\")
	index := index1
	if index2 > index1 {
		index = index2
	}

	return fullPath[index+1:]
}

//try to delete empty dir. true: delete an empty dir, false: delete nothing.
func DeleteEmptyDir(dirPath string) bool {
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		panic(result.BadRequest("occur error while reading %s %s", dirPath, err.Error()))
	}
	if len(dir) == 0 {
		//empty dir
		err = os.Remove(dirPath)
		if err != nil {
			panic(result.BadRequest("occur error while deleting %s %s", dirPath, err.Error()))
		}
		return true
	}

	return false
}

//delete empty dir recursive, delete until not empty.
func DeleteEmptyDirRecursive(dirPath string) {

	fmt.Printf("recursive delete %v \n", dirPath)

	tmpPath := dirPath
	for DeleteEmptyDir(tmpPath) {

		dir := GetDirOfPath(tmpPath)

		fmt.Printf("try to delete %v\n", dir)

		tmpPath = dir
	}
}

//get conf path.
func GetConfPath() string {

	homePath := GetHomePath()
	filePath := homePath + "/conf"
	exists := PathExists(filePath)

	if !exists {
		err := os.MkdirAll(filePath, 0777)
		if err != nil {
			panic("error while mkdir " + err.Error())
		}
	}

	return filePath
}

//get log path.
func GetLogPath() string {

	homePath := GetHomePath()
	filePath := homePath + "/log"
	exists := PathExists(filePath)

	if !exists {
		err := os.MkdirAll(filePath, 0777)
		if err != nil {
			panic("error while mkdir " + err.Error())
		}
	}

	return filePath
}

//copy file
func CopyFile(srcPath string, destPath string) (nBytes int64) {

	srcFileStat, err := os.Stat(srcPath)
	if err != nil {
		panic(err)
	}

	if !srcFileStat.Mode().IsRegular() {
		panic(fmt.Errorf("%s is not a regular file", srcPath))
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = srcFile.Close()
		if err != nil {
			panic(err)
		}
	}()

	destFile, err := os.Create(destPath)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = destFile.Close()
		if err != nil {
			panic(err)
		}
	}()

	nBytes, err = io.Copy(destFile, srcFile)
	return nBytes
}

//1. replace \\ to /
//2. clean path.
//3. trim suffix /
func UniformPath(p string) string {
	p = strings.Replace(p, "\\", "/", -1)
	p = path.Clean(p)
	p = strings.TrimSuffix(p, "/")
	return p
}
