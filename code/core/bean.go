package core

/**
 * 系统中的Bean接口，即系统中单例模式
 */
type Bean interface {
	//初始化方法
	Init()
	//系统清理方法
	Cleanup()
	//所有配置都加载完成后调用的方法，包括数据库加载完毕
	Bootstrap()
	//快速的Panic方法
	PanicError(err error)
}
