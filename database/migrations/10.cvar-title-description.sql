-- +migrate Up
ALTER TABLE `cvars`
  ADD COLUMN `cvar_title` varchar(255) NOT NULL DEFAULT '' AFTER `cvar_main_type`,
  ADD COLUMN `cvar_description` varchar(512) NOT NULL DEFAULT '' AFTER `cvar_title`;

-- +migrate Down
ALTER TABLE `cvars`
  DROP COLUMN `cvar_description`,
  DROP COLUMN `cvar_title`;
