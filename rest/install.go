package rest

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

//首次运行的时候，将自动安装数据库等内容。
func InstallDatabase() {

	db, err := gorm.Open("mysql", CONFIG.MysqlUrl)
	if err != nil {
		LOGGER.Panic(fmt.Sprintf("无法打开%s", CONFIG.MysqlUrl))
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
			LOGGER.Panic(db.Error.Error())
		}
		LOGGER.Info("创建DownloadToken表")

	}

	matter := &Matter{}
	hasTable = db.HasTable(matter)
	if !hasTable {
		createMatter := "CREATE TABLE `tank20_matter` (`uuid` char(36) NOT NULL,`puuid` varchar(45) DEFAULT NULL COMMENT '上一级的uuid',`user_uuid` char(36) DEFAULT NULL COMMENT '上传的用户id',`dir` tinyint(1) DEFAULT '0' COMMENT '是否是文件夹',`alien` tinyint(1) DEFAULT '0',`name` varchar(255) DEFAULT NULL COMMENT '文件名称',`md5` varchar(45) DEFAULT NULL COMMENT '文件的md5值',`size` bigint(20) DEFAULT '0' COMMENT '文件大小',`privacy` tinyint(1) DEFAULT '0' COMMENT '文件是否是公有的',`path` varchar(255) DEFAULT NULL,`sort` bigint(20) DEFAULT NULL,`modify_time` timestamp NULL DEFAULT NULL,`create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',PRIMARY KEY (`uuid`),UNIQUE KEY `id_UNIQUE` (`uuid`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='file表';"
		db = db.Exec(createMatter)
		if db.Error != nil {
			LOGGER.Panic(db.Error.Error())
		}
		LOGGER.Info("创建Matter表")

	}

	preference := &Preference{}
	hasTable = db.HasTable(preference)
	if !hasTable {
		createPreference := "CREATE TABLE `tank20_preference` (`uuid` char(36) NOT NULL,`name` varchar(45) DEFAULT NULL COMMENT '网站名称',`logo_url` varchar(255) DEFAULT NULL,`favicon_url` varchar(255) DEFAULT NULL,`footer_line1` varchar(1024) DEFAULT NULL,`footer_line2` varchar(1024) DEFAULT NULL,`version` varchar(45) DEFAULT NULL,`sort` bigint(20) DEFAULT NULL,`modify_time` timestamp NULL DEFAULT NULL,`create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',PRIMARY KEY (`uuid`),UNIQUE KEY `id_UNIQUE` (`uuid`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='网站偏好设置表';"
		db = db.Exec(createPreference)
		if db.Error != nil {
			LOGGER.Panic(db.Error.Error())
		}
		LOGGER.Info("创建Preference表")
	}

	session := &Session{}
	hasTable = db.HasTable(session)
	if !hasTable {

		createSession := "CREATE TABLE `tank20_session` (`uuid` char(36) NOT NULL,`authentication` char(36) DEFAULT NULL COMMENT '认证身份，存放在cookie中',`user_uuid` char(36) DEFAULT NULL COMMENT '用户uuid',`ip` varchar(45) DEFAULT NULL COMMENT '用户的ip地址',`expire_time` timestamp NULL DEFAULT NULL,`sort` bigint(20) DEFAULT NULL,`modify_time` timestamp NULL DEFAULT NULL,`create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',PRIMARY KEY (`uuid`),UNIQUE KEY `id_UNIQUE` (`uuid`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='session表';"
		db = db.Exec(createSession)
		if db.Error != nil {
			LOGGER.Panic(db.Error.Error())
		}
		LOGGER.Info("创建Session表")
	}

	uploadToken := &UploadToken{}
	hasTable = db.HasTable(uploadToken)
	if !hasTable {

		createUploadToken := "CREATE TABLE `tank20_upload_token` (`uuid` char(36) NOT NULL,`user_uuid` char(36) DEFAULT NULL COMMENT '用户uuid',`folder_uuid` char(36) DEFAULT NULL,`matter_uuid` char(36) DEFAULT NULL,`filename` varchar(255) DEFAULT NULL COMMENT '文件后缀名的过滤，可以只允许用户上传特定格式的文件。',`privacy` tinyint(1) DEFAULT '1',`size` bigint(20) DEFAULT '0',`expire_time` timestamp NULL DEFAULT NULL,`ip` varchar(45) DEFAULT NULL COMMENT '消费者的ip',`sort` bigint(20) DEFAULT NULL,`modify_time` timestamp NULL DEFAULT NULL,`create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',PRIMARY KEY (`uuid`),UNIQUE KEY `id_UNIQUE` (`uuid`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='上传的token表';"
		db = db.Exec(createUploadToken)
		if db.Error != nil {
			LOGGER.Panic(db.Error.Error())
		}
		LOGGER.Info("创建UploadToken表")
	}

	user := &User{}
	hasTable = db.HasTable(user)
	if !hasTable {

	}

}
