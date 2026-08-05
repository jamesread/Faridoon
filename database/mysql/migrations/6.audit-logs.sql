-- +migrate Up
CREATE TABLE `logs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `created` datetime NOT NULL,
  `actor_user_id` int(11) DEFAULT NULL,
  `actor_username` varchar(32) DEFAULT NULL,
  `action` varchar(64) NOT NULL,
  `entity_type` varchar(32) DEFAULT NULL,
  `entity_id` int(11) DEFAULT NULL,
  `detail` text DEFAULT NULL,
  `ip` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `logs_created_idx` (`created`),
  KEY `logs_action_idx` (`action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE IF EXISTS `logs`;
