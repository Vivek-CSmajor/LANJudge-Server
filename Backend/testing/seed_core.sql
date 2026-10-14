PRAGMA foreign_keys = OFF;
BEGIN TRANSACTION;

-- Minimal required DB bootstrap.
-- Keep exactly one active contest row so worker penalty logic has start_time.
DELETE FROM contests;

INSERT INTO contests (id, title, start_time, end_time, status) VALUES
  (1, 'LANJudge Local Test Contest', '2026-05-09 09:00:00', '2026-05-09 12:00:00', 'running');

DELETE FROM problems;

INSERT INTO problems (id, title, time_limit, memory_limit, solved_count) VALUES
  (1, 'Simple Addition',  1000000000, 256, 0),
  (2, 'Multiplication',   1000000000, 256, 0),
  (3, 'Array Sum',        1000000000, 256, 0),
  (4, 'Maximum Subarray', 1000000000, 256, 0),
  (5, 'Pair Difference',  1000000000, 256, 0),
  (6, 'Grid Paths',       1000000000, 256, 0),
  (7, 'Bracket Balance',  1000000000, 256, 0),
  (8, 'Prime Queries',    1000000000, 256, 0);

COMMIT;
