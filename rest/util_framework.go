package rest


//带有panic恢复的方法
func  PanicHandler() {
	if err := recover(); err != nil {
		LOGGER.Error("异步任务错误: %v", err)
	}
}

//带有panic恢复的方法
func  SafeMethod(f func()) {
	defer PanicHandler()
	//执行函数
	f()
}
