-- +migrate Up
ALTER TABLE `quotes`
  ADD COLUMN `markdown_enabled` tinyint(1) NOT NULL DEFAULT 0 AFTER `syntaxHighlighting`;

-- +migrate Down
ALTER TABLE `quotes`
  DROP COLUMN `markdown_enabled`;
