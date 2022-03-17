CREATE TABLE IF NOT EXISTS `entries` (
  `_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `_key` varbinary(255) NOT NULL,
  `_value` blob NOT NULL,
  `_expire_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`_id`),
  UNIQUE KEY `kv_data_key_IDX` (`_key`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
