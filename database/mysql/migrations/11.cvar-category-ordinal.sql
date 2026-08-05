-- +migrate Up
ALTER TABLE `cvars`
  ADD COLUMN `cvar_category` varchar(255) NOT NULL DEFAULT '' AFTER `cvar_description`,
  ADD COLUMN `cvar_ordinal` int(11) NOT NULL DEFAULT 0 AFTER `cvar_category`;

-- +migrate Down
ALTER TABLE `cvars`
  DROP COLUMN `cvar_ordinal`,
  DROP COLUMN `cvar_category`;
