package rest

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	//root matter's uuid
	MATTER_ROOT = "root"
	//cache directory name.
	MATTER_CACHE = "cache"
	//zip file temp directory.
	MATTER_ZIP             = "zip"
	MATTER_NAME_MAX_LENGTH = 200
	MATTER_NAME_MAX_DEPTH  = 32
	//matter name pattern
	MATTER_NAME_PATTERN = `[\\/:*?"<>|]`
)

/**
 * file is too common. so we use matter as file.
 */
type Matter struct {
	Uuid       string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort       int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	Puuid      string    `json:"puuid" gorm:"type:char(36);index:idx_matter_puuid"` //index should unique globally.
	UserUuid   string    `json:"userUuid" gorm:"type:char(36);index:idx_matter_uu"`
	Username   string    `json:"username" gorm:"type:varchar(45) not null"`
	Dir        bool      `json:"dir" gorm:"type:tinyint(1) not null;default:0"`
	Name       string    `json:"name" gorm:"type:varchar(255) not null"`
	Md5        string    `json:"md5" gorm:"type:varchar(45)"`
	Size       int64     `json:"size" gorm:"type:bigint(20) not null;default:0"`
	Privacy    bool      `json:"privacy" gorm:"type:tinyint(1) not null;default:0"`
	Path       string    `json:"path" gorm:"type:varchar(1024)"`
	Times      int64     `json:"times" gorm:"type:bigint(20) not null;default:0"`
	Parent     *Matter   `json:"parent" gorm:"-"`
	Children   []*Matter `json:"-" gorm:"-"`
	Prop       string    `json:"prop" gorm:"type:varchar(1024) not null;default:'{}'"`
	VisitTime  time.Time `json:"visitTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	Deleted    bool      `json:"deleted" gorm:"type:tinyint(1) not null;index:idx_matter_del;default:0"`
	DeleteTime time.Time `json:"deleteTime" gorm:"type:timestamp not null;index:idx_matter_delt;default:'2018-01-01 00:00:00'"`
}

// get matter's absolute path. the Path property is relative path in db.
func (this *Matter) AbsolutePath() string {
	return GetUserMatterRootDir(this.Username) + this.Path
}

func (this *Matter) MimeType() string {
	return util.GetMimeType(util.GetExtension(this.Name))
}

//Create a root matter. It's convenient for copy and move
func NewRootMatter(user *User) *Matter {
	matter := &Matter{}
	matter.Uuid = MATTER_ROOT
	matter.UserUuid = user.Uuid
	matter.Username = user.Username
	matter.Dir = true
	matter.Path = ""
	matter.CreateTime = user.CreateTime
	matter.UpdateTime = user.UpdateTime
	matter.VisitTime = user.UpdateTime

	return matter
}

//get user's space absolute path
func GetUserSpaceRootDir(username string) (rootDirPath string) {

	rootDirPath = fmt.Sprintf("%s/%s", core.CONFIG.MatterPath(), username)

	return rootDirPath
}

//get user's root absolute path
func GetUserMatterRootDir(username string) (rootDirPath string) {

	rootDirPath = fmt.Sprintf("%s/%s/%s", core.CONFIG.MatterPath(), username, MATTER_ROOT)

	return rootDirPath
}

//get user's cache absolute path
func GetUserCacheRootDir(username string) (rootDirPath string) {

	rootDirPath = fmt.Sprintf("%s/%s/%s", core.CONFIG.MatterPath(), username, MATTER_CACHE)

	return rootDirPath
}

//get user's zip absolute path
func GetUserZipRootDir(username string) (rootDirPath string) {

	rootDirPath = fmt.Sprintf("%s/%s/%s", core.CONFIG.MatterPath(), username, MATTER_ZIP)

	return rootDirPath
}

//check matter's name. If error, panic.
func CheckMatterName(request *http.Request, name string) string {

	if name == "" {
		panic(result.BadRequest("name cannot be null"))
	}
	if strings.HasPrefix(name, " ") || strings.HasSuffix(name, " ") {
		panic(result.BadRequest("name cannot start with or end with space"))
	}
	if m, _ := regexp.MatchString(MATTER_NAME_PATTERN, name); m {
		panic(result.BadRequestI18n(request, i18n.MatterNameContainSpecialChars))
	}

	if len(name) > MATTER_NAME_MAX_LENGTH {
		panic(result.BadRequestI18n(request, i18n.MatterNameLengthExceedLimit, len(name), MATTER_NAME_MAX_LENGTH))
	}
	return name
}

//fetch the props
func (this *Matter) FetchPropMap() map[string]string {

	m := make(map[string]string)
	json := this.Prop
	if json == "" {
		json = EMPTY_JSON_MAP
	}

	err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(json), &m)
	if err != nil {
		panic(err)
	}

	return m
}

//fetch the props
func (this *Matter) SetPropMap(propMap map[string]string) {

	b, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(propMap)
	if err != nil {
		panic(err)
	}

	this.Prop = string(b)
}
