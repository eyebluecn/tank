CREATE TABLE `tank31_bridge` (
  `uuid` char(36) NOT NULL DEFAULT '',
  `sort` bigint(20) NOT NULL,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `share_uuid` char(36) DEFAULT NULL,
  `matter_uuid` char(36) DEFAULT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tank31_dashboard` (
  `uuid` char(36) NOT NULL DEFAULT '',
  `sort` bigint(20) NOT NULL,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `invoke_num` bigint(20) NOT NULL,
  `total_invoke_num` bigint(20) NOT NULL DEFAULT '0',
  `uv` bigint(20) NOT NULL DEFAULT '0',
  `total_uv` bigint(20) NOT NULL DEFAULT '0',
  `matter_num` bigint(20) NOT NULL DEFAULT '0',
  `total_matter_num` bigint(20) NOT NULL DEFAULT '0',
  `file_size` bigint(20) NOT NULL DEFAULT '0',
  `total_file_size` bigint(20) NOT NULL DEFAULT '0',
  `avg_cost` bigint(20) NOT NULL DEFAULT '0',
  `dt` varchar(45) NOT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `uuid` (`uuid`),
  KEY `idx_dt` (`dt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tank31_download_token` (
  `uuid` char(36) NOT NULL DEFAULT '',
  `sort` bigint(20) NOT NULL,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `user_uuid` char(36) NOT NULL,
  `matter_uuid` char(36) NOT NULL,
  `expire_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `ip` varchar(128) NOT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `uuid` (`uuid`),
  KEY `idx_mu` (`matter_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tank31_footprint` (
  `uuid` char(36) NOT NULL DEFAULT '',
  `sort` bigint(20) NOT NULL,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `user_uuid` char(36) DEFAULT NULL,
  `ip` varchar(128) NOT NULL,
  `host` varchar(45) NOT NULL,
  `uri` varchar(255) NOT NULL,
  `params` text,
  `cost` bigint(20) NOT NULL DEFAULT '0',
  `success` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tank31_image_cache` (
  `uuid` char(36) NOT NULL DEFAULT '',
  `sort` bigint(20) NOT NULL,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `name` varchar(255) NOT NULL,
  `user_uuid` char(36) DEFAULT NULL,
  `username` varchar(45) NOT NULL,
  `matter_uuid` char(36) DEFAULT NULL,
  `matter_name` varchar(255) NOT NULL,
  `mode` varchar(512) DEFAULT NULL,
  `md5` varchar(45) DEFAULT NULL,
  `size` bigint(20) NOT NULL DEFAULT '0',
  `path` varchar(512) DEFAULT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `uuid` (`uuid`),
  KEY `idx_mu` (`matter_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tank31_matter` (
  `uuid` char(36) NOT NULL DEFAULT '',
  `sort` bigint(20) NOT NULL,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `puuid` char(36) DEFAULT NULL,
  `user_uuid` char(36) DEFAULT NULL,
  `username` varchar(45) NOT NULL,
  `dir` tinyint(1) NOT NULL DEFAULT '0',
  `name` varchar(255) NOT NULL,
  `md5` varchar(45) DEFAULT NULL,
  `size` bigint(20) NOT NULL DEFAULT '0',
  `privacy` tinyint(1) NOT NULL DEFAULT '0',
  `path` varchar(1024) DEFAULT NULL,
  `times` bigint(20) NOT NULL DEFAULT '0',
  `prop` varchar(1024) NOT NULL DEFAULT '{}',
  `visit_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `delete_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `uuid` (`uuid`),
  KEY `idx_puuid` (`puuid`),
  KEY `idx_uu` (`user_uuid`),
  KEY `idx_del` (`deleted`),
  KEY `idx_delt` (`delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tank31_preference` (
  `uuid` char(36) NOT NULL DEFAULT '',
  `sort` bigint(20) NOT NULL,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `name` varchar(45) DEFAULT NULL,
  `logo_url` varchar(255) DEFAULT NULL,
  `favicon_url` varchar(255) DEFAULT NULL,
  `copyright` varchar(1024) DEFAULT NULL,
  `record` varchar(1024) DEFAULT NULL,
  `download_dir_max_size` bigint(20) NOT NULL DEFAULT '-1',
  `download_dir_max_num` bigint(20) NOT NULL DEFAULT '-1',
  `default_total_size_limit` bigint(20) NOT NULL DEFAULT '-1',
  `allow_register` tinyint(1) NOT NULL DEFAULT '0',
  `preview_config` text,
  `scan_config` text,
  `deleted_keep_days` bigint(20) NOT NULL DEFAULT '7',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tank31_session` (
  `uuid` char(36) NOT NULL DEFAULT '',
  `sort` bigint(20) NOT NULL,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `user_uuid` char(36) DEFAULT NULL,
  `ip` varchar(128) NOT NULL,
  `expire_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tank31_share` (
  `uuid` char(36) NOT NULL DEFAULT '',
  `sort` bigint(20) NOT NULL,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `name` varchar(255) DEFAULT NULL,
  `share_type` varchar(45) DEFAULT NULL,
  `username` varchar(45) DEFAULT NULL,
  `user_uuid` char(36) DEFAULT NULL,
  `download_times` bigint(20) NOT NULL DEFAULT '0',
  `code` varchar(45) NOT NULL,
  `expire_infinity` tinyint(1) NOT NULL DEFAULT '0',
  `expire_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tank31_upload_token` (
  `uuid` char(36) NOT NULL DEFAULT '',
  `sort` bigint(20) NOT NULL,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `user_uuid` char(36) NOT NULL,
  `folder_uuid` char(36) NOT NULL,
  `matter_uuid` char(36) NOT NULL,
  `expire_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `filename` varchar(255) NOT NULL,
  `privacy` tinyint(1) NOT NULL DEFAULT '0',
  `size` bigint(20) NOT NULL DEFAULT '0',
  `ip` varchar(128) NOT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tank31_user` (
  `uuid` char(36) NOT NULL DEFAULT '',
  `sort` bigint(20) NOT NULL,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `role` varchar(45) DEFAULT NULL,
  `username` varchar(45) NOT NULL,
  `password` varchar(255) DEFAULT NULL,
  `avatar_url` varchar(255) DEFAULT NULL,
  `last_ip` varchar(128) DEFAULT NULL,
  `last_time` timestamp NOT NULL DEFAULT '2018-01-01 00:00:00',
  `size_limit` bigint(20) NOT NULL DEFAULT '-1',
  `total_size_limit` bigint(20) NOT NULL DEFAULT '-1',
  `total_size` bigint(20) NOT NULL DEFAULT '0',
  `status` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
