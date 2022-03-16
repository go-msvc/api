-- CREATE DATABASE `example` IF NOT EXISTS;
-- USE `example`;

DROP TABLE IF EXISTS `sessions`;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `accounts`;
CREATE TABLE `accounts` (
  `id` int primary key auto_increment,
  `name` varchar(32) NOT NULL,
  `time_created` datetime NOT NULL default now(),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

INSERT INTO `accounts` SET name="AdaptIT";

CREATE TABLE `users` (
  `id` int primary key auto_increment,
  `account_id` int NOT NULL,
  `username` varchar(32) NOT NULL,
  `password_hash` varchar(64) NOT NULL,
  `email` varchar(128),
  `phone` varchar(32),
  `time_created` datetime NOT NULL default now(),
  `active` tinyint NOT NULL default 1,
  `admin` tinyint NOT NULL default 0,
  UNIQUE KEY `username` (`username`)
  -- FOREIGN KEY (account_id) REFERENCES accounts(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- admin password is asdf@123456
insert into `users` set `account_id`=0,`username`="admin",`password_hash`="a97a5ec32e1df88ea3a6528d8b9007de5cfa2956",`email`="admin@users.com",`phone`="0821234567",`active`=1,`admin`=1;

-- in sessions we repeat the account.id and user.username because this table is hit in each
-- api operation that must authenticate, so we do not want to do expensive joins in those queries
-- but still get the info loaded for the session claim
CREATE TABLE `sessions` (
  `token` varchar(64) NOT NULL,
  `account_id` int NOT NULL,
  `user_id` int NOT NULL,
  `username` varchar(32) NOT NULL,
  `time_created` datetime NOT NULL default now(),
  `time_expire` datetime NOT NULL default now(),
  UNIQUE KEY `session_token` (`token`),
  UNIQUE KEY `session_user` (`user_id`),
  FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

