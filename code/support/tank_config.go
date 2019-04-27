package support

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/json-iterator/go"
	"io/ioutil"
	"os"
	"time"
	"unsafe"
)

/*
如果你需要在本地127.0.0.1创建默认的数据库和账号，使用以下语句。
create database tank;
grant all privileges on tank.* to tank identified by 'tank123';
flush privileges;
*/

//依赖外部定义的变量。
type TankConfig struct {
	//默认监听端口号
	serverPort int
	//网站是否已经完成安装
	installed bool
	//上传的文件路径，要求不以/结尾。如果没有指定，默认在根目录下的matter文件夹中。eg: /var/www/matter
	matterPath string
	//数据库连接信息。
	mysqlUrl string
	//配置文件中的项
	item *ConfigItem
}

//和tank.json文件中的键值一一对应。
type ConfigItem struct {
	//默认监听端口号
	ServerPort int
	//上传的文件路径，要求不以/结尾。如果没有指定，默认在根目录下的matter文件夹中。eg: /var/www/matter
	MatterPath string
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
}

//验证配置文件的正确性。
func (this *ConfigItem) validate() bool {

	if this.ServerPort == 0 {
		core.LOGGER.Error("ServerPort 未配置")
		return false
	}

	if this.MysqlUsername == "" {
		core.LOGGER.Error("MysqlUsername 未配置")
		return false
	}

	if this.MysqlPassword == "" {
		core.LOGGER.Error("MysqlPassword 未配置")
		return false
	}

	if this.MysqlHost == "" {
		core.LOGGER.Error("MysqlHost 未配置")
		return false
	}

	if this.MysqlPort == 0 {
		core.LOGGER.Error("MysqlPort 未配置")
		return false
	}

	if this.MysqlSchema == "" {
		core.LOGGER.Error("MysqlSchema 未配置")
		return false
	}

	return true

}

//验证配置文件是否完好
func (this *TankConfig) Init() {

	//JSON初始化
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

	//默认从6010端口启动
	this.serverPort = 6010

	this.ReadFromConfigFile()

}

//系统如果安装好了就调用这个方法。
func (this *TankConfig) ReadFromConfigFile() {

	//读取配置文件
	filePath := util.GetConfPath() + "/tank.json"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		core.LOGGER.Warn("无法找到配置文件：%s 即将进入安装过程！", filePath)
		this.installed = false
	} else {
		this.item = &ConfigItem{}
		core.LOGGER.Warn("读取配置文件：%s", filePath)
		err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(content, this.item)
		if err != nil {
			core.LOGGER.Error("配置文件格式错误！ 即将进入安装过程！")
			this.installed = false
			return
		}

		//验证项是否齐全
		itemValidate := this.item.validate()
		if !itemValidate {
			core.LOGGER.Error("配置文件信息不齐全！ 即将进入安装过程！")
			this.installed = false
			return
		}

		//使用配置项中的文件路径
		if this.item.MatterPath == "" {
			this.matterPath = util.GetHomePath() + "/matter"
		} else {
			this.matterPath = this.item.MatterPath
		}
		util.MakeDirAll(this.matterPath)

		//使用配置项中的端口
		if this.item.ServerPort != 0 {
			this.serverPort = this.item.ServerPort
		}

		this.mysqlUrl = util.GetMysqlUrl(this.item.MysqlPort, this.item.MysqlHost, this.item.MysqlSchema, this.item.MysqlUsername, this.item.MysqlPassword)
		this.installed = true

		core.LOGGER.Info("使用配置文件：%s", filePath)
		core.LOGGER.Info("上传文件存放路径：%s", this.matterPath)
	}
}

//是否已经安装
func (this *TankConfig) Installed() bool {
	return this.installed
}

//启动端口
func (this *TankConfig) ServerPort() int {
	return this.serverPort
}

//获取mysql链接
func (this *TankConfig) MysqlUrl() string {
	return this.mysqlUrl
}

//文件存放路径
func (this *TankConfig) MatterPath() string {
	return this.matterPath
}

//完成安装过程，主要是要将配置写入到文件中
func (this *TankConfig) FinishInstall(mysqlPort int, mysqlHost string, mysqlSchema string, mysqlUsername string, mysqlPassword string) {

	var configItem = &ConfigItem{
		//默认监听端口号
		ServerPort: core.CONFIG.ServerPort(),
		//上传的文件路径，要求不以/结尾。如果没有指定，默认在根目录下的matter文件夹中。eg: /var/www/matter
		MatterPath: core.CONFIG.MatterPath(),
		//mysql相关配置。
		//数据库端口
		MysqlPort: mysqlPort,
		//数据库Host
		MysqlHost: mysqlHost,
		//数据库名字
		MysqlSchema: mysqlSchema,
		//用户名
		MysqlUsername: mysqlUsername,
		//密码
		MysqlPassword: mysqlPassword,
	}

	//用json的方式输出返回值。为了让格式更好看。
	jsonStr, _ := jsoniter.ConfigCompatibleWithStandardLibrary.MarshalIndent(configItem, "", " ")

	//写入到配置文件中（不能使用os.O_APPEND 否则会追加）
	filePath := util.GetConfPath() + "/tank.json"
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0777)
	util.PanicError(err)
	_, err = f.Write(jsonStr)
	util.PanicError(err)
	err = f.Close()
	util.PanicError(err)

	this.ReadFromConfigFile()

}
