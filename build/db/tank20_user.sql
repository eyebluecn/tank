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