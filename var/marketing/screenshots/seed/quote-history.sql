-- Fictional submission history: the API stamps every quote with "now", so the
-- seed would otherwise create all of them in the same second. Spread them over
-- the previous days (newest = highest id) with fixed times of day.
UPDATE quotes
SET created = TIMESTAMP(CURDATE() - INTERVAL (12 - id) DAY,
                        MAKETIME(9 + (id * 7) % 10, (id * 13) % 60, 0));
