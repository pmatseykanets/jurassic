CREATE TABLE IF NOT EXISTS dinosaurs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v1mc(),
    name TEXT NOT NULL,
    species TEXT NOT NULL,
    cage_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (cage_id) REFERENCES cages (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS dinosaurs_cage_id_idx ON dinosaurs (cage_id);
