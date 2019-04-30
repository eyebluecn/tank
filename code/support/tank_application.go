package support

import (
	"flag"
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"syscall"
)

const (
	//启动web服务，默认是这种方式
	MODE_WEB = "web"
	//映射本地文件到云盘中
	MODE_MIRROR = "mirror"
	//将远程的一个文件爬取到蓝眼云盘中
	MODE_CRAWL = "crawl"
)

//命令行输入相关的对象
type TankApplication struct {
	//模式
	mode string

	//蓝眼云盘的主机，需要带上协议和端口号。默认： http://127.0.0.1:core.DEFAULT_SERVER_PORT
	host string
	//用户名
	username string
	//密码
	password string

	//源文件/文件夹，本地绝对路径/远程资源url
	src string
	//目标(表示的是文件夹)路径，蓝眼云盘中的路径。相对于root的路径。
	dest string
	//同名文件或文件夹是否直接替换 true 全部替换； false 跳过
	overwrite bool
	//拉取文件存储的名称
	filename string
}

//启动应用。可能是web形式，也可能是命令行工具。
func (this *TankApplication) Start() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("ERROR:%v\r\n", err)
		}
	}()

	//超级管理员信息
	modePtr := flag.String("mode", this.mode, "cli mode web/mirror/crawl")
	hostPtr := flag.String("host", this.username, "tank host")
	usernamePtr := flag.String("username", this.username, "username")
	passwordPtr := flag.String("password", this.password, "password")
	srcPtr := flag.String("src", this.src, "src absolute path")
	destPtr := flag.String("dest", this.dest, "destination path in tank.")
	overwritePtr := flag.Bool("overwrite", this.overwrite, "whether same file overwrite")
	filenamePtr := flag.String("filename", this.filename, "filename when crawl")

	//flag.Parse()方法必须要在使用之前调用。
	flag.Parse()

	this.mode = *modePtr
	this.host = *hostPtr
	this.username = *usernamePtr
	this.password = *passwordPtr
	this.src = *srcPtr
	this.dest = *destPtr
	this.overwrite = *overwritePtr
	this.filename = *filenamePtr

	//默认采用web的形式启动
	if this.mode == "" || strings.ToLower(this.mode) == MODE_WEB {

		//直接web启动
		this.HandleWeb()

	} else {

		//准备蓝眼云盘地址
		if this.host == "" {
			this.host = fmt.Sprintf("http://127.0.0.1:%d", core.DEFAULT_SERVER_PORT)
		}

		//准备用户名
		if this.username == "" {
			panic(result.BadRequest("%s模式下，用户名必填", this.mode))
		}

		//准备密码
		if this.password == "" {

			if util.EnvDevelopment() {

				panic(result.BadRequest("IDE中请运行请直接使用 -password yourPassword 的形式输入密码"))

			} else {

				fmt.Print("Enter Password:")
				bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
				if err != nil {
					panic(err)
				}

				this.password = string(bytePassword)
				fmt.Println()
			}
		}

		if strings.ToLower(this.mode) == MODE_MIRROR {

			//映射本地文件到蓝眼云盘
			this.HandleMirror()

		} else if strings.ToLower(this.mode) == MODE_CRAWL {

			//将远程文件拉取到蓝眼云盘中
			this.HandleCrawl()

		} else {
			panic(result.BadRequest("不能处理命名行模式： %s \r\n", this.mode))
		}
	}

}

//采用Web的方式启动应用
func (this *TankApplication) HandleWeb() {

	//第1步。日志
	tankLogger := &TankLogger{}
	core.LOGGER = tankLogger
	tankLogger.Init()
	defer tankLogger.Destroy()

	//第2步。配置
	tankConfig := &TankConfig{}
	core.CONFIG = tankConfig
	tankConfig.Init()

	//第3步。全局运行的上下文
	tankContext := &TankContext{}
	core.CONTEXT = tankContext
	tankContext.Init()
	defer tankContext.Destroy()

	//第4步。启动http服务
	http.Handle("/", core.CONTEXT)
	core.LOGGER.Info("App started at http://localhost:%v", core.CONFIG.ServerPort())

	dotPort := fmt.Sprintf(":%v", core.CONFIG.ServerPort())
	err1 := http.ListenAndServe(dotPort, nil)
	if err1 != nil {
		log.Fatal("ListenAndServe: ", err1)
	}

}

//处理本地映射的情形
func (this *TankApplication) HandleMirror() {

	if this.src == "" {
		panic("src 必填")
	}
	if this.dest == "" {
		panic("dest 必填")
	}

	fmt.Printf("开始映射本地文件 %s 到蓝眼云盘 %s\r\n", this.src, this.dest)

	urlString := fmt.Sprintf("%s/api/matter/mirror", this.host)

	params := url.Values{
		"srcPath":         {this.src},
		"destPath":        {this.dest},
		"overwrite":       {fmt.Sprintf("%v", this.overwrite)},
		core.USERNAME_KEY: {this.username},
		core.PASSWORD_KEY: {this.password},
	}

	response, err := http.PostForm(urlString, params)
	util.PanicError(err)

	bodyBytes, err := ioutil.ReadAll(response.Body)

	webResult := &result.WebResult{}

	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(bodyBytes, webResult)
	if err != nil {
		fmt.Printf("返回格式错误！%s \r\n", err.Error())
		return
	}

	if webResult.Code == result.OK.Code {
		fmt.Println("success")
	} else {
		fmt.Printf("error %s\r\n", webResult.Msg)
	}

}

//拉取远程文件到本地。
func (this *TankApplication) HandleCrawl() {

	if this.src == "" {
		panic("src 必填")
	}
	if this.dest == "" {
		panic("dest 必填")
	}

	if this.filename == "" {
		panic("filename 必填")
	}

	fmt.Printf("开始映射拉取远程文件 %s 到蓝眼云盘 %s\r\n", this.src, this.dest)

	urlString := fmt.Sprintf("%s/api/matter/crawl", this.host)

	params := url.Values{
		"url":             {this.src},
		"destPath":        {this.dest},
		"filename":        {this.filename},
		core.USERNAME_KEY: {this.username},
		core.PASSWORD_KEY: {this.password},
	}

	response, err := http.PostForm(urlString, params)
	util.PanicError(err)

	bodyBytes, err := ioutil.ReadAll(response.Body)

	webResult := &result.WebResult{}

	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(bodyBytes, webResult)
	if err != nil {
		fmt.Printf("返回格式错误！%s \r\n", err.Error())
		return
	}

	if webResult.Code == result.OK.Code {
		fmt.Println("success")
	} else {
		fmt.Printf("error %s\r\n", webResult.Msg)
	}

}
