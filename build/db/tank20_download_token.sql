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