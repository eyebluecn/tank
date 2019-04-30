package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"time"
)

//@Service
type ShareService struct {
	BaseBean
	shareDao *ShareDao
	userDao  *UserDao
}

//初始化方法
func (this *ShareService) Init() {
	this.BaseBean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := core.CONTEXT.GetBean(this.shareDao)
	if b, ok := b.(*ShareDao); ok {
		this.shareDao = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

//获取某个分享的详情。
func (this *ShareService) Detail(uuid string) *Share {

	share := this.shareDao.CheckByUuid(uuid)

	return share
}

//验证一个shareUuid和shareCode是否匹配和有权限。
func (this *ShareService) CheckShare(shareUuid string, code string, user *User) *Share {

	share := this.shareDao.CheckByUuid(shareUuid)
	//如果是自己的分享，可以不要提取码
	if user == nil || user.Uuid != share.UserUuid {
		//没有登录，或者查看的不是自己的分享，要求有验证码
		if code == "" {
			panic(result.CustomWebResult(result.NEED_SHARE_CODE, "提取码必填"))
		} else if share.Code != code {
			panic(result.CustomWebResult(result.SHARE_CODE_ERROR, "提取码错误"))
		} else {
			if !share.ExpireInfinity {
				if share.ExpireTime.Before(time.Now()) {
					panic(result.BadRequest("分享已过期"))
				}
			}
		}
	}

	return share
}
