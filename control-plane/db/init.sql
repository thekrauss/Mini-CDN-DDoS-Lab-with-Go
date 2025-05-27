
-- Table des tenants (clients ESN)
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Table des nœuds
CREATE TABLE IF NOT EXISTS nodes (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    ip INET NOT NULL,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    status TEXT DEFAULT 'offline',
    tags JSONB DEFAULT '{}'::jsonb,
    last_seen TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index pour les recherches fréquentes
CREATE INDEX IF NOT EXISTS idx_nodes_tenant_id ON nodes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_nodes_ip ON nodes(ip);
CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status);
CREATE INDEX IF NOT EXISTS idx_nodes_last_seen ON nodes(last_seen);

