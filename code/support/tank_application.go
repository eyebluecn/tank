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
	//start web. This is the default mode.
	MODE_WEB = "web"
	//mirror local files to EyeblueTank.
	MODE_MIRROR = "mirror"
	//crawl remote file to EyeblueTank
	MODE_CRAWL = "crawl"
	//Current version.
	MODE_VERSION = "version"
	//migrate 2.0 to 3.0
	MODE_MIGRATE_20_TO_30 = "migrate20to30"
)

type TankApplication struct {
	//mode
	mode string

	//EyeblueTank host and port  default: http://127.0.0.1:core.DEFAULT_SERVER_PORT
	host string
	//username
	username string
	//password
	password string

	//source file/directory different mode has different usage.
	src string
	//destination directory path(relative to root) in EyeblueTank
	dest string
	//true: overwrite, false:skip
	overwrite bool
	filename  string
}

//Start the application.
func (this *TankApplication) Start() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("ERROR:%v\r\n", err)
		}
	}()

	modePtr := flag.String("mode", this.mode, "cli mode web/mirror/crawl")
	hostPtr := flag.String("host", this.username, "tank host")
	usernamePtr := flag.String("username", this.username, "username")
	passwordPtr := flag.String("password", this.password, "password")
	srcPtr := flag.String("src", this.src, "src absolute path")
	destPtr := flag.String("dest", this.dest, "destination path in tank.")
	overwritePtr := flag.Bool("overwrite", this.overwrite, "whether same file overwrite")
	filenamePtr := flag.String("filename", this.filename, "filename when crawl")

	//flag.Parse() must invoke before use.
	flag.Parse()

	this.mode = *modePtr
	this.host = *hostPtr
	this.username = *usernamePtr
	this.password = *passwordPtr
	this.src = *srcPtr
	this.dest = *destPtr
	this.overwrite = *overwritePtr
	this.filename = *filenamePtr

	//default start as web.
	if this.mode == "" || strings.ToLower(this.mode) == MODE_WEB {

		this.HandleWeb()

	} else if strings.ToLower(this.mode) == MODE_VERSION {

		this.HandleVersion()

	} else {

		//default host.
		if this.host == "" {
			this.host = fmt.Sprintf("http://127.0.0.1:%d", core.DEFAULT_SERVER_PORT)
		}

		if this.username == "" {
			panic(result.BadRequest("in mode %s, username is required", this.mode))
		}

		if this.password == "" {

			if util.EnvDevelopment() {
				panic(result.BadRequest("If run in IDE, use -password yourPassword to input password"))
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

		} else if strings.ToLower(this.mode) == MODE_MIGRATE_20_TO_30 {

			this.HandleMigrate20to30()

		} else if strings.ToLower(this.mode) == MODE_CRAWL {

			this.HandleCrawl()

		} else {
			panic(result.BadRequest("cannot handle mode %s \r\n", this.mode))
		}
	}

}

func (this *TankApplication) HandleWeb() {

	//Step 1. Logger
	tankLogger := &TankLogger{}
	core.LOGGER = tankLogger
	tankLogger.Init()
	defer tankLogger.Destroy()

	//Step 2. Configuration
	tankConfig := &TankConfig{}
	core.CONFIG = tankConfig
	tankConfig.Init()

	//Step 3. Global Context
	tankContext := &TankContext{}
	core.CONTEXT = tankContext
	tankContext.Init()
	defer tankContext.Destroy()

	//Step 4. Start http
	http.Handle("/", core.CONTEXT)
	core.LOGGER.Info("App started at http://localhost:%v", core.CONFIG.ServerPort())

	dotPort := fmt.Sprintf(":%v", core.CONFIG.ServerPort())
	err1 := http.ListenAndServe(dotPort, nil)
	if err1 != nil {
		log.Fatal("ListenAndServe: ", err1)
	}

}

func (this *TankApplication) HandleMirror() {

	if this.src == "" {
		panic("src is required")
	}
	if this.dest == "" {
		panic("dest is required")
	}

	fmt.Printf("start mirror %s to EyeblueTank %s\r\n", this.src, this.dest)

	urlString := fmt.Sprintf("%s/api/matter/mirror", this.host)

	params := url.Values{
		"srcPath":         {this.src},
		"destPath":        {this.dest},
		"overwrite":       {fmt.Sprintf("%v", this.overwrite)},
		core.USERNAME_KEY: {this.username},
		core.PASSWORD_KEY: {this.password},
	}

	response, err := http.PostForm(urlString, params)
	core.PanicError(err)

	bodyBytes, err := ioutil.ReadAll(response.Body)

	webResult := &result.WebResult{}

	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(bodyBytes, webResult)
	if err != nil {
		fmt.Printf("error response format %s \r\n", err.Error())
		return
	}

	if webResult.Code == result.OK.Code {
		fmt.Println("success")
	} else {
		fmt.Printf("error %s\r\n", webResult.Msg)
	}

}

func (this *TankApplication) HandleCrawl() {

	if this.src == "" {
		panic("src is required")
	}
	if this.dest == "" {
		panic("dest is required")
	}

	if this.filename == "" {
		panic("filename is required")
	}

	fmt.Printf("crawl %s to EyeblueTank %s\r\n", this.src, this.dest)

	urlString := fmt.Sprintf("%s/api/matter/crawl", this.host)

	params := url.Values{
		"url":             {this.src},
		"destPath":        {this.dest},
		"filename":        {this.filename},
		core.USERNAME_KEY: {this.username},
		core.PASSWORD_KEY: {this.password},
	}

	response, err := http.PostForm(urlString, params)
	core.PanicError(err)

	bodyBytes, err := ioutil.ReadAll(response.Body)

	webResult := &result.WebResult{}

	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(bodyBytes, webResult)
	if err != nil {
		fmt.Printf("Error response format %s \r\n", err.Error())
		return
	}

	if webResult.Code == result.OK.Code {
		fmt.Println("success")
	} else {
		fmt.Printf("error %s\r\n", webResult.Msg)
	}

}

//fetch the application version
func (this *TankApplication) HandleVersion() {

	fmt.Printf("EyeblueTank %s\r\n", core.VERSION)

}

//migrate 2.0 to 3.0
func (this *TankApplication) HandleMigrate20to30() {

	if this.src == "" {
		panic("src is required")
	}

	fmt.Printf("start migrating 2.0 to 3.0. MatterPath2.0 = %s \r\n", this.src)

	urlString := fmt.Sprintf("%s/api/preference/migrate20to30", this.host)

	params := url.Values{
		"matterPath":      {this.src},
		core.USERNAME_KEY: {this.username},
		core.PASSWORD_KEY: {this.password},
	}

	response, err := http.PostForm(urlString, params)
	core.PanicError(err)

	bodyBytes, err := ioutil.ReadAll(response.Body)

	webResult := &result.WebResult{}

	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(bodyBytes, webResult)
	if err != nil {
		fmt.Printf("error response format %s \r\n", err.Error())
		return
	}

	if webResult.Code == result.OK.Code {
		fmt.Println("success")
	} else {
		fmt.Printf("error %s\r\n", webResult.Msg)
	}

}
