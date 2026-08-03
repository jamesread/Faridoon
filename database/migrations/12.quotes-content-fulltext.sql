-- +migrate Up
ALTER TABLE `quotes` ADD FULLTEXT KEY `quotes_content_ft` (`content`);

-- +migrate Down
ALTER TABLE `quotes` DROP INDEX `quotes_content_ft`;
