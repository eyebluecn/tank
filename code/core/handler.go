package core

//run a method with panic recovery.
func RunWithRecovery(f func()) {
	defer func() {
		if err := recover(); err != nil {
			LOGGER.Error("error in async method: %v", err)
		}
	}()

	//execute the method
	f()
}

//shortcut for panic check
func PanicError(err error) {
	if err != nil {
		panic(err)
	}
}
