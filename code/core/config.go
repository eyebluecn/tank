package core

const (
	//authentication key of cookie
	COOKIE_AUTH_KEY = "_ak"

	USERNAME_KEY = "_username"
	PASSWORD_KEY = "_password"

	DEFAULT_SERVER_PORT = 6010

	//db table's prefix. tank31_ means current version is tank:3.1.x
	TABLE_PREFIX = "tank31_"

	VERSION = "3.1.1"
)

type Config interface {
	Installed() bool
	ServerPort() int
	//get the mysql url. eg. tank:tank123@tcp(127.0.0.1:3306)/tank?charset=utf8&parseTime=True&loc=Local
	MysqlUrl() string
	//files storage location.
	MatterPath() string
	//when installed by user. Write configs to tank.json
	FinishInstall(mysqlPort int, mysqlHost string, mysqlSchema string, mysqlUsername string, mysqlPassword string, mysqlCharset string)
}
