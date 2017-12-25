package rest

//这里是用来定义系统中接口的访问级别的，不同角色将对不同feature具有访问权限。
const (
	//公共接口，所有人均可访问。
	FEATURE_PUBLIC = "FEATURE_TYPE_PUBLIC"
	//管理用户，只有超级管理员可以访问。
	FEATURE_USER_MANAGE = "FEATURE_USER_MANAGE"
	//管理文件，只有超级管理员可以访问。
	FEATURE_MATTER_MANAGE = "FEATURE_MATTER_MANAGE"
	//查看自己资料，普通用户和超级管理员可以访问。
	FEATURE_USER_MINE = "FEATURE_USER_MINE"
	//其他，超级管理员可以访问。
	FEATURE_OTHER = "FEATURE_OTHER"
)
