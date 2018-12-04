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