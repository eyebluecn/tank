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