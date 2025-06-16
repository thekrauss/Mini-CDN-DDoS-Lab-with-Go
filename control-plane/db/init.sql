CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Table principale : nodes enregistrés
CREATE TABLE IF NOT EXISTS nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hostname TEXT NOT NULL,
    ip_address TEXT,
    location TEXT,
    os TEXT,
    version TEXT,
    status TEXT DEFAULT 'offline',
    tenant_id UUID NOT NULL,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tags TEXT[],
    is_blacklisted BOOLEAN DEFAULT FALSE,
    UNIQUE (hostname, tenant_id)
);


-- Métriques collectées par agent
CREATE TABLE IF NOT EXISTS node_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id UUID REFERENCES nodes(id) ON DELETE CASCADE,
    tenant_id UUID,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    cpu_usage FLOAT,
    mem_usage FLOAT,
    bandwidth_rx BIGINT,
    bandwidth_tx BIGINT,
    connections INT,
    disk_io BIGINT,
    uptime BIGINT,
    status TEXT
);


-- Suivi des pings/états (heartbeat)
CREATE TABLE IF NOT EXISTS node_status_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id UUID REFERENCES nodes(id) ON DELETE CASCADE,
    status TEXT,
    message TEXT,
    logged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Commandes envoyées aux nœuds
CREATE TABLE IF NOT EXISTS commands (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id UUID REFERENCES nodes(id) ON DELETE CASCADE,
    command TEXT NOT NULL,
    arguments TEXT,
    status TEXT DEFAULT 'sent', -- sent | ack | failed | executed
    issued_by UUID, -- utilisateur qui l’a émis
    tenant_id UUID NOT NULL,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    executed_at TIMESTAMP
);

-- Configuration personnalisée d’un tenant
CREATE TABLE IF NOT EXISTS tenants_config (
    tenant_id UUID PRIMARY KEY,
    max_nodes INT DEFAULT 10,
    max_bandwidth BIGINT DEFAULT 1000000000,
    max_services INT DEFAULT 5,
    allow_geo_restriction BOOLEAN DEFAULT FALSE,
    settings JSONB
);

-- Services déployés par le control-plane
CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id UUID REFERENCES nodes(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    service_name TEXT,
    status TEXT DEFAULT 'running',
    version TEXT,
    port INT,
    deployed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Alertes déclenchées sur les nœuds
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id UUID REFERENCES nodes(id) ON DELETE CASCADE,
    severity TEXT CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    message TEXT,
    category TEXT,
    resolved BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP
);

-- Audit des actions techniques (infra)
CREATE TABLE IF NOT EXISTS infra_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID,
    role TEXT,
    action TEXT,
    target TEXT,
    details TEXT,
    ip_address TEXT,
    user_agent TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tenant_id UUID
);
