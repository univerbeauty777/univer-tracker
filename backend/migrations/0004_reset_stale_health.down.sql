-- Down migration is a no-op: the data we'd need to "restore" is the
-- buggy state itself. Safe to leave the recomputed values in place.
SELECT 1;
