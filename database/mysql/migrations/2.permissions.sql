-- +migrate Up

UPDATE permissions SET `description` = "Has all permissions" WHERE `key` = "SUPERUSER";
INSERT INTO permissions (`key`, `description`) VALUES ('BYPASS_APPROVAL', 'New quotes will be auto-approved');
INSERT INTO permissions (`key`, `description`) VALUES ('APPROVE_QUOTES', 'Approve quotes in the approval queue');

-- +migrate Down
