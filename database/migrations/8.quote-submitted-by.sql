-- +migrate Up
ALTER TABLE `quotes`
  ADD COLUMN `submitted_by_user_id` int(11) DEFAULT NULL AFTER `syntaxHighlighting`,
  ADD COLUMN `submitted_by_username` varchar(32) DEFAULT NULL AFTER `submitted_by_user_id`;

-- +migrate Down
ALTER TABLE `quotes`
  DROP COLUMN `submitted_by_username`,
  DROP COLUMN `submitted_by_user_id`;
