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