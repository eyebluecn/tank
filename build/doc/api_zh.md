
# 蓝眼云盘api接口




## 一、实体

在详细介绍各controller中的接口前，有必要先介绍蓝眼云盘中的各实体，所有的实体基类均为`Base`

#### Base

`Base`定义如下，所有会在数据库中持久化的实体均会继承`Base`，`Controller`在返回实体给前端时，会将字段和值序列化成json字符串，其中键就和每个实体字段后面的`json`标签一致，下文也会有详细例子介绍

```
type Base struct {
    //唯一标识
	Uuid       string    `gorm:"primary_key" json:"uuid"`
	//排序用的字段，一般是时间戳来表示序号先后
	Sort       int64     `json:"sort"`
	//修改时间
	ModifyTime time.Time `json:"modifyTime"`
	//创建时间
	CreateTime time.Time `json:"createTime"`
}
```

#### Preference

`Preference`是整个网站的偏好设置，网站的名称，logo，favicon，版权，备案号等信息均由这个实体负责。定义如下：

```
type Preference struct {
    //继承Base，也就是说Base中的Uuid,Sort,ModifyTime,CreateTime这里也会有
	Base
	//网站名称
	Name        string `json:"name"`
	//网站的logo url
	LogoUrl     string `json:"logoUrl"`
	//网站的favicon.ico url
	FaviconUrl  string `json:"faviconUrl"`
	//网站下方第一行信息，一般是版权信息
	FooterLine1 string `json:"footerLine1"`
	//网站下方的第二行信息，一般是本案信息
	FooterLine2 string `json:"footerLine2"`
	//当前运行的蓝眼博客版本，这个字段不可修改，每次发版时硬编码
	Version     string `json:"version"`
}
```


#### Matter

`Matter`是代表文件（文件夹是一种特殊的文件），为了和系统的`file`重复，这里使用`matter`，这个实体是蓝眼云盘最重要的实体：

```
type Matter struct {
    //继承Base，也就是说Base中的Uuid,Sort,ModifyTime,CreateTime这里也会有
	Base
	//所在的文件夹的uuid，如果在根目录下，这个字段为 root
	Puuid    string  `json:"puuid"`
	//创建这个文件的用户uuid
	UserUuid string  `json:"userUuid"`
	//该文件是否是文件夹
	Dir      bool    `json:"dir"`
	//该文件是否是应用文件，比如用户上传的头像这个字段就为true. 一般上传的文件就为false
	Alien    bool    `json:"alien"`
	//文件名，带后缀名。例如：avatar.jpg
	Name     string  `json:"name"`
    //文件的md5值，目前该功能尚未实现，作为保留字段
	Md5      string  `json:"md5"`
	//文件大小，单位 byte。比如某个文件1M大，那么这里就为： 1048576
	Size     int64   `json:"size"`
	//文件是否为私有，如果true则该文件只能作者或超级管理员可以下载，如果false所有人均可以通过下载链接下载。
	Privacy  bool    `json:"privacy"`
	//文件在磁盘中的路径，前端无需关心这个字段。但是后端在寻找文件时这个字段非常关键。
	Path     string  `json:"path"`
	//该文件的父级matter，该字段不会持久化到数据集，属于获取matter详情时临时组装出来的。
	Parent   *Matter `gorm:"-" json:"parent"`
}
```


#### User

`User`是代表用户：

```
type User struct {
    //继承Base，也就是说Base中的Uuid,Sort,ModifyTime,CreateTime这里也会有
	Base
	//角色，有以下枚举值：GUEST(游客，不会持久化到数据库),USER(普通用户),ADMINISTRATOR(超级管理员)
	Role      string    `json:"role"`
	//用户名，在Matter的path字段中很有用。
	Username  string    `json:"username"`
	//密码，默认不会返回给前端
	Password  string    `json:"-"`
	//邮箱，作为登录时的账户
	Email     string    `json:"email"`
	//手机
	Phone     string    `json:"phone"`
	//性别，有以下枚举值：MALE(男性),FEMALE(女性),UNKNOWN(未知)
	Gender    string    `json:"gender"`
	//城市
	City      string    `json:"city"`
	//头像Url
	AvatarUrl string    `json:"avatarUrl"`
	//上次登录时的ip
	LastIp    string    `json:"lastIp"`
	//上次登录的时间
	LastTime  time.Time `json:"lastTime"`
	//该用户允许上传的单文件最大大小
	SizeLimit int64     `json:"sizeLimit"`
	//状态，有以下枚举值：OK(正常),DISABLED(被禁用)
	Status    string    `json:"status"`
}
```

### UploadToken 

用于给陌生人上传的token

```
type UploadToken struct {
    //继承Base，也就是说Base中的Uuid,Sort,ModifyTime,CreateTime这里也会有
	Base
	//颁发该token的用户，系统中任何用户都能颁发token
	UserUuid   string    `json:"userUuid"`
	//使用这个token上传文件就必须上传在这个文件夹下
	FolderUuid string    `json:"folderUuid"`
	//陌生人上传好了的文件uuid
	MatterUuid string    `json:"matterUuid"`
	//过期时间
	ExpireTime time.Time `json:"expireTime"`
	//使用这个token上传文件就必须是这个文件名
	Filename   string    `json:"filename"`
	//使用这个token上传文件就必须是这个公私有性
	Privacy    bool      `json:"privacy"`
	//使用这个token上传文件就必须这个大小
	Size       int64     `json:"size"`
	//使用这个token上传文件陌生人的ip
	Ip         string    `json:"ip"`
}
```


### DownloadToken 

用于给陌生人下载的token，一个matter如果Privacy=true，那么就意味着只有自己或者超级管理员可以下载，如果让某些自己信任的用户也能下载，那么就需要生成`DownloadToken`给这些用户来下载。

```
type DownloadToken struct {
    //继承Base，也就是说Base中的Uuid,Sort,ModifyTime,CreateTime这里也会有
	Base
	//颁发该token的用户
	UserUuid   string    `json:"userUuid"`
	//该token只能下载这个文件
	MatterUuid string    `json:"matterUuid"`
	//有效期截止
	ExpireTime time.Time `json:"expireTime"`
	//下载者的ip
	Ip         string    `json:"ip"`
}

```

### Pager

在前端请求一个列表时，通常返回的都是一个`Pager`，`Pager`中就是装的各个实体的列表。

```
type Pager struct {
    //当前页数，0基
	Page       int         `json:"page"`
	//每页的大小
	PageSize   int         `json:"pageSize"`
	//总的条目数
	TotalItems int         `json:"totalItems"`
	//总的页数
	TotalPages int         `json:"totalPages"`
	//实体的数组
	Data       interface{} `json:"data"`
}
```

### WebResult

`WebResult`并不是会持久化到数据库中实体，`WebResult`是在`controller`返回数据给前端时包装的一层，有了`WebResult`后每个接口返回的数据会更加统一，方便了前端的统一处理。

```
type WebResult struct {
    //状态码，具体每个码的意义参考下文
	Code int       `json:"code"`
	//一句话描述请求结果，通常会是出错时指明出错原因，或者修改权限等小操作时提示的`操作成功`
	Msg  string      `json:"msg"`
	//内容可能是一个实体，也可能是一个 Pager.
	Data interface{} `json:"data"`
}

```

状态码对应关系如下：

```
const (
	//正常
	RESULT_CODE_OK = 200
	//未登录
	RESULT_CODE_LOGIN = -400
	//没有权限
	RESULT_CODE_UNAUTHORIZED = -401
	//请求错误
	RESULT_CODE_BAD_REQUEST = -402
	//没有找到
	RESULT_CODE_NOT_FOUND = -404
	//登录过期
	RESULT_CODE_LOGIN_EXPIRED = -405
	//该登录用户不是有效用户
	RESULT_CODE_LOGIN_INVALID = -406
	//提交的表单验证不通过
	RESULT_CODE_FORM_INVALID = -410
	//请求太频繁
	RESULT_CODE_FREQUENCY = -420
	//服务器出错。
	RESULT_CODE_SERVER_ERROR = -500
	//远程服务不可用
	RESULT_CODE_NOT_AVAILABLE = -501
	//并发异常
	RESULT_CODE_CONCURRENCY = -511
	//远程微服务没有找到
	RESULT_CODE_SERVICE_NOT_FOUND = -600
	//远程微服务连接超时
	RESULT_CODE_SERVICE_TIME_OUT = -610
	//通用的异常
	RESULT_CODE_UTIL_EXCEPTION = -700
)
```

## 二、返回规范

蓝眼云盘采用前后端分离的模式，前端调用后端接口时，url均以`/api`开头，返回均是json字符串。

- 返回的json字符串的key均为小写开头的驼峰法，具体参考实体类中`json`标签

- 返回的时间格式均为 `YYYY-MM-dd HH:mm:ss`（例如：2018-01-06 17:57:00）

返回内容均是由`WebResult`进行包装，因此具有高度的统一性，在这里我们约定一些说法，后面介绍`Controller`时便不再赘述。

1. 返回一个`XX`实体

    指的是`WebResult`的`Code=200`, `Data=一个XX实体对象`。
    
    例：返回一个User，则前端会收到以下json字符串：
    ```
    {
      "code": 200,
      "msg": "",
      "data": {
        "uuid": "eed2c66d-1de6-47ff-645e-b67beaa10365",
        "sort": 1514803034507,
        "modifyTime": "2018-01-06 18:00:58",
        "createTime": "2018-01-01 18:37:15",
        "role": "USER",
        "username": "demo",
        "email": "demo@tank.eyeblue.cn",
        "phone": "",
        "gender": "UNKNOWN",
        "city": "",
        "avatarUrl": "/api/alien/download/ea490cb6-368e-436d-71c0-fcfb08854c80/1180472.png",
        "lastIp": "124.78.220.82",
        "lastTime": "2018-01-06 18:00:58",
        "sizeLimit": 1048576,
        "status": "OK"
      }
    }
    ```

2. 返回`XX`的`Pager`

    指的是`WebResult`的`Code=200`, `Data=XX的Pager`。
    
    例：返回`User`的`Pager`，则前端会收到以下json字符串：
    ```
    {
      "code": 200,
      "msg": "",
      "data": {
        "page": 0,
        "pageSize": 10,
        "totalItems": 2,
        "totalPages": 1,
        "data": [
          {
            "uuid": "6a661938-8289-4957-4096-5a1b584bf371",
            "sort": 1515057859613,
            "modifyTime": "2018-01-04 17:26:01",
            "createTime": "2018-01-04 17:24:20",
            "role": "ADMINISTRATOR",
            "username": "simba",
            "email": "y******5@163.com",
            "phone": "",
            "gender": "MALE",
            "city": "",
            "avatarUrl": "/api/alien/download/d1e453cb-3170-4bdb-73f2-fa0372ee017b/1180480.png",
            "lastIp": "180.173.103.207",
            "lastTime": "2018-01-04 17:26:01",
            "sizeLimit": -1,
            "status": "OK"
          },
          {
            "uuid": "e59be6a3-f806-463e-553a-4c5892eedf78",
            "sort": 1514881002975,
            "modifyTime": "2018-01-02 16:16:43",
            "createTime": "2018-01-02 16:16:43",
            "role": "USER",
            "username": "blog_dev",
            "email": "b*****@tank.eyeblue.cn",
            "phone": "",
            "gender": "MALE",
            "city": "",
            "avatarUrl": "/api/alien/download/fdca6eee-d009-4eb3-5ad4-15ba3701cb2e/jump.jpg",
            "lastIp": "",
            "lastTime": "2018-01-02 16:16:43",
            "sizeLimit": 1048576,
            "status": "OK"
          }
        ]
      }
    }
    ```
3. 返回错误信息：yyy

    指的是`WebResult`的`Code=-400`, `Msg=yyy`。(这里的Code具体值参考上文的code表)
    
    例：返回错误信息："【新建文件夹】已经存在了，请使用其他名称。"，则前端会收到以下json字符串：
    ```
    {
      "code": -700,
      "msg": "【新建文件夹】已经存在了，请使用其他名称。",
      "data": null
    }
    ```
    
4. 返回成功信息：zzz

    指的是`WebResult`的`Code=200`, `Msg=zzz`。(这里的Code具体值参考上文的code表)
    
    例：返回成功信息："删除成功。"，则前端会收到以下json字符串：
    ```
    {
      "code": 200,
      "msg": "删除成功。",
      "data": null
    }
    ```


## 三、接口

蓝眼云盘所有的接口均定义在`controller`中，总共定义了以下`controller`：

名称 | 所在文件 | 描述
--------- | ---- | -----------
PreferenceController | `preference_controller.go` | 网站标题，logo，版权说明等信息的增删改查
MatterController | `matter_controller.go` | 站内创建文件夹，上传文件，删除文件，修改权限等
UserController | `user_controller.go` | 登录，管理操作站内用户
AlienController | `alien_controller.go` | 第三方授权上传，下载，预处理

每个接口都有不同的访问级别，系统中定义了三种访问级别，分别是：

`游客` < `注册用户` < `管理员`

### PreferenceController

该Controller负责网站中的偏好设置，主要操作`Preference`实体

----------

#### /api/preference/fetch

**功能**：读取网站偏好设置，网站名称,logo,版权,备案信息均从此接口读取。

**访问级别**：`游客`,`注册用户`,`管理员`

**请求参数**：无

**返回**: 一个`Preference`实体


----------


