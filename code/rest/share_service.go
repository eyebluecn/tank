package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"strings"
	"time"
)

//@Service
type ShareService struct {
	BaseBean
	shareDao  *ShareDao
	matterDao *MatterDao
	bridgeDao *BridgeDao
	userDao   *UserDao
}

func (this *ShareService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.shareDao)
	if b, ok := b.(*ShareDao); ok {
		this.shareDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
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
					panic(result.BadRequest("share expired"))
				}
			}
		}
	}
	return share
}

//根据某个shareUuid和code，某个用户是否有权限获取 shareRootUuid 下面的 matterUuid
//如果是根目录下的文件，那么shareRootUuid传root.
func (this *ShareService) ValidateMatter(shareUuid string, code string, user *User, shareRootUuid string, matter *Matter) {

	if matter == nil {
		panic(result.Unauthorized("matter cannot be nil"))
	}

	//如果文件是自己的，那么放行
	if user != nil && matter.UserUuid == user.Uuid {
		return
	}

	if shareUuid == "" || code == "" || shareRootUuid == "" {
		panic(result.Unauthorized("shareUuid,code,shareRootUuid cannot be null"))
	}

	share := this.CheckShare(shareUuid, code, user)

	//如果shareRootUuid是根，那么matterUuid在bridge中应该有记录
	if shareRootUuid == MATTER_ROOT {

		this.bridgeDao.CheckByShareUuidAndMatterUuid(share.Uuid, matter.Uuid)

	} else {
		//验证 shareRootMatter是否在被分享。
		shareRootMatter := this.matterDao.CheckByUuid(shareRootUuid)
		this.bridgeDao.CheckByShareUuidAndMatterUuid(share.Uuid, shareRootMatter.Uuid)

		//保证 puuid对应的matter是shareRootMatter的子文件夹。
		child := strings.HasPrefix(matter.Path, shareRootMatter.Path)
		if !child {
			panic(result.BadRequest("%s is not %s's children", matter.Uuid, shareRootUuid))
		}
	}

}
