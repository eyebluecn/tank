package rest

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nu7hatch/gouuid"
	"regexp"
	"time"
)

//首次运行的时候，将自动安装数据库等内容。
func InstallDatabase() {

	db, err := gorm.Open("mysql", CONFIG.MysqlUrl)
	if err != nil {
		LogPanic(fmt.Sprintf("无法打开%s", CONFIG.MysqlUrl))
	}
	if db != nil {
		defer db.Close()
	}

	//这个方法只会简单查看表是否存在，不会去比照每个字段的。因此如果用户自己修改表结构将会出现不可预测的错误。
	var hasTable = true
	downloadToken := &DownloadToken{}
	hasTable = db.HasTable(downloadToken)
	if !hasTable {

		createDownloadToken := "CREATE TABLE `tank20_download_token` (`uuid` char(36) NOT NULL,`user_uuid` char(36) DEFAULT NULL COMMENT '用户uuid',`matter_uuid` char(36) DEFAULT NULL COMMENT '文件标识',`expire_time` timestamp NULL DEFAULT NULL COMMENT '授权访问的次数',`ip` varchar(45) DEFAULT NULL COMMENT '消费者的ip',`sort` bigint(20) DEFAULT NULL,`modify_time` timestamp NULL DEFAULT NULL,`create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',PRIMARY KEY (`uuid`),UNIQUE KEY `id_UNIQUE` (`uuid`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='下载的token表';"
		db = db.Exec(createDownloadToken)
		if db.Error != nil {
			LogPanic(db.Error)
		}
		LogInfo("创建DownloadToken表")

	}

	matter := &Matter{}
	hasTable = db.HasTable(matter)
	if !hasTable {
		createMatter := "CREATE TABLE `tank20_matter` (`uuid` char(36) NOT NULL,`puuid` varchar(45) DEFAULT NULL COMMENT '上一级的uuid',`user_uuid` char(36) DEFAULT NULL COMMENT '上传的用户id',`dir` tinyint(1) DEFAULT '0' COMMENT '是否是文件夹',`alien` tinyint(1) DEFAULT '0',`name` varchar(255) DEFAULT NULL COMMENT '文件名称',`md5` varchar(45) DEFAULT NULL COMMENT '文件的md5值',`size` bigint(20) DEFAULT '0' COMMENT '文件大小',`privacy` tinyint(1) DEFAULT '0' COMMENT '文件是否是公有的',`path` varchar(255) DEFAULT NULL,`sort` bigint(20) DEFAULT NULL,`modify_time` timestamp NULL DEFAULT NULL,`create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',PRIMARY KEY (`uuid`),UNIQUE KEY `id_UNIQUE` (`uuid`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='file表';"
		db = db.Exec(createMatter)
		if db.Error != nil {
			LogPanic(db.Error)
		}
		LogInfo("创建Matter表")

	}

	preference := &Preference{}
	hasTable = db.HasTable(preference)
	if !hasTable {
		createPreference := "CREATE TABLE `tank20_preference` (`uuid` char(36) NOT NULL,`name` varchar(45) DEFAULT NULL COMMENT '网站名称',`logo_url` varchar(255) DEFAULT NULL,`favicon_url` varchar(255) DEFAULT NULL,`footer_line1` varchar(1024) DEFAULT NULL,`footer_line2` varchar(1024) DEFAULT NULL,`version` varchar(45) DEFAULT NULL,`sort` bigint(20) DEFAULT NULL,`modify_time` timestamp NULL DEFAULT NULL,`create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',PRIMARY KEY (`uuid`),UNIQUE KEY `id_UNIQUE` (`uuid`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='网站偏好设置表';"
		db = db.Exec(createPreference)
		if db.Error != nil {
			LogPanic(db.Error)
		}
		LogInfo("创建Preference表")
	}

	session := &Session{}
	hasTable = db.HasTable(session)
	if !hasTable {

		createSession := "CREATE TABLE `tank20_session` (`uuid` char(36) NOT NULL,`authentication` char(36) DEFAULT NULL COMMENT '认证身份，存放在cookie中',`user_uuid` char(36) DEFAULT NULL COMMENT '用户uuid',`ip` varchar(45) DEFAULT NULL COMMENT '用户的ip地址',`expire_time` timestamp NULL DEFAULT NULL,`sort` bigint(20) DEFAULT NULL,`modify_time` timestamp NULL DEFAULT NULL,`create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',PRIMARY KEY (`uuid`),UNIQUE KEY `id_UNIQUE` (`uuid`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='session表';"
		db = db.Exec(createSession)
		if db.Error != nil {
			LogPanic(db.Error)
		}
		LogInfo("创建Session表")
	}

	uploadToken := &UploadToken{}
	hasTable = db.HasTable(uploadToken)
	if !hasTable {

		createUploadToken := "CREATE TABLE `tank20_upload_token` (`uuid` char(36) NOT NULL,`user_uuid` char(36) DEFAULT NULL COMMENT '用户uuid',`folder_uuid` char(36) DEFAULT NULL,`matter_uuid` char(36) DEFAULT NULL,`filename` varchar(255) DEFAULT NULL COMMENT '文件后缀名的过滤，可以只允许用户上传特定格式的文件。',`privacy` tinyint(1) DEFAULT '1',`size` bigint(20) DEFAULT '0',`expire_time` timestamp NULL DEFAULT NULL,`ip` varchar(45) DEFAULT NULL COMMENT '消费者的ip',`sort` bigint(20) DEFAULT NULL,`modify_time` timestamp NULL DEFAULT NULL,`create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',PRIMARY KEY (`uuid`),UNIQUE KEY `id_UNIQUE` (`uuid`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='上传的token表';"
		db = db.Exec(createUploadToken)
		if db.Error != nil {
			LogPanic(db.Error)
		}
		LogInfo("创建UploadToken表")
	}

	user := &User{}
	hasTable = db.HasTable(user)
	if !hasTable {

		//验证超级管理员的信息
		if m, _ := regexp.MatchString(`^[0-9a-zA-Z_]+$`, CONFIG.AdminUsername); !m {
			LogPanic(`超级管理员用户名必填，且只能包含字母，数字和'_''`)
		}

		if len(CONFIG.AdminPassword) < 6 {
			LogPanic(`超级管理员密码长度至少为6位`)
		}

		if CONFIG.AdminEmail == "" {
			LogPanic("超级管理员邮箱必填！")
		}

		createUser := "CREATE TABLE `tank20_user` (`uuid` char(36) NOT NULL,`role` varchar(45) DEFAULT 'USER',`username` varchar(255) DEFAULT NULL COMMENT '昵称',`password` varchar(255) DEFAULT NULL COMMENT '密码',`email` varchar(45) DEFAULT NULL COMMENT '邮箱',`phone` varchar(45) DEFAULT NULL COMMENT '电话',`gender` varchar(45) DEFAULT 'UNKNOWN' COMMENT '性别，默认未知',`city` varchar(45) DEFAULT NULL COMMENT '城市',`avatar_url` varchar(255) DEFAULT NULL COMMENT '头像链接',`last_time` datetime DEFAULT NULL COMMENT '上次登录使劲按',`last_ip` varchar(45) DEFAULT NULL,`size_limit` int(11) DEFAULT '-1' COMMENT '该账号上传文件的大小限制，单位byte。<0 表示不设限制',`status` varchar(45) DEFAULT 'OK',`sort` bigint(20) DEFAULT NULL,`modify_time` timestamp NULL DEFAULT NULL,`create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',PRIMARY KEY (`uuid`),UNIQUE KEY `id_UNIQUE` (`uuid`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户表描述';"
		db = db.Exec(createUser)
		if db.Error != nil {
			LogPanic(db.Error)
		}
		LogInfo("创建User表")

		user := &User{}
		timeUUID, _ := uuid.NewV4()
		user.Uuid = string(timeUUID.String())
		user.CreateTime = time.Now()
		user.ModifyTime = time.Now()
		user.LastTime = time.Now()
		user.Sort = time.Now().UnixNano() / 1e6
		user.Role = USER_ROLE_ADMINISTRATOR
		user.Username = CONFIG.AdminUsername
		user.Password = GetBcrypt(CONFIG.AdminPassword)
		user.Email = CONFIG.AdminEmail
		user.Phone = ""
		user.Gender = USER_GENDER_UNKNOWN
		user.SizeLimit = -1
		user.Status = USER_STATUS_OK

		db.Create(user)

	}

}
