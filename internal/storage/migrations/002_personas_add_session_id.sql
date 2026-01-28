-- PersonaForge Database Schema Migration
-- Version: 002
-- Description: Allow guest-scoped custom personas via session_id

ALTER TABLE personas
ADD COLUMN IF NOT EXISTS session_id VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_personas_session_id ON personas(session_id);

-- Defaults have user_id NULL and session_id NULL, so allow that.
-- Custom personas must be owned by either a user_id or a session_id.
ALTER TABLE personas
DROP CONSTRAINT IF EXISTS personas_owner_check;

ALTER TABLE personas
ADD CONSTRAINT personas_owner_check
CHECK (
    is_default = true OR user_id IS NOT NULL OR session_id IS NOT NULL
);


