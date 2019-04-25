package tool

import (
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

//判断文件或文件夹是否已经存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//获取GOPATH路径
func GetGoPath() string {

	return build.Default.GOPATH

}

//获取该应用可执行文件的位置。
//例如：C:\Users\lishuang\AppData\Local\Temp
func GetHomePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	//如果exPath中包含了 /private/var/folders 我们认为是在Mac的开发环境中
	macDev := strings.HasPrefix(exPath, "/private/var/folders")
	if macDev {
		exPath = GetGoPath() + "/src/tank/tmp"
	}

	//如果exPath中包含了 \\AppData\\Local\\Temp 我们认为是在Win的开发环境中
	systemUser, err := user.Current()
	winDev := strings.HasPrefix(exPath, systemUser.HomeDir+"\\AppData\\Local\\Temp")
	if winDev {
		exPath = GetGoPath() + "/src/tank/tmp"
	}

	return exPath
}

//获取前端静态资源的位置。如果你在开发模式下，可以将这里直接返回tank/build下面的html路径。
//例如：C:/Users/lishuang/AppData/Local/Temp/html
func GetHtmlPath() string {

	homePath := GetHomePath()
	filePath := homePath + "/html"
	exists, err := PathExists(filePath)
	if err != nil {
		panic("判断上传文件是否存在时出错！")
	}
	if !exists {
		err = os.MkdirAll(filePath, 0777)
		if err != nil {
			panic("创建上传文件夹时出错！")
		}
	}

	return filePath
}

//如果文件夹存在就不管，不存在就创建。 例如：/var/www/matter
func MakeDirAll(dirPath string) string {

	exists, err := PathExists(dirPath)
	if err != nil {
		panic("判断文件是否存在时出错！")
	}
	if !exists {
		//TODO:文件权限需要进一步考虑
		err = os.MkdirAll(dirPath, 0777)
		if err != nil {
			panic("创建文件夹时出错！")
		}
	}

	return dirPath
}

//获取到一个Path的文件夹路径，eg /var/www/xx.log -> /var/www
func GetDirOfPath(fullPath string) string {

	index1 := strings.LastIndex(fullPath, "/")
	//可能是windows的环境
	index2 := strings.LastIndex(fullPath, "\\")
	index := index1
	if index2 > index1 {
		index = index2
	}

	return fullPath[:index]
}

//获取到一个Path 中的文件名，eg /var/www/xx.log -> xx.log
func GetFilenameOfPath(fullPath string) string {

	index1 := strings.LastIndex(fullPath, "/")
	//可能是windows的环境
	index2 := strings.LastIndex(fullPath, "\\")
	index := index1
	if index2 > index1 {
		index = index2
	}

	return fullPath[index+1:]
}

//尝试删除空文件夹 true表示删掉了一个空文件夹，false表示没有删掉任何东西
func DeleteEmptyDir(dirPath string) bool {
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		LOGGER.Error("尝试读取目录%s时出错 %s", dirPath, err.Error())
		panic("尝试读取目录时出错 " + err.Error())
	}
	if len(dir) == 0 {
		//空文件夹
		err = os.Remove(dirPath)
		if err != nil {
			LOGGER.Error("删除磁盘上的文件夹%s出错 %s", dirPath, err.Error())
		}
		return true
	} else {
		LOGGER.Info("文件夹不为空，%v", len(dir))
	}
	return false
}

//递归尝试删除空文件夹，一直空就一直删，直到不空为止
func DeleteEmptyDirRecursive(dirPath string) {

	fmt.Printf("递归删除删 %v \n", dirPath)

	tmpPath := dirPath
	for DeleteEmptyDir(tmpPath) {

		dir := GetDirOfPath(tmpPath)

		fmt.Printf("尝试删除 %v\n", dir)

		tmpPath = dir
	}
}

//移除某个文件夹。 例如：/var/www/matter => /var/www
func RemoveDirectory(dirPath string) string {

	exists, err := PathExists(dirPath)
	if err != nil {
		panic("判断文件是否存在时出错！")
	}
	if exists {

		err = os.Remove(dirPath)
		if err != nil {
			panic("删除文件夹时出错！")
		}
	}

	return dirPath
}

//获取配置文件存放的位置
//例如：C:\Users\lishuang\AppData\Local\Temp/conf
func GetConfPath() string {

	homePath := GetHomePath()
	filePath := homePath + "/conf"
	exists, err := PathExists(filePath)
	if err != nil {
		panic("判断日志文件夹是否存在时出错！")
	}
	if !exists {
		err = os.MkdirAll(filePath, 0777)
		if err != nil {
			panic("创建日志文件夹时出错！")
		}
	}

	return filePath
}

//获取日志的路径
//例如：默认存放于 home/log
func GetLogPath() string {

	homePath := GetHomePath()
	filePath := homePath + "/log"
	exists, err := PathExists(filePath)
	if err != nil {
		panic("判断日志文件夹是否存在时出错！")
	}
	if !exists {
		err = os.MkdirAll(filePath, 0777)
		if err != nil {
			panic("创建日志文件夹时出错！")
		}
	}

	return filePath
}


//复制文件
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
