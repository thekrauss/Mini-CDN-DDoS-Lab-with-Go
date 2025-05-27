# 📘 Plan d'action : Construire une infrastructure SaaS multi-cloud pour ESN

## 🎯 Objectif final

Créer une plateforme SaaS multi-tenant capable de superviser, déployer, mitiger les attaques et gérer les applications des clients d'une ESN, même s’ils utilisent différents fournisseurs cloud (AWS, GCP, OVHcloud, etc).
- 🎯 Développer une architecture distribuée en Go avec gRPC et REST
- 🛡️ Expérimenter la détection et mitigation de DDoS
- ⚙️ Gérer dynamiquement des nœuds distants (registration, ping, health, metrics)
- 📈 Intégrer des outils DevOps modernes (Prometheus, Grafana, Docker, Terraform)
- ☁️ Préparer une solution SaaS cloud-native multi-client

---

## 🧱 Étape 1 : MVP Technique local (Mini CDN + DDoS)

### 🔧 Objectif

Avoir une base fonctionnelle locale avec : control-plane, worker, load-balancer, détection DDoS simple, et monitoring.

### ✅ Actions

* Implémenter `proto/node.proto` avec : `RegisterNode`, `SendMetrics`, `Ping`
* Créer `control-plane` : serveur gRPC + REST gateway (grpc-gateway)
* Créer `worker-node` : client gRPC, s’enregistre et ping
* Créer `load-balancer` : proxy HTTP avec round robin vers workers actifs
* Ajouter `pkg/ddos` : flood naïf (req/s/IP), blocage temporaire
* Ajouter `pkg/metrics` : intégration Prometheus
* Créer un dashboard Grafana pour visualiser les stats
* Dockeriser tous les composants

---

## 🌐 Étape 2 : Communication sécurisée

### 🔧 Objectif

Sécuriser les échanges et rendre la plateforme déployable à distance.

### ✅ Actions

* Implémenter le chiffrement mTLS entre tous les composants
* Ajouter un middleware JWT pour sécuriser l’API REST
* Ajouter `pkg/auth`, `pkg/config`, `pkg/logger`
* Générer et distribuer des certificats via script ou Vault
* Healthcheck régulier des worker-nodes (Ping + timeouts)
* Interface CLI pour enregistrer un node manuellement

---

## 🧑‍💼 Étape 3 : Gestion multi-client (multi-tenant)

### 🔧 Objectif

Isoler logiquement les ressources entre les clients ESN

### ✅ Actions

* Ajouter PostgreSQL pour stocker : clients, users, nodes, services, alertes
* Ajout d’une table `tenants` liée à toutes les ressources
* Middleware gRPC et HTTP basé sur JWT claims (`tenant_id`)
* Implémenter le RBAC (admin / viewer / opérateur)
* Modifier `proto` pour inclure `tenant_id` dans tous les appels

---

## ☁️ Étape 4 : Déploiement cloud et abstraction fournisseur

### 🔧 Objectif

Supporter AWS, GCP, OVH ou tout cloud via une architecture agent-based

### ✅ Actions

* Créer une interface Go `CloudProvider` avec méthodes :

  * `DeployService(ctx, config)`
  * `GetMetrics(ctx, serviceID)`
  * `RestartService(ctx, serviceID)`
* Implémenter AWSProvider, GCPProvider, OVHProvider
* Créer un agent en Go que chaque client installe (containerisé ou systemd)
* Le control-plane interagit uniquement avec l’agent, pas directement avec AWS/GCP
* Support de mTLS + JWT côté agent
* Fournir des scripts de déploiement (Ansible, bash ou Helm charts)

---

## 📊 Étape 5 : Monitoring & orchestration applicative

### 🔧 Objectif

Superviser et piloter les applications des clients via agent

### ✅ Actions

* Étendre l’agent pour lancer / restart / arrêter des services locaux
* Ajouter des hooks sur logs : récupération et envoi au control-plane
* Ajout de règles de supervision (alertes, seuils, uptime)
* Intégrer Fluent-bit ou Vector comme agent de logs côté client
* Centraliser logs/métriques par `tenant_id` dans Prometheus/Grafana/ELK

---

## 🔐 Étape 6 : Renforcement sécurité + DDoS

### 🔧 Objectif

Avoir une protection évoluée des nœuds

### ✅ Actions

* Intégrer IP fingerprinting (User-Agent, geo, ASN)
* Support GeoIP (MaxMind) pour blocage géographique
* Listes noires et blanches persistantes
* Ajout de règles dynamiques envoyées depuis le control-plane
* Simulateur d’attaque mis à jour (Go TCP flood, slowloris, HTTP flood)

---

## 📦 Étape 7 : Industrialisation & ESN-ready SaaS

### 🔧 Objectif

Préparer la plateforme pour un usage réel ESN et clients multiples

### ✅ Actions

* Interface admin en Next.js (multi-tenant, dashboard, alertes)
* Interface client SaaS (accès limité à ses services/nœuds)
* Génération de rapports d’état par client (PDF ou JSON)
* Support billing : usage par node / service / quota
* CI/CD : GitHub Actions, déploiement automatique Helm/Terraform
* Backups réguliers des configs, règles, métriques

---

## 🧠 Bonus

* Ajout support K8s : déploiement des services comme Pod
* Intégration avec outils ITSM (Jira, ServiceNow)
* Ajout de tests E2E sur infrastructure simulée avec Kind

---

## 📍 Résultat attendu

Une plateforme capable de :

* Superviser des nœuds sur n'importe quel cloud
* Déployer ou relancer des apps métiers client
* Mitiger des attaques réseau simples à moyennes
* Offrir un dashboard multi-tenant sécurisé et exploitable
* Être utilisée par ton ESN comme service managé facturable



---

## 🧰 Stack technologique recommandée

| Domaine        | Outils/Technos                        |
|----------------|--------------------------------------|
| Backend        | Go, gRPC, grpc-gateway, PostgreSQL   |
| Frontend       | Next.js, Tailwind, React Query       |
| Monitoring     | Prometheus, Grafana, Loki            |
| Auth & Sécurité| OAuth2, JWT, mTLS, Casbin            |
| Orchestration  | Docker, Helm, Terraform, K8s         |
| CI/CD          | GitHub Actions, DockerHub            |
| Cloud Providers| AWS, GCP, OVH, Scaleway, Hetzner     |


                         ┌──────────────────────────┐
                         │    Frontend / Dashboard   │
                         │  (Next.js / React UI)     │
                         └────────────┬─────────────┘
                                      │
                             HTTPS + JWT Auth
                                      │
                         ┌────────────▼─────────────┐
                         │     Authentication        │
                         │    (auth-service)         │
                         │ - JWT login/register      │
                         │ - OAuth2 / RBAC           │
                         └────────────┬─────────────┘
                                      │
                             gRPC / REST APIs
                                      │
                    ┌────────────────▼────────────────┐
                    │          Control Plane           │
                    │     (Orchestrateur central)      │
                    │ - Registre des nodes             │
                    │ - Commandes vers worker-nodes    │
                    │ - APIs REST/gRPC pour dashboard  │
                    │ - Gestion tenants / métriques    │
                    └──────┬───────────┬──────────────┘
                           │           │
                    gRPC/mTLS     Prometheus Push
                           │           │
               ┌───────────▼────────────┐
               │       Worker Node      │
               │ - Trafic HTTP          │
               │ - Protection DDoS      │
               │ - Exposition metrics   │
               │ - Exécution de jobs    │
               └────────────────────────┘



# 🛰️ SentinelPlane – Plateforme d’orchestration sécurisée multi-tenant

> Supervisez, déployez et sécurisez vos infrastructures distribuées avec un agent universel.

---

## 🎯 Objectif

Construire une solution SaaS modulaire pour :
- Superviser des services clients (nœuds, métriques, état)
- Orchestrer à distance les applications déployées
- Mitiger les attaques réseau (type DDoS) en environnement cloud ou on-premise
- Supporter plusieurs clients (multi-tenant) dans une même plateforme

---

## 🧱 Architecture

```
         +-------------------+
         |     Dashboard     |  ← Admins / Clients
         +--------+----------+
                  |
        REST API / gRPC (auth mTLS + JWT)
                  |
         +--------▼---------+
         |  Control-Plane   |  ← PostgreSQL, Prometheus
         +----+--------+----+
              |        |
      gRPC + TLS   gRPC + TLS
         +--▼--+    +--▼--+
         |Node1|    |NodeN|  ← Worker nodes multi-cloud
         +-----+    +-----+
```

---

## 🧠 Fonctionnalités

| Fonction                        | Description |
|---------------------------------|-------------|
| ✅ Orchestration distribuée     | Contrôle central des nœuds |
| ✅ Supervision (Prometheus)     | Metrics, état, logs |
| ✅ API REST + gRPC              | Interface CLI / UI front |
| ✅ DDoS simplifié               | Détection req/s/IP, blocage, alerting |
| ✅ Authentification JWT & mTLS  | Sécurité multi-tenant par défaut |
| ✅ PostgreSQL durable           | Enregistrement, alertes, tenants |
| ✅ Simulateur d’attaques        | Tests de résilience intégrés |
| ✅ Cloud-agnostique             | Déploiement possible sur AWS, GCP, OVH… |

---

## ⚙️ Stack technique

- Go, gRPC, grpc-gateway
- PostgreSQL, Prometheus, Grafana
- Docker, Terraform, GitHub Actions
- Architecture modulaire : `control-plane`, `worker-node`, `simulator`, `dashboard`

---

## 🔐 Cas d’usage pour une ESN

- Supervision de VM/containers multi-client
- Orchestration d’API métiers chez les clients
- Alerte automatique en cas de surcharge, DDoS, panne
- Interface client personnalisable
- Base pour un service managé / produit SaaS

---

## 📦 Modules

```
SentinelPlane/
├── control-plane/      # Orchestrateur principal (gRPC + REST)
├── worker-node/        # Agent distant (metrics + execution)
├── simulator/          # Générateur de trafic (normal + attaque)
├── shared-proto/       # Contrats gRPC partagés
├── deploy/             # Docker Compose / K8s / Terraform
├── config/             # Fichiers YAML + gestion Viper
└── README.md           # Ce fichier
```

---

## 🚀 Lancement (dev)

```bash
cd control-plane
go run cmd/main.go --config=config/config.yaml
```

---

## ✨ Nom du projet

**SentinelPlane** – *A control plane to deploy, observe and protect cloud-native infrastructures across multi-tenant environments.*