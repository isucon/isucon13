DELETE FROM themes;
DELETE FROM users;
DELETE FROM livestreams;
DELETE FROM tags;
DELETE FROM livestream_tags;
DELETE FROM livestream_viewers_history;
DELETE FROM livecomments;
DELETE FROM livecomment_reports;
DELETE FROM ng_words;
DELETE FROM reactions;

SET GLOBAL innodb_monitor_enable = '%';

# performance-schema-consumer-events-stages-current= 1
# performance-schema-consumer-events-stages-history= 1
# performance-schema-consumer-events-stages-history-long= 1