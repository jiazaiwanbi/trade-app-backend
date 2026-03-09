CREATE TABLE IF NOT EXISTS schema_migration_heartbeat (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO schema_migration_heartbeat (name)
VALUES ('init')
ON CONFLICT (name) DO NOTHING;