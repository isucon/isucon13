DELETE FROM tags;
DELETE FROM livestreams;
DELETE FROM superchats;
DELETE FROM reactions;
DELETE FROM livestream_viewers;
DELETE FROM users;

INSERT INTO users (name, display_name, password, description) VALUES ('isupipe', 'isupipe', '1sup1pe', 'isupipe owner');
INSERT INTO tags (name) VALUES ('chair');
INSERT INTO tags (name) VALUES ('fruits');
INSERT INTO tags (name) VALUES ('cat');
INSERT INTO tags (name) VALUES ('dog');
