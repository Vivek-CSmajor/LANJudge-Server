PRAGMA foreign_keys = OFF;
BEGIN TRANSACTION;

-- Minimal required DB bootstrap.
-- Keep exactly one active contest row so worker penalty logic has start_time.
DELETE FROM contests;

INSERT INTO contests (id, title, start_time, end_time, status) VALUES
  (1, 'CampusCode Local Test Contest', '2026-05-09 09:00:00', '2026-05-09 12:00:00', 'running');

COMMIT;
