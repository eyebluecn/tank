package util

//带有panic恢复的方法
func RunWithRecovery(f func()) {
	defer func() {
		if err := recover(); err != nil {
			//TODO 全局日志记录
			//LOGGER.Error("异步任务错误: %v", err)
		}
	}()

	//执行函数
	f()
}

//处理错误的统一方法 可以省去if err!=nil 这段代码
func PanicError(err error) {
	if err != nil {
		panic(err)
	}
}
