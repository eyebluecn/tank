CREATE TABLE `tank20_dashboard`
(
  `uuid`             char(36)    NOT NULL,
  `invoke_num`       bigint(20) NOT NULL DEFAULT '0' COMMENT '当日访问量',
  `total_invoke_num` bigint(20) NOT NULL DEFAULT '0' COMMENT '截至目前总访问量',
  `uv`               bigint(20) NOT NULL DEFAULT '0' COMMENT '当日UV',
  `total_uv`         bigint(20) NOT NULL DEFAULT '0' COMMENT '截至目前总UV',
  `matter_num`       bigint(20) NOT NULL DEFAULT '0' COMMENT '文件数量',
  `total_matter_num` bigint(20) NOT NULL DEFAULT '0' COMMENT '截至目前文件数量',
  `file_size`        bigint(20) NOT NULL DEFAULT '0' COMMENT '当日文件大小',
  `total_file_size`  bigint(20) NOT NULL DEFAULT '0' COMMENT '截至目前文件总大小',
  `avg_cost`         bigint(20) NOT NULL DEFAULT '0' COMMENT '请求平均耗时 ms',
  `dt`               varchar(45) NOT NULL COMMENT '日期',
  `sort`             bigint(20) NOT NULL DEFAULT '0',
  `update_time`      timestamp NULL DEFAULT NULL,
  `create_time`      timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `uq_dt` (`dt`),
  KEY                `idx_dt` (`dt`),
  KEY                `idx_ct` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='汇总表，离线统计';

CREATE TABLE `tank20_download_token`
(
  `uuid`        char(36) NOT NULL,
  `user_uuid`   char(36)    DEFAULT NULL COMMENT '用户uuid',
  `matter_uuid` char(36)    DEFAULT NULL COMMENT '文件标识',
  `expire_time` timestamp NULL DEFAULT NULL COMMENT '授权访问的次数',
  `ip`          varchar(45) DEFAULT NULL COMMENT '消费者的ip',
  `sort`        bigint(20) DEFAULT NULL,
  `update_time` timestamp NULL DEFAULT NULL,
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `id_UNIQUE` (`uuid`),
  KEY           `id_mu` (`matter_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='下载的token表';

CREATE TABLE `tank20_footprint`
(
  `uuid`        char(36)  NOT NULL,
  `user_uuid`   char(36)           DEFAULT NULL,
  `ip`          varchar(45)        DEFAULT NULL,
  `host`        varchar(45)        DEFAULT NULL,
  `uri`         varchar(255)       DEFAULT NULL,
  `params`      text,
  `cost`        int(11) DEFAULT '0' COMMENT '耗时 ms',
  `success`     tinyint(1) DEFAULT '1',
  `sort`        bigint(20) NOT NULL DEFAULT '0',
  `update_time` timestamp NULL DEFAULT NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`uuid`),
  KEY           `idx_ct` (`create_time`),
  KEY           `dix_ip` (`ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='访问记录表';

CREATE TABLE `tank20_image_cache`
(
  `uuid`        char(36) NOT NULL,
  `user_uuid`   char(36)     DEFAULT NULL COMMENT '上传的用户id',
  `matter_uuid` char(36)     DEFAULT NULL,
  `mode`        varchar(512) DEFAULT NULL COMMENT '请求的uri',
  `md5`         varchar(45)  DEFAULT NULL COMMENT '文件的md5值',
  `size`        bigint(20) DEFAULT '0' COMMENT '文件大小',
  `path`        varchar(255) DEFAULT NULL,
  `sort`        bigint(20) DEFAULT NULL,
  `update_time` timestamp NULL DEFAULT NULL,
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `id_UNIQUE` (`uuid`),
  KEY           `idx_mu` (`matter_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='图片缓存表';

CREATE TABLE `tank20_matter`
(
  `uuid`        char(36) NOT NULL,
  `puuid`       varchar(45)  DEFAULT NULL COMMENT '上一级的uuid',
  `user_uuid`   char(36)     DEFAULT NULL COMMENT '上传的用户id',
  `dir`         tinyint(1) DEFAULT '0' COMMENT '是否是文件夹',
  `alien`       tinyint(1) DEFAULT '0',
  `name`        varchar(255) DEFAULT NULL COMMENT '文件名称',
  `md5`         varchar(45)  DEFAULT NULL COMMENT '文件的md5值',
  `size`        bigint(20) DEFAULT '0' COMMENT '文件大小',
  `privacy`     tinyint(1) DEFAULT '0' COMMENT '文件是否是公有的',
  `path`        varchar(255) DEFAULT NULL,
  `times`       bigint(20) DEFAULT '0' COMMENT '下载次数',
  `sort`        bigint(20) DEFAULT NULL,
  `update_time` timestamp NULL DEFAULT NULL,
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `id_UNIQUE` (`uuid`),
  KEY           `idx_uu` (`user_uuid`),
  KEY           `idx_ct` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='file表';

CREATE TABLE `tank20_preference`
(
  `uuid`         char(36) NOT NULL,
  `name`         varchar(45)   DEFAULT NULL COMMENT '网站名称',
  `logo_url`     varchar(255)  DEFAULT NULL,
  `favicon_url`  varchar(255)  DEFAULT NULL,
  `footer_line1` varchar(1024) DEFAULT NULL,
  `footer_line2` varchar(1024) DEFAULT NULL,
  `version`      varchar(45)   DEFAULT NULL,
  `sort`         bigint(20) DEFAULT NULL,
  `update_time`  timestamp NULL DEFAULT NULL,
  `create_time`  timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `id_UNIQUE` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='网站偏好设置表';

CREATE TABLE `tank20_session`
(
  `uuid`        char(36) NOT NULL,
  `user_uuid`   char(36)    DEFAULT NULL COMMENT '用户uuid',
  `ip`          varchar(45) DEFAULT NULL COMMENT '用户的ip地址',
  `expire_time` timestamp NULL DEFAULT NULL,
  `sort`        bigint(20) DEFAULT NULL,
  `update_time` timestamp NULL DEFAULT NULL,
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `id_UNIQUE` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='session表';

CREATE TABLE `tank20_upload_token`
(
  `uuid`        char(36) NOT NULL,
  `user_uuid`   char(36)     DEFAULT NULL COMMENT '用户uuid',
  `folder_uuid` char(36)     DEFAULT NULL,
  `matter_uuid` char(36)     DEFAULT NULL,
  `filename`    varchar(255) DEFAULT NULL COMMENT '文件后缀名的过滤，可以只允许用户上传特定格式的文件。',
  `privacy`     tinyint(1) DEFAULT '1',
  `size`        bigint(20) DEFAULT '0',
  `expire_time` timestamp NULL DEFAULT NULL,
  `ip`          varchar(45)  DEFAULT NULL COMMENT '消费者的ip',
  `sort`        bigint(20) DEFAULT NULL,
  `update_time` timestamp NULL DEFAULT NULL,
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `id_UNIQUE` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='上传的token表';

CREATE TABLE `tank20_user`
(
  `uuid`        char(36) NOT NULL,
  `role`        varchar(45)  DEFAULT 'USER',
  `username`    varchar(255) DEFAULT NULL COMMENT '昵称',
  `password`    varchar(255) DEFAULT NULL COMMENT '密码',
  `email`       varchar(45)  DEFAULT NULL COMMENT '邮箱',
  `phone`       varchar(45)  DEFAULT NULL COMMENT '电话',
  `gender`      varchar(45)  DEFAULT 'UNKNOWN' COMMENT '性别，默认未知',
  `city`        varchar(45)  DEFAULT NULL COMMENT '城市',
  `avatar_url`  varchar(255) DEFAULT NULL COMMENT '头像链接',
  `last_time`   datetime     DEFAULT NULL COMMENT '上次登录使劲按',
  `last_ip`     varchar(45)  DEFAULT NULL,
  `size_limit`  int(11) DEFAULT '-1' COMMENT '该账号上传文件的大小限制，单位byte。<0 表示不设限制',
  `status`      varchar(45)  DEFAULT 'OK',
  `sort`        bigint(20) DEFAULT NULL,
  `update_time` timestamp NULL DEFAULT NULL,
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `id_UNIQUE` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户表描述';
