package core

const (
	//用户身份的cookie字段名
	COOKIE_AUTH_KEY = "_ak"

	//使用用户名密码给接口授权key
	USERNAME_KEY = "_username"
	PASSWORD_KEY = "_password"

	//默认端口号
	DEFAULT_SERVER_PORT = 6010

	//数据库表前缀 tank30_表示当前应用版本是tank:3.0.x版，数据库结构发生变化必然是中型升级
	TABLE_PREFIX = "tank30_"

	//当前版本
	VERSION = "3.0.0.beta1"
)

type Config interface {
	//是否已经安装
	Installed() bool
	//启动端口
	ServerPort() int
	//获取mysql链接
	MysqlUrl() string

	//文件存放路径
	MatterPath() string
	//完成安装过程，主要是要将配置写入到文件中
	FinishInstall(mysqlPort int, mysqlHost string, mysqlSchema string, mysqlUsername string, mysqlPassword string)
}
