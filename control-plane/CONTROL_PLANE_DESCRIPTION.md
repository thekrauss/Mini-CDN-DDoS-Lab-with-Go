# 🧠 Control Plane – Orchestrateur central

## 🎯 Objectif

Le `control-plane` est le **cœur logique de la plateforme**. Il centralise les communications, les décisions et les actions entre les agents `worker-node`, les utilisateurs, le dashboard et les autres services (auth, monitoring).

---

## 🧭 Rôle du `control-plane`

| Fonction                          | Description |
|----------------------------------|-------------|
| 🔗 Enregistrement des worker-nodes | Lorsqu’un agent démarre, il s’identifie et s’enregistre |
| 📡 Heartbeat / Ping               | Vérifie périodiquement que les nœuds sont en ligne |
| 📊 Collecte de métriques         | Agrège les données (CPU, trafic, uptime) et les expose |
| 📤 Commandes distantes           | Envoie des instructions aux agents (restart, push, stop…) |
| 🔐 Contrôle d’accès              | Applique RBAC et filtre les ressources par `tenant_id` |
| 🌐 Interface API REST/gRPC       | Permet aux dashboards et aux outils de gestion d’interagir |

---

---

## 🔐 Sécurité intégrée

| Mécanisme        | Description |
|------------------|-------------|
| JWT              | Chaque requête est associée à un utilisateur authentifié |
| RBAC             | Rôle admin / opérateur / viewer par tenant |
| Multi-tenant     | Les données sont filtrées automatiquement par `tenant_id` |
| mTLS             | Communication sécurisée avec les agents (mutual TLS) |

---

## 📡 API exposées

| Type     | Endpoint                         | Usage |
|----------|----------------------------------|-------|
| gRPC     | `RegisterNode()`                | Appelé par l’agent au démarrage |
| gRPC     | `Ping()`                        | Heartbeat pour maintenir le node actif |
| gRPC     | `SendMetrics()`                 | Exposition des métriques |
| gRPC     | `RestartService()`              | Redémarre un service à distance |
| REST     | `/nodes`                        | Affiche tous les nœuds visibles par l’utilisateur |
| REST     | `/metrics`                      | Route Prometheus exporter |

---

## 📦 Base de données utilisée

PostgreSQL avec les tables suivantes :
- `nodes` : ID, nom, tenant_id, status, last_seen, tags
- `tenants` : ID, nom, description
- `metrics` : node_id, timestamp, cpu, mem, net
- `users` : ID, email, role, tenant_id
- `audit_logs` : action, user_id, timestamp

---

## 🧰 Stack utilisée

| Composant     | Technologie         |
|---------------|---------------------|
| Serveur       | Go, gRPC, grpc-gateway |
| Auth          | JWT, Casbin (RBAC)  |
| DB            | PostgreSQL          |
| Monitoring    | Prometheus          |
| Logs          | Zap / Zerolog       |
| Config        | Viper + YAML        |

---

## ✅ Avantages

| Atout                       | Impact |
|-----------------------------|--------|
| Séparation claire services  | Meilleure scalabilité |
| Multi-tenant natif          | Adapté ESN/SaaS       |
| REST + gRPC exposés         | Compatibilité dashboard / CLI |
| Sécurisé dès le début       | JWT, RBAC, mTLS       |

---

## 🧠 Résumé

Le `control-plane` est un composant critique qui agit comme **chef d’orchestre** du système. Il gère l’enregistrement, la supervision, l’accès sécurisé et l’interaction avec tous les nœuds de la plateforme.

Il constitue la **colonne vertébrale** de la plateforme SaaS, garantissant cohérence, sécurité et extensibilité pour tous les clients d’une ESN.
