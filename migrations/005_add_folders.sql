ALTER TABLE documents ADD COLUMN folder TEXT;
ALTER TABLE documents ADD COLUMN reference_ids TEXT DEFAULT '[]';
