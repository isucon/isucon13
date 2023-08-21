CREATE DATABASE IF NOT EXISTS `isupipe`;

DROP USER IF EXISTS `isucon`@`%`;
CREATE USER isucon IDENTIFIED BY 'isucon';
GRANT ALL PRIVILEGES ON isupipe.* TO 'isucon'@'%';
