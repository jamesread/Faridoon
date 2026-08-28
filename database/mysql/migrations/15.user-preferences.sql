-- +migrate Up
CREATE TABLE user_preferences (
  id INT NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) DEFAULT NULL,
  updated_at DATETIME(3) DEFAULT NULL,
  user_id INT NOT NULL,
  language VARCHAR(32) NOT NULL DEFAULT '',
  sidebar_enabled TINYINT(1) NOT NULL DEFAULT 1,
  theme_toggle_enabled TINYINT(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY idx_user_preferences_user_id (user_id),
  CONSTRAINT fk_user_preferences_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE IF EXISTS user_preferences;
