-- +migrate Up
CREATE TABLE `webhooks` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `url` varchar(2048) NOT NULL,
  `event` varchar(64) NOT NULL,
  `secret` varchar(255) NOT NULL,
  `enabled` tinyint(4) NOT NULL DEFAULT 1,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `webhooks_event_enabled_idx` (`event`,`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE IF EXISTS `webhooks`;
