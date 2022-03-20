package support

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/json-iterator/go"
	"gorm.io/gorm/schema"
	"io/ioutil"
	"os"
	"time"
	"unsafe"
)

type TankConfig struct {
	//server port
	serverPort int
	//whether installed
	installed bool
	//file storage location. eg./var/www/matter
	matterPath string
	//mysql url.
	mysqlUrl string
	//sqlite file path
	sqliteFolder string
	//configs in tank.json
	item *ConfigItem
}

//tank.json config items.
type ConfigItem struct {
	//server port
	ServerPort int
	//file storage location. eg./var/www/matter
	MatterPath string
	//********db configurations.********
	//default value is "mysql"
	DbType string
	//********mysql configurations..********
	//mysql port
	MysqlPort int
	//mysql host
	MysqlHost string
	//mysql schema
	MysqlSchema string
	//mysql username
	MysqlUsername string
	//mysql password
	MysqlPassword string
	//mysql charset
	MysqlCharset string
	//********sqlite configurations..********
	//default value is matter/
	SqliteFolder string
}

//validate whether the config file is ok
func (this *ConfigItem) validate() bool {

	if this.ServerPort == 0 {
		core.LOGGER.Error("ServerPort is not configured")
		return false
	}

	if this.DbType == "sqlite" {

	} else {

		if this.MysqlUsername == "" {
			core.LOGGER.Error("MysqlUsername  is not configured")
			return false
		}

		if this.MysqlPassword == "" {
			core.LOGGER.Error("MysqlPassword  is not configured")
			return false
		}

		if this.MysqlHost == "" {
			core.LOGGER.Error("MysqlHost  is not configured")
			return false
		}

		if this.MysqlPort == 0 {
			core.LOGGER.Error("MysqlPort  is not configured")
			return false
		}

		if this.MysqlSchema == "" {
			core.LOGGER.Error("MysqlSchema  is not configured")
			return false
		}

		if this.MysqlCharset == "" {
			core.LOGGER.Error("MysqlCharset  is not configured")
			return false
		}
	}

	return true

}

func (this *TankConfig) Init() {

	//JSON init.
	jsoniter.RegisterTypeDecoderFunc("time.Time", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		//if use time.UTC there will be 8 hours gap.
		t, err := time.ParseInLocation("2006-01-02 15:04:05", iter.ReadString(), time.Local)
		if err != nil {
			iter.Error = err
			return
		}
		*((*time.Time)(ptr)) = t
	})

	jsoniter.RegisterTypeEncoderFunc("time.Time", func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		t := *((*time.Time)(ptr))
		//if use time.UTC there will be 8 hours gap.
		stream.WriteString(t.Local().Format("2006-01-02 15:04:05"))
	}, nil)

	//default server port.
	this.serverPort = core.DEFAULT_SERVER_PORT

	this.ReadFromConfigFile()

}

func (this *TankConfig) ReadFromConfigFile() {

	//read from tank.json
	filePath := util.GetConfPath() + "/tank.json"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		core.LOGGER.Warn("cannot find config file %s, installation will start!", filePath)
		this.installed = false
	} else {
		this.item = &ConfigItem{}
		core.LOGGER.Warn("read config file %s", filePath)
		err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(content, this.item)
		if err != nil {
			core.LOGGER.Error("config file error, installation will start!")
			this.installed = false
			return
		}

		//use default server port
		if this.item.ServerPort != 0 {
			this.serverPort = this.item.ServerPort
		}

		//check the integrity
		itemValidate := this.item.validate()
		if !itemValidate {
			core.LOGGER.Error("config file not integrity, installation will start!")
			this.installed = false
			return
		}

		//use default file location.
		if this.item.MatterPath == "" {
			this.matterPath = util.GetHomePath() + "/matter"
		} else {
			this.matterPath = util.UniformPath(this.item.MatterPath)
		}
		util.MakeDirAll(this.matterPath)

		this.mysqlUrl = util.GetMysqlUrl(this.item.MysqlPort, this.item.MysqlHost, this.item.MysqlSchema, this.item.MysqlUsername, this.item.MysqlPassword, this.item.MysqlCharset)
		this.installed = true

		core.LOGGER.Info("use config file: %s", filePath)
		core.LOGGER.Info("file storage location: %s", this.matterPath)
	}
}

//whether installed.
func (this *TankConfig) Installed() bool {
	return this.installed
}

//server port
func (this *TankConfig) ServerPort() int {
	return this.serverPort
}

//get the db type
func (this *TankConfig) DbType() string {
	return this.item.DbType
}

//mysql url
func (this *TankConfig) MysqlUrl() string {
	return this.mysqlUrl
}

//get the sqlite path
func (this *TankConfig) SqliteFolder() string {
	if this.sqliteFolder == "" {
		//use default file location.
		if this.item == nil || this.item.SqliteFolder == "" {
			this.sqliteFolder = util.GetHomePath() + "/matter"
		} else {
			this.sqliteFolder = util.UniformPath(this.item.SqliteFolder)
		}
		util.MakeDirAll(this.sqliteFolder)
	}

	return this.sqliteFolder
}

//matter path
func (this *TankConfig) MatterPath() string {
	return this.matterPath
}

//matter path
func (this *TankConfig) NamingStrategy() schema.NamingStrategy {
	return schema.NamingStrategy{
		TablePrefix:   core.TABLE_PREFIX,
		SingularTable: true,
	}
}

//TODO: Finish the installation. Write config to tank.json. add sqlite support.
func (this *TankConfig) FinishInstall(dbType string, mysqlPort int, mysqlHost string, mysqlSchema string, mysqlUsername string, mysqlPassword string, mysqlCharset string) {

	var configItem = &ConfigItem{
		DbType: dbType,
		//server port
		ServerPort: core.CONFIG.ServerPort(),
		//file storage location. eg./var/www/matter
		MatterPath:    core.CONFIG.MatterPath(),
		MysqlPort:     mysqlPort,
		MysqlHost:     mysqlHost,
		MysqlSchema:   mysqlSchema,
		MysqlUsername: mysqlUsername,
		MysqlPassword: mysqlPassword,
		MysqlCharset:  mysqlCharset,
	}

	//pretty json.
	jsonStr, _ := jsoniter.ConfigCompatibleWithStandardLibrary.MarshalIndent(configItem, "", " ")

	//Write to tank.json (cannot use os.O_APPEND  or append)
	filePath := util.GetConfPath() + "/tank.json"
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0777)
	core.PanicError(err)
	_, err = f.Write(jsonStr)
	core.PanicError(err)
	err = f.Close()
	core.PanicError(err)

	this.ReadFromConfigFile()

}
