CREATE TABLE IF NOT EXISTS cages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v1mc(),
    status TEXT NOT NULL,
    capacity INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