#### /api/preference/edit

**功能**：编辑网站偏好设置，修改网站名称,logo,版权,备案信息。

**访问级别**：`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
name | `string` | 必填 | 网站名称
logoUrl | `string` | 选填 | 网站logoUrl,如果不填默认使用蓝眼云盘logo
faviconUrl | `string` | 选填 | 网站faviconUrl,如果不填默认使用蓝眼云盘favicon.ico
footerLine1 | `string` | 选填 | 网站版权所有信息
footerLine2 | `string` | 选填 | 网站备案信息

**返回**: 一个`Preference`实体

----------

### MatterController

该Controller负责站内创建文件夹，上传文件，删除文件，修改权限等，主要操作`Matter`实体

----------

#### /api/matter/create/directory

**功能**：创建文件夹

**访问级别**：`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
puuid | `string` | 必填 | 准备创建的目录所在的目录，如果在根目录下创建传`root`
name | `string` | 必填 | 文件夹名称， 不能包含以下特殊符号：`< > \| * ? / \`

**返回**: 新建的这个文件夹的`Matter`实体

----------


#### /api/matter/upload

**功能**：上传文件

**访问级别**：`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
userUuid | `string` | 选填 | 该文件的归属人，如果不填则使用当前登录的用户uuid；如果是管理员要给用户A上传文件，那么这里应该填A的uuid.
alien | `bool` | 选填 | 是否是应用数据，true表示应用数据,这时会自动放置在`应用数据`这个文件夹下，比如上传头像。false表示非应用数据，这时`puuid`字段为必填
puuid | `bool` | 选填 | 文件上传到哪个目录下。 如果`alien`=false，这个字段必填。
privacy | `bool` | 选填 | 文件的私有性，默认`false`
file | `file` | 必填 | 文件，在浏览器中是通过`<input type="file" name="file"/>`来选择的


**返回**: 刚上传的这个文件的`Matter`实体

----------


#### /api/matter/delete

**功能**：删除文件或者文件夹

**访问级别**：`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
uuid | `string` | 必填 | 待删除的文件或文件夹的uuid


**返回**: 成功信息“删除成功”

----------


#### /api/matter/delete/batch

**功能**：批量删除文件或文件夹

**访问级别**：`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
uuids | `string` | 必填 | 待删除的文件或文件夹的uuids,用逗号(,)分隔。

**返回**: 成功信息“删除成功”

----------

#### /api/matter/rename

**功能**：重命名文件或文件夹

**访问级别**：`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
uuid | `string` | 必填 | 文件的uuid
name | `string` | 必填 | 新名字，不能包含以下特殊符号：`< > \| * ? / \`

**返回**: 刚重命名的这个文件的`Matter`实体

----------
#### /api/matter/change/privacy

**功能**：改变文件的公私有属性

**访问级别**：`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
uuid | `string` | 必填 | 文件的uuid
privacy | `bool` | 选填 | 文件的私有性，默认`false`

**返回**: 成功信息“设置成功”

----------

#### /api/matter/move

**功能**：将一个文件夹或者文件移入到另一个文件夹下。

**访问级别**：`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
srcUuids | `string` | 必填 | 待移动的文件或文件夹的uuids,用逗号(,)分隔。
destUuid | `string` | 必填 | 目标文件夹，根目录用`root`

**返回**: 成功信息“设置成功”

----------


#### /api/matter/detail

**功能**：产看文件详情

**访问级别**：`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
uuid | `string` | 必填 | 该文件的uuid。
destUuid | `string` | 必填 | 目标文件夹，根目录用`root`

**返回**: 这个文件的`Matter`实体

----------

#### /api/matter/page

**功能**：按照分页的方式获取某个文件夹下文件和子文件夹的列表

