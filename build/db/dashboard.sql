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