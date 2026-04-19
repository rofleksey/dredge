-- Human-readable rule label; existing rows get a placeholder until edited in the UI.
ALTER TABLE rules
    ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '<blank>';
