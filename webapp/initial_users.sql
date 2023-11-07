--- NOTE: パスワードは `test`
INSERT INTO users (id, name, display_name, description, password) VALUES (1, 'test001', '検証用ユーザ', '社内検証用', '$2a$10$6Yw4aufjz8wAgFCLXHJ/BugeJpZ6qY5ofvrfpWwJOLxgTg5eaq3l.');
INSERT INTO themes (user_id, dark_mode) VALUES (1, false);