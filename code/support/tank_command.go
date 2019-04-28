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
)

//命令行输入相关的对象
type TankCommand struct {
	//模式
	mode string

	//蓝眼云盘的主机，需要带上协议和端口号。默认： http://127.0.0.1:core.DEFAULT_SERVER_PORT
	host string
	//用户名
	username string
	//密码
	password string

	//源文件/文件夹，本地绝对路径
	src string
	//目标(表示的是文件夹)路径，蓝眼云盘中的路径。相对于root的路径。
	dest string
	//同名文件或文件夹是否直接替换 true 全部替换； false 跳过
	overwrite bool
}

//第三级. 从程序参数中读取配置项
func (this *TankCommand) Cli() bool {

	//超级管理员信息
	modePtr := flag.String("mode", this.mode, "cli mode web/mirror")
	hostPtr := flag.String("host", this.username, "tank host")
	usernamePtr := flag.String("username", this.username, "username")
	passwordPtr := flag.String("password", this.password, "password")
	srcPtr := flag.String("src", this.src, "src absolute path")
	destPtr := flag.String("dest", this.dest, "destination path in tank.")
	overwritePtr := flag.Bool("overwrite", this.overwrite, "whether same file overwrite")

	//flag.Parse()方法必须要在使用之前调用。
	flag.Parse()

	this.mode = *modePtr
	this.host = *hostPtr
	this.username = *usernamePtr
	this.password = *passwordPtr
	this.src = *srcPtr
	this.dest = *destPtr
	this.overwrite = *overwritePtr

	//准备模式
	if this.mode == "" || strings.ToLower(this.mode) == MODE_WEB {
		return false
	}

	//准备蓝眼云盘地址
	if this.host == "" {
		this.host = fmt.Sprintf("http://127.0.0.1:%d", core.DEFAULT_SERVER_PORT)
	}

	//准备用户名
	if this.username == "" {
		fmt.Println("用户名必填")
		return true
	}

	//准备密码
	if this.password == "" {

		if util.EnvDevelopment() {

			fmt.Println("IDE中请运行请直接使用 -password yourPassword 的形式输入密码")
			return true

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

		this.HandleMirror()

	} else {

		fmt.Printf("不能处理命名行模式： %s \r\n", this.mode)
	}

	return true
}

//处理本地映射的情形
func (this *TankCommand) HandleMirror() {

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

	if webResult.Code == result.CODE_WRAPPER_OK.Code {
		fmt.Println("success")
	} else {
		fmt.Printf("error %s\r\n", webResult.Msg)
	}

}
