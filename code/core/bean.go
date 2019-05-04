package core

/**
 * bean interface means singleton in application
 */
type Bean interface {
	//init the bean when constructing
	Init()
	//cleanup the bean when system's cleanup
	Cleanup()
	//when everything(including db's connection) loaded, this method will be invoked.
	Bootstrap()
	//shortcut for panic check.
	PanicError(err error)
}
