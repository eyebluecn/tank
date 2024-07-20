package i18n

import (
	"golang.org/x/text/language"
	"net/http"
)

const (
	LANG_KEY = "_lang"
)

var matcher = language.NewMatcher([]language.Tag{
	// The first language is used as fallback.
	language.English,
	language.Chinese,
})

type Item struct {
	English string
	Chinese string
}

var (
	UsernameOrPasswordCannotNull   = &Item{English: `username or password cannot be null`, Chinese: `用户名或密码不能为空`}
	UsernameOrPasswordError        = &Item{English: `username or password error`, Chinese: `用户名或密码错误`}
	UsernameExist                  = &Item{English: `username "%s" exists`, Chinese: `用户名"%s"已存在`}
	UsernameNotExist               = &Item{English: `username "%s" not exists`, Chinese: `用户名"%s"不存在`}
	UsernameIsNotAdmin             = &Item{English: `username "%s" is not admin user`, Chinese: `用户名"%s"不是管理员账号`}
	UsernameError                  = &Item{English: `username can only be letters, numbers or _`, Chinese: `用户名必填，且只能包含中文，字母，数字和'_'`}
	UserRoleError                  = &Item{English: `user role is not correct`, Chinese: `用户角色设置错误`}
	UserRegisterNotAllowd          = &Item{English: `admin has banned register`, Chinese: `管理员已禁用自主注册`}
	UserPasswordLengthError        = &Item{English: `password at least 6 chars`, Chinese: `密码长度至少为6位`}
	UserOldPasswordError           = &Item{English: `old password error`, Chinese: `旧密码不正确`}
	UserDisabled                   = &Item{English: `user has been disabled`, Chinese: `用户已经被禁用了`}
	MatterDestinationMustDirectory = &Item{English: `destination must be directory'`, Chinese: `目标对象只能是文件夹。`}
	MatterExist                    = &Item{English: `"%s" already exists, invalid operation`, Chinese: `"%s" 已经存在了，操作无效`}
	MatterRecycleBinExist          = &Item{English: `"%s" already exists in recycle bin, invalid operation`, Chinese: `"%s" 已经存在于回收站，请彻底删除后再操作`}
	MatterDepthExceedLimit         = &Item{English: `directory's depth exceed the limit %d > %d`, Chinese: `文件加层数超过限制 %d > %d `}
	MatterNameLengthExceedLimit    = &Item{English: `filename's length exceed the limit %d > %d`, Chinese: `文件名称长度超过限制 %d > %d `}
	MatterSelectNumExceedLimit     = &Item{English: `selected files' num exceed the limit %d > %d`, Chinese: `选择的文件数量超出限制了 %d > %d `}
	MatterSelectSizeExceedLimit    = &Item{English: `selected files' size exceed the limit %s > %s`, Chinese: `选择的文件大小超出限制了 %s > %s `}
	MatterSizeExceedLimit          = &Item{English: `uploaded file's size exceed the size limit %s > %s `, Chinese: `上传的文件超过了限制 %s > %s `}
	MatterSizeExceedTotalLimit     = &Item{English: `file's size exceed the total size limit %s > %s `, Chinese: `上传的文件超过了总大小限制 %s > %s `}
	MatterNameContainSpecialChars  = &Item{English: `file name cannot contain special chars \ / : * ? " < > |"`, Chinese: `名称中不能包含以下特殊符号：\ / : * ? " < > |`}
	MatterMoveRecursive            = &Item{English: `directory cannot be moved to itself or its children`, Chinese: `文件夹不能把自己移入到自己中，也不可以移入到自己的子文件夹下。`}
	MatterNameNoChange             = &Item{English: `filename not change, invalid operation`, Chinese: `文件名没有改变，操作无效！`}
	ShareNumExceedLimit            = &Item{English: `sharing files' num exceed the limit %d > %d`, Chinese: `一次分享的文件数量超出限制了 %d > %d `}
	ShareCodeRequired              = &Item{English: `share code required`, Chinese: `提取码必填`}
	ShareCodeError                 = &Item{English: `share code error`, Chinese: `提取码错误`}
	CronValidateError              = &Item{English: `cron error. five fields needed. eg: 1 * * * *`, Chinese: `Cron表达式错误，必须为5位。例如：1 * * * *`}
	SpaceNameError                 = &Item{English: `space's name can only be letters, numbers or _`, Chinese: `空间名称必填，且只能包含中文，字母，数字和'_'`}
	SpaceNameExist                 = &Item{English: `space's name "%s" exists`, Chinese: `空间名称"%s"已被占用，请使用其他名字`}
	SpaceExclusive                 = &Item{English: `user can only own ONE space`, Chinese: `一个用户只能拥有一个私有空间`}
	SpaceMemberExist               = &Item{English: `space member %s exists`, Chinese: `用户 %s 已经是空间的成员`}
	PermissionDenied               = &Item{English: `permission denied.`, Chinese: `没有操作权限`}
)

func (this *Item) Message(request *http.Request) string {

	if request == nil {
		return this.English
	}

	lang, _ := request.Cookie(LANG_KEY)
	formLangStr := request.FormValue(LANG_KEY)
	acceptLangStr := request.Header.Get("Accept-Language")
	var cookieLangStr string
	if lang != nil {
		cookieLangStr = lang.Value
	}
	tag, _ := language.MatchStrings(matcher, cookieLangStr, formLangStr, acceptLangStr)

	tagBase, _ := tag.Base()
	chineseBase, _ := language.Chinese.Base()

	if tagBase == chineseBase {
		return this.Chinese
	} else {
		return this.English
	}

}
