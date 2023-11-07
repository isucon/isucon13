CREATE DATABASE IF NOT EXISTS `isupipe`;

DROP USER IF EXISTS `isucon`@`%`;
CREATE USER isucon IDENTIFIED BY 'isucon';
GRANT ALL PRIVILEGES ON isupipe.* TO 'isucon'@'%';

CREATE DATABASE IF NOT EXISTS `isudns`;

DROP USER IF EXISTS `isudns`@`%`;
CREATE USER isudns IDENTIFIED BY 'isudns';
GRANT ALL PRIVILEGES ON isudns.* TO 'isudns'@'%';

-- NOTE: initializeの名前解決に必要
INSERT INTO domains (id, name, type) VALUES (1, 'u.isucon.dev', 'NATIVE');
INSERT INTO records (domain_id, name, type, content, ttl, prio) VALUES (1, 'u.isucon.dev', 'SOA', 'localhost hostmaster.u.isucon.dev 0 10800 3600 604800 3600', 3600, NULL);
INSERT INTO records (domain_id, name, type, content, ttl, prio) VALUES (1, 'ns1.u.isucon.dev', 'NS', 'ns1.u.isucon.dev', 3600, NULL);
INSERT INTO records (domain_id, name, type, content, ttl, prio) VALUES (1, 'u.isucon.dev', 'A', '127.0.0.1', 3600, NULL);
INSERT INTO records (domain_id, name, type, content, ttl, prio) VALUES (1, 'pipe.u.isucon.dev', 'A', '127.0.0.1', 3600, NULL);
INSERT INTO records (domain_id, name, type, content, ttl, prio) VALUES (1, 'test001.u.isucon.dev', 'A', '127.0.0.1', 3600, NULL);
