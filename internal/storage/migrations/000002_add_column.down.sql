BEGIN;
    ALTER TABLE urls DROP COLUMN is_deleted;
COMMIT;