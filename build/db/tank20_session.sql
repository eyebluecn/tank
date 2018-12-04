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