-- Preserve completed MFA verification across refresh-token rotation.
ALTER TABLE sessions
    ADD COLUMN mfa_verified BOOLEAN NOT NULL DEFAULT FALSE;