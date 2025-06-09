-- Enable uuid extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


CREATE TABLE IF NOT EXISTS utilisateurs (
    id_utilisateur UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    login_id VARCHAR(255) UNIQUE,
    nom VARCHAR(100) NOT NULL,
    prenom VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    genre VARCHAR(10) CHECK (genre IN ('Homme', 'Femme')),
    telephone VARCHAR(20) UNIQUE NOT NULL,
    mot_de_passe_hash TEXT NOT NULL,
    role  VARCHAR(50) NOT NULL,
    permissions TEXT[],
    tenant_id UUID NOT NULL,
    status TEXT,
    is_active BOOLEAN,
    date_inscription TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    derniere_connexion TIMESTAMP,
    last_activity TIMESTAMP DEFAULT now(),
    token_exp TIMESTAMP,
    mfa_enabled BOOLEAN DEFAULT false,
    photo_profil TEXT
);

CREATE TABLE IF NOT EXISTS tenant (
    id_tenant UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nom VARCHAR(150) NOT NULL,
    adresse TEXT,
    ville VARCHAR(100),
    code_postal VARCHAR(20),
    contact_telephone VARCHAR(20),
    contact_email VARCHAR(150),
    directeur_nom VARCHAR(100),
    directeur_contact VARCHAR(100),
    type_etablissement VARCHAR(100),
    parametres_specifiques TEXT,
    date_creation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    validation_status VARCHAR(50) DEFAULT 'En attente',
    logo_url TEXT
);

CREATE TABLE IF NOT EXISTS utilisateurs_permissions (
    id_utilisateur UUID REFERENCES utilisateurs(id_utilisateur) ON DELETE CASCADE,
    permission TEXT NOT NULL,
    PRIMARY KEY (id_utilisateur, permission)
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id_audit UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID REFERENCES utilisateurs(id_utilisateur) ON DELETE SET NULL,
    role VARCHAR(50),
    action VARCHAR(255),
    target_id UUID,
    target_type VARCHAR(100),
    details TEXT,
    ip_address VARCHAR(50),
    user_agent TEXT,
    action_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50),
    session_id UUID,
    id_tenant UUID REFERENCES tenant(id_tenant) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    user_id UUID PRIMARY KEY REFERENCES utilisateurs(id_utilisateur) ON DELETE CASCADE,
    token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ip_address VARCHAR(50),
    user_agent TEXT,
);

//audit_logs
Ajoute target_json JSONB au lieu de details TEXT si tu veux des logs structurés
Indexe sur action_time, admin_id, tenant_id pour recherches rapides

//refresh_tokens
Ajoute une colonne ip, user_agent, device_id
Prends en compte l’expiration à l’invalidation du token global (pas juste l’accès)

//utilisateurs
Ajouter un champ origin ou auth_provider si tu veux un jour supporter SSO (Keycloak, LDAP…)
Ajouter un champ language ou locale si tu veux gérer plusieurs langues (ex. en/FR pour dashboard)