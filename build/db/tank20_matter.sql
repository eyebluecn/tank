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