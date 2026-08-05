-- +migrate Up
ALTER TABLE `users` MODIFY `password` varchar(255) DEFAULT NULL;

-- +migrate Down
ALTER TABLE `users` MODIFY `password` varchar(64) DEFAULT NULL;