**访问级别**：`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
puuid | `string` | 选填 | 文件夹uuid，如果根目录填`root`
page | `int` | 选填 | 当前页数，0基，默认0
pageSize | `int` | 选填 | 每页条目数，默认200
userUuid | `string` | 选填 | 筛选文件拥有者，对于普通用户使用当前登录的用户uuid.
name | `string` | 选填 | 模糊筛选文件名 
dir | `bool` | 选填 | 筛选是否为文件夹
orderDir | `DESC`或`ASC` | 选填 | 按文件夹排序，`DESC`降序排，`ASC`升序排
orderCreateTime | `DESC`或`ASC` | 选填 | 按创建时间排序，`DESC`降序排，`ASC`升序排
orderSize | `DESC`或`ASC` | 选填 | 按文件大小排序，`DESC`降序排，`ASC`升序排
orderName | `DESC`或`ASC` | 选填 | 按名称排序，`DESC`降序排，`ASC`升序排
extensions | `string` | 选填 | 按文件后缀名筛选，逗号(,)分隔。例：`jpg,png,pdf`


**返回**: `Matter`的`Pager`

----------

#### /api/alien/download/{uuid}/{filename}

这个接口的定义在在`AlienController`中的，具体使用请参考[蓝眼云盘编程接口](https://github.com/eyebluecn/tank/blob/master/build/doc/alien_zh.md) 这里只做简单的下载使用。

**功能**：下载文件。

**访问级别**：`游客`,`注册用户`,`管理员`

**请求参数**： 均是放置在url中

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
uuid | `string` | 必填 | 要下载的文件uuid
filename | `string` | 必填 | 要下载的文件名

**返回**: 二进制的文件

----------

### UserController

该Controller负责站内创建文件夹，上传文件，删除文件，修改权限等，主要操作`Matter`实体

----------

#### /api/user/create

**功能**：创建用户

**访问级别**：`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
username | `string` | 必填 | 用户名，且只能包含字母，数字和'_'
password | `string` | 必填 | 密码，密码长度至少为6位
email | `string` | 必填 | 邮箱，作为用户登录的凭证
avatarUrl | `string` | 选填 | 头像
phone | `string` | 选填 | 电话
gender | `string` | 选填 | 性别
role | `string` | 选填 | 角色
city | `string` | 选填 | 城市
sizeLimit | `string` | 选填 | 用户上传单文件限制，单位byte. 如果负数表示无限制。


**返回**: 新建的`User`实体

----------

#### /api/user/edit

**功能**：编辑用户

**访问级别**：`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
uuid | `string` | 必填 | 待编辑的用户uuid
avatarUrl | `string` | 选填 | 头像
phone | `string` | 选填 | 电话
gender | `string` | 选填 | 性别
role | `string` | 选填 | 角色
city | `string` | 选填 | 城市
sizeLimit | `string` | 选填 | 用户上传单文件限制，单位byte. 如果负数表示无限制。


**返回**: 编辑的`User`实体

----------

#### /api/user/change/password

**功能**：用户修改密码

**访问级别**：`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
oldPassword | `string` | 必填 | 旧密码
newPassword | `string` | 必填 | 新密码

**返回**: 修改密码的`User`实体

----------


#### /api/user/change/password

**功能**：管理员重置用户密码

**访问级别**：`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
userUuid | `string` | 必填 | 待重置密码的用户uuid
password | `string` | 必填 | 密码

**返回**: 修改密码的`User`实体

----------

#### /api/user/login

**功能**：登录

**访问级别**：`游客`,`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
email | `string` | 必填 | 邮箱
password | `string` | 必填 | 密码

**返回**: 当前登录的`User`实体

----------

#### /api/user/logout

**功能**：登录

**访问级别**：`注册用户`,`管理员`

**请求参数**：无

**返回**: 成功信息"退出成功！"

----------

#### /api/user/detail

**功能**：查看用户详情

**访问级别**：`游客`,`注册用户`,`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
uuid | `string` | 必填 | 待查看的用户uuid

**返回**: `User`实体

----------

#### /api/user/page

**功能**：查看用户列表

**访问级别**：`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
page | `int` | 选填 | 当前页数，0基，默认0
pageSize | `int` | 选填 | 每页条目数，默认200
username | `string` | 选填 | 模糊筛选用户名
email | `string` | 选填 | 模糊筛选邮箱
phone | `string` | 选填 | 模糊筛选电话
orderLastTime | `DESC`或`ASC` | 选填 | 按上次登录时间排序，`DESC`降序排，`ASC`升序排
orderCreateTime | `DESC`或`ASC` | 选填 | 按创建时间排序，`DESC`降序排，`ASC`升序排

**返回**: `User`实体的`Pager`

----------

#### /api/user/disable

**功能**：禁用用户

**访问级别**：`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
uuid | `string` | 必填 | 待操作的用户

**返回**: 被操作的用户`User`实体

----------

#### /api/user/enable

**功能**：启用用户

**访问级别**：`管理员`

**请求参数**：

名称 | 类型 | 必填性 | 描述
--------- | ---- | ---- | -----------
uuid | `string` | 必填 | 待操作的用户

**返回**: 被操作的用户`User`实体

----------

### AlienController

这个Controller主要暴露给第三方使用，详情请参考：[蓝眼云盘编程接口](https://github.com/eyebluecn/tank/blob/master/build/doc/alien_zh.md)

