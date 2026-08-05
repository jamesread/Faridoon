-- +migrate Up
CREATE TABLE `cvars` (
  `cvar_key` varchar(255) NOT NULL,
  `cvar_value_int` tinyint(4) DEFAULT NULL,
  `cvar_value_string` varchar(255) DEFAULT NULL,
  `cvar_main_type` varchar(255) NOT NULL,
  PRIMARY KEY (`cvar_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE IF EXISTS `cvars`;
