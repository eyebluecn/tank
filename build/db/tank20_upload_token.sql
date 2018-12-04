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