package rest

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/json-iterator/go"
	"io/ioutil"
	"time"
	"unsafe"
	"os"
	"strconv"
)

const (
	//用户身份的cookie字段名
	COOKIE_AUTH_KEY = "_ak"

	//数据库表前缀 tank100表示当前应用版本是tank:1.0.x版，数据库结构发生变化必然是中型升级
	TABLE_PREFIX = "tank10_"

	//当前版本
	VERSION = "1.0.2"
)

/*
如果你需要在本地127.0.0.1创建默认的数据库和账号，使用以下语句。
create database tank;
grant all privileges on tank.* to tank identified by 'tank123';
flush privileges;
*/
/*
 你也可以在运行时的参数中临时修改一些配置项：
-MysqlHost=127.0.0.1 -MysqlPort=3306 -MysqlSchema=tank -MysqlUsername=tank -MysqlPassword=tank123
*/
var (
	CONFIG = &Config{
		//以下内容是默认配置项。

		//默认监听端口号
		ServerPort: 6010,
		//将日志输出到控制台。
		LogToConsole: true,
		//mysql相关配置。
		//数据库端口
		MysqlPort: 3306,
		//数据库Host
		MysqlHost: "127.0.0.1",
		//数据库名字
		MysqlSchema: "tank",
		//用户名
		MysqlUsername: "tank",
		//密码
		MysqlPassword: "tank123",
		//数据库连接信息。这一项是上面几项组合而得，不可直接配置。
		MysqlUrl: "%MysqlUsername:%MysqlPassword@tcp(%MysqlHost:%MysqlPort)/%MysqlSchema?charset=utf8&parseTime=True&loc=Local",
		//超级管理员用户名，只能包含英文和数字
		AdminUsername: "admin",
		//超级管理员邮箱
		AdminEmail: "admin@tank.eyeblue.cn",
		//超级管理员密码
		AdminPassword: "123456",
	}
)

//依赖外部定义的变量。
type Config struct {
	//默认监听端口号
	ServerPort int

	//将日志输出到控制台。
	LogToConsole bool

	//mysql相关配置。
	//数据库端口
	MysqlPort int
	//数据库Host
	MysqlHost string
	//数据库名字
	MysqlSchema string
	//用户名
	MysqlUsername string
	//密码
	MysqlPassword string
	//数据库连接信息。
	MysqlUrl string

	//超级管理员用户名，只能包含英文和数字
	AdminUsername string
	//超级管理员邮箱
	AdminEmail string
	//超级管理员密码
	AdminPassword string
}

//验证配置文件的正确性。
func (this *Config) validate() {

	if this.ServerPort == 0 {
		LogPanic("ServerPort 未配置")
	}

	if this.MysqlUsername == "" {
		LogPanic("MysqlUsername 未配置")
	}

	if this.MysqlPassword == "" {
		LogPanic("MysqlPassword 未配置")
	}

	if this.MysqlHost == "" {
		LogPanic("MysqlHost 未配置")
	}

	if this.MysqlPort == 0 {
		LogPanic("MysqlPort 未配置")
	}

	if this.MysqlSchema == "" {
		LogPanic("MysqlSchema 未配置")
	}

	this.MysqlUrl = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", this.MysqlUsername, this.MysqlPassword, this.MysqlHost, this.MysqlPort, this.MysqlSchema)

}

//init方法只要这个包被引入了就一定会执行。
func init() {

	//json中需要去特殊处理时间。
	jsoniter.RegisterTypeDecoderFunc("time.Time", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		//如果使用time.UTC，那么时间会相差8小时
		t, err := time.ParseInLocation("2006-01-02 15:04:05", iter.ReadString(), time.Local)
		if err != nil {
			iter.Error = err
			return
		}
		*((*time.Time)(ptr)) = t
	})

	jsoniter.RegisterTypeEncoderFunc("time.Time", func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		t := *((*time.Time)(ptr))
		//如果使用time.UTC，那么时间会相差8小时
		stream.WriteString(t.Local().Format("2006-01-02 15:04:05"))
	}, nil)
}

//第一级. 从配置文件conf/tank.json中读取配置项
func LoadConfigFromFile() {
	//读取配置文件
	filePath := GetConfPath() + "/tank.json"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		LogWarning(fmt.Sprintf("无法找到配置文件：%s,错误：%v\n将使用config.go中的默认配置项。", filePath, err))
	} else {
		// 用 json.Unmarshal
		err := json.Unmarshal(content, CONFIG)
		if err != nil {
			LogPanic("配置文件格式错误！")
		}
	}

}

