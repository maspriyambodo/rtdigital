-- Epic 3: Data Keluarga dan Warga (rollback).

DROP TRIGGER IF EXISTS trg_households_validate_head ON households;
DROP FUNCTION IF EXISTS validate_household_head();

ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_resident_tenant;

DROP TABLE IF EXISTS household_members;
DROP TABLE IF EXISTS households;
DROP TABLE IF EXISTS residents;
DROP TABLE IF EXISTS house_units;