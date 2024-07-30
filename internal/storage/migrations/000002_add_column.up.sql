BEGIN;
    ALTER TABLE urls ADD COLUMN is_deleted boolean;
COMMIT;