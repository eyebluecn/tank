package core

/**
 * 从命令行输入的相关信息
 */
type Command interface {

	//判断是否为命名行模式，如果是直接按照命名行模式处理，并返回true。如果不是返回false.
	Cli() bool
}
