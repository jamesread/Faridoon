-- +migrate Up
INSERT INTO `groups` (id, title) VALUES (1, 'Admins');
INSERT INTO `groups` (id, title) VALUES (2, 'Users');
INSERT INTO permissions (id, `key`) VALUES (1, 'SUPERUSER');
INSERT INTO privileges_g (`group`, `permission`) VALUES (1, 1);

-- +migrate Down
