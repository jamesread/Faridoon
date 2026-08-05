-- +migrate Up
CREATE TABLE webhook_targets (
  id INT NOT NULL AUTO_INCREMENT,
  url VARCHAR(2048) NOT NULL,
  secret VARCHAR(255) NOT NULL,
  enabled TINYINT NOT NULL DEFAULT 1,
  created DATETIME DEFAULT NULL,
  updated DATETIME DEFAULT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE webhook_events (
  id INT NOT NULL AUTO_INCREMENT,
  webhook_target_id INT NOT NULL,
  event VARCHAR(64) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY webhook_events_target_event_uidx (webhook_target_id, event),
  KEY webhook_events_event_idx (event),
  CONSTRAINT webhook_events_target_fk
    FOREIGN KEY (webhook_target_id) REFERENCES webhook_targets (id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO webhook_targets (id, url, secret, enabled, created, updated)
SELECT id, url, secret, enabled, created, updated FROM webhooks;

INSERT INTO webhook_events (webhook_target_id, event)
SELECT id, event FROM webhooks;

DROP TABLE IF EXISTS webhooks;

-- +migrate Down
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

INSERT INTO webhooks (id, url, event, secret, enabled, created, updated)
SELECT t.id, t.url, COALESCE(
  (SELECT e.event FROM webhook_events e WHERE e.webhook_target_id = t.id ORDER BY e.event LIMIT 1),
  'approval.requested'
), t.secret, t.enabled, t.created, t.updated
FROM webhook_targets t;

DROP TABLE IF EXISTS webhook_events;
DROP TABLE IF EXISTS webhook_targets;
