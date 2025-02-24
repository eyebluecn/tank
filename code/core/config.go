package core

import "gorm.io/gorm/schema"

const (
	//authentication key of cookie
	COOKIE_AUTH_KEY = "_ak"

	USERNAME_KEY = "_username"
	PASSWORD_KEY = "_password"

	DEFAULT_SERVER_PORT = 6010

	//db table's prefix. tank41_ means current version is tank:4.1.x
	TABLE_PREFIX = "tank41_"

	VERSION = "4.1.2"
)

type Config interface {
	Installed() bool
	ServerPort() int
	//get the db type
	DbType() string
	//get the mysql url. eg. tank:tank123@tcp(127.0.0.1:3306)/tank?charset=utf8&parseTime=True&loc=Local
	MysqlUrl() string
	//get the sqlite path
	SqliteFolder() string
	//files storage location.
	MatterPath() string
	//table name strategy
	NamingStrategy() schema.NamingStrategy
	//when installed by user. Write configs to tank.json
	FinishInstall(dbType string, mysqlPort int, mysqlHost string, mysqlSchema string, mysqlUsername string, mysqlPassword string, mysqlCharset string)
}