//第二级. 从环境变量中读取配置项
func LoadConfigFromEnvironment() {

	tmpServerPort := os.Getenv("TANK_SERVER_PORT")
	if tmpServerPort != "" {
		i, e := strconv.Atoi(tmpServerPort)
		if e == nil {
			CONFIG.ServerPort = i
		} else {
			LogPanic(fmt.Sprintf("环境变量TANK_SERVER_PORT必须为整数！%v", tmpServerPort))
		}
	}

	tmpMysqlPort := os.Getenv("TANK_MYSQL_PORT")
	if tmpMysqlPort != "" {
		i, e := strconv.Atoi(tmpMysqlPort)
		if e == nil {
			CONFIG.MysqlPort = i
		} else {
			LogPanic(fmt.Sprintf("环境变量TANK_MYSQL_PORT必须为整数！%v", tmpMysqlPort))
		}
	}

	tmpMysqlHost := os.Getenv("TANK_MYSQL_HOST")
	if tmpMysqlHost != "" {
		CONFIG.MysqlHost = tmpMysqlHost
	}
	tmpMysqlSchema := os.Getenv("TANK_MYSQL_SCHEMA")
	if tmpMysqlSchema != "" {
		CONFIG.MysqlSchema = tmpMysqlSchema
	}
	tmpMysqlUsername := os.Getenv("TANK_MYSQL_USERNAME")
	if tmpMysqlUsername != "" {
		CONFIG.MysqlUsername = tmpMysqlUsername
	}
	tmpMysqlPassword := os.Getenv("TANK_MYSQL_PASSWORD")
	if tmpMysqlPassword != "" {
		CONFIG.MysqlPassword = tmpMysqlPassword
	}
	tmpAdminUsername := os.Getenv("TANK_ADMIN_USERNAME")
	if tmpAdminUsername != "" {
		CONFIG.AdminUsername = tmpAdminUsername
	}
	tmpAdminEmail := os.Getenv("TANK_ADMIN_EMAIL")
	if tmpAdminEmail != "" {
		CONFIG.AdminEmail = tmpAdminEmail
	}
	tmpAdminPassword := os.Getenv("TANK_ADMIN_PASSWORD")
	if tmpAdminPassword != "" {
		CONFIG.AdminPassword = tmpAdminPassword
	}

}

//第三级. 从程序参数中读取配置项
func LoadConfigFromArguments() {

	//从运行时参数中读取，运行时参数具有更高优先级。
	//系统端口号
	ServerPortPtr := flag.Int("ServerPort", CONFIG.ServerPort, "server port")

	//系统端口号
	LogToConsolePtr := flag.Bool("LogToConsole", CONFIG.LogToConsole, "write log to console. for debug.")

	//mysql相关配置。
	MysqlPortPtr := flag.Int("MysqlPort", CONFIG.MysqlPort, "mysql port")
	MysqlHostPtr := flag.String("MysqlHost", CONFIG.MysqlHost, "mysql host")
	MysqlSchemaPtr := flag.String("MysqlSchema", CONFIG.MysqlSchema, "mysql schema")
	MysqlUsernamePtr := flag.String("MysqlUsername", CONFIG.MysqlUsername, "mysql username")
	MysqlPasswordPtr := flag.String("MysqlPassword", CONFIG.MysqlPassword, "mysql password")

	//超级管理员信息
	AdminUsernamePtr := flag.String("AdminUsername", CONFIG.AdminUsername, "administrator username")
	AdminEmailPtr := flag.String("AdminEmail", CONFIG.AdminEmail, "administrator email")
	AdminPasswordPtr := flag.String("AdminPassword", CONFIG.AdminPassword, "administrator password")

	//flag.Parse()方法必须要在使用之前调用。
	flag.Parse()

	if *ServerPortPtr != CONFIG.ServerPort {
		CONFIG.ServerPort = *ServerPortPtr
	}

	if *LogToConsolePtr != CONFIG.LogToConsole {
		CONFIG.LogToConsole = *LogToConsolePtr
	}

	if *MysqlPortPtr != CONFIG.MysqlPort {
		CONFIG.MysqlPort = *MysqlPortPtr
	}

	if *MysqlHostPtr != CONFIG.MysqlHost {
		CONFIG.MysqlHost = *MysqlHostPtr
	}

	if *MysqlSchemaPtr != CONFIG.MysqlSchema {
		CONFIG.MysqlSchema = *MysqlSchemaPtr
	}

	if *MysqlUsernamePtr != CONFIG.MysqlUsername {
		CONFIG.MysqlUsername = *MysqlUsernamePtr
	}

	if *MysqlPasswordPtr != CONFIG.MysqlPassword {
		CONFIG.MysqlPassword = *MysqlPasswordPtr
	}

	if *AdminUsernamePtr != CONFIG.AdminUsername {
		CONFIG.AdminUsername = *AdminUsernamePtr
	}

	if *AdminEmailPtr != CONFIG.AdminEmail {
		CONFIG.AdminEmail = *AdminEmailPtr
	}

	if *AdminPasswordPtr != CONFIG.AdminPassword {
		CONFIG.AdminPassword = *AdminPasswordPtr
	}

}

//三种方式指定配置项，后面的策略会覆盖前面的策略。
func PrepareConfigs() {

	//第一级. 从配置文件conf/tank.json中读取配置项
	LoadConfigFromFile()

	//第二级. 从环境变量中读取配置项
	LoadConfigFromEnvironment()

	//第三级. 从程序参数中读取配置项
	LoadConfigFromArguments()

	//验证配置项的正确性
	CONFIG.validate()

	//安装程序开始导入初始表和初始数据。
	InstallDatabase()

}
