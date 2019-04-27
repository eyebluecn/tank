package core

//该文件中记录的是应用系统中全局变量。主要有日志LOGGER和上下文CONTEXT

//日志系统必须高保
//全局唯一的日志对象(在main函数中初始化)
var LOGGER Logger

//全局唯一配置
var CONFIG Config

//全局唯一的上下文(在main函数中初始化)
var CONTEXT Context
