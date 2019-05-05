package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"net/http"
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

func (this *ShareService) Detail(uuid string) *Share {

	share := this.shareDao.CheckByUuid(uuid)

	return share
}

// check whether shareUuid and shareCode matches. check whether user can access.
func (this *ShareService) CheckShare(request *http.Request, shareUuid string, code string, user *User) *Share {

	share := this.shareDao.CheckByUuid(shareUuid)
	//if self, not need shareCode
	if user == nil || user.Uuid != share.UserUuid {
		//if not login or not self's share, shareCode is required.
		if code == "" {
			panic(result.CustomWebResultI18n(request, result.NEED_SHARE_CODE, i18n.ShareCodeRequired))
		} else if share.Code != code {
			panic(result.CustomWebResultI18n(request, result.SHARE_CODE_ERROR, i18n.ShareCodeError))
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

//check whether a user can access a matter. shareRootUuid is matter's parent(or parent's parent and so on)
func (this *ShareService) ValidateMatter(request *http.Request, shareUuid string, code string, user *User, shareRootUuid string, matter *Matter) {

	if matter == nil {
		panic(result.Unauthorized("matter cannot be nil"))
	}

	//if self's matter, ok.
	if user != nil && matter.UserUuid == user.Uuid {
		return
	}

	if shareUuid == "" || code == "" || shareRootUuid == "" {
		panic(result.Unauthorized("shareUuid,code,shareRootUuid cannot be null"))
	}

	share := this.CheckShare(request, shareUuid, code, user)

	//if shareRootUuid is root. Bridge must has record.
	if shareRootUuid == MATTER_ROOT {

		this.bridgeDao.CheckByShareUuidAndMatterUuid(share.Uuid, matter.Uuid)

	} else {
		//check whether shareRootMatter is being sharing
		shareRootMatter := this.matterDao.CheckByUuid(shareRootUuid)
		this.bridgeDao.CheckByShareUuidAndMatterUuid(share.Uuid, shareRootMatter.Uuid)

		// shareRootMatter is ancestor of matter.
		child := strings.HasPrefix(matter.Path, shareRootMatter.Path)
		if !child {
			panic(result.BadRequest("%s is not %s's children", matter.Uuid, shareRootUuid))
		}
	}

}
