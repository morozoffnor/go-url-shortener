BEGIN;
    ALTER TABLE urls ADD COLUMN is_deleted boolean DEFAULT false;
COMMIT;