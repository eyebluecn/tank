package rest

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
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

//获取该应用可执行文件的位置。
//例如：C:\Users\lishuang\AppData\Local\Temp
func GetHomePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
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
		err = os.MkdirAll(dirPath, 0777)
		if err != nil {
			panic("创建文件夹时出错！")
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

//获取某个用户文件应该存放的位置。这个是相对GetFilePath的路径
//例如：/zicla/2006-01-02/1510122428000
func GetUserFilePath(username string, cache bool) (string, string) {

	now := time.Now()
	datePath := now.Format("2006-01-02")
	//毫秒时间戳
	timestamp := now.UnixNano() / 1e6
	
	//如果是缓存文件夹，那么统一放在cache这个文件夹下面
	if cache {
		datePath = fmt.Sprintf("cache/%s", datePath)
	}
	filePath := CONFIG.MatterPath
	absolutePath := fmt.Sprintf("%s/%s/%s/%d", filePath, username, datePath, timestamp)
	relativePath := fmt.Sprintf("/%s/%s/%d", username, datePath, timestamp)

	exists, err := PathExists(absolutePath)
	if err != nil {
		panic("判断上传文件是否存在时出错！请检查文件夹 " + filePath + " 的访问权限。")
	}
	if !exists {
		err = os.MkdirAll(absolutePath, 0777)
		if err != nil {
			panic("创建上传文件夹时出错！")
		}
	}

	return absolutePath, relativePath
}
