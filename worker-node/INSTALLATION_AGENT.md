# 🛰️ Installation de l'Agent `worker-node` via le Dashboard

## 🎯 Objectif

Permettre à un administrateur ESN de déployer facilement un agent `worker-node` sur un VPS distant ou un cloud, en assurant sécurité, traçabilité, et simplicité.

---

## 🧠 Contexte

Dans une architecture SaaS multi-tenant, chaque client peut déployer un ou plusieurs agents `worker-node` sur ses serveurs. Ces agents :

- S'enregistrent automatiquement auprès du `control-plane`
- Reçoivent des instructions distantes (restart, push config, blocage IP…)
- Exposent des métriques de supervision

---

## 🖥️ Étapes d'installation via le dashboard

### 1. Connexion d'un administrateur

L'administrateur se connecte à l'interface de gestion web (Next.js / React), avec un JWT signé.

---

### 2. Création d'un nouveau nœud

L'admin :
- Sélectionne un **client (tenant)** existant
- Choisit le **type de worker** (HTTP server, CDN, logger…)
- Donne un **nom ou tag** (ex: `client-X-eu1`)
- Lance la génération du **token d’installation**

---

### 3. Génération automatisée

Le backend génère automatiquement :
- ✅ Un **JWT ou token UUID signé**
- ✅ Un **fichier de configuration YAML**
- ✅ Un **lien de téléchargement du binaire**
- ✅ Une **commande `curl | bash` personnalisée**

Exemple :

```bash
curl -sSL https://cdn.example.com/install.sh | bash -s -- --token abc.def.ghi --tenant client-X
```

---

## 🧰 Ce que fait le script `install.sh`

1. Télécharge le binaire `worker-node`
2. Génère un fichier `config.yaml` :
   ```yaml
   tenant: client-X
   token: abc.def.ghi
   control_plane_url: "https://control.cdncdncdn.com:50051"
   metrics_enabled: true
   ```

✅ Paramétrage initial de l’agent	Le config.yaml contient tous les paramètres que le binaire worker-node doit lire au démarrage
✅ Connexion au control-plane	URL du control-plane, port gRPC/REST, endpoint mTLS
✅ Authentification	Jeton JWT ou clé d’authentification mTLS pour prouver l’identité du nœud
✅ Identité du tenant	Pour savoir à quel client (tenant_id) appartient ce nœud
✅ Activation de modules	Ex: monitoring Prometheus, logs, règles de mitigation
✅ Paramètres réseau	Tags, région, ports exposés (si CDN HTTP intégré)
✅ Reprise automatique après redémarrage	Si l’agent crash ou le VPS redémarre, il peut repartir en relisant config.yaml

« Le config.yaml est la mémoire locale de l’agent. Il lui permet de redémarrer, se reconnecter, et fonctionner de manière autonome. C’est aussi un point central pour le déploiement automatisé et l’exploitation à distance dans une architecture multi-tenant. »


3. Crée un service `systemd` pour lancer et maintenir l'agent actif
4. Démarre le service

---

## 🔐 Sécurité

| Mécanisme                  | Rôle |
|---------------------------|------|
| Jeton JWT signé           | Authentifie l'agent au premier enregistrement |
| mTLS                      | Authentification mutuelle agent / control-plane |
| IP d’origine surveillée   | Anti-spoofing |
| Logs d’enregistrement     | Auditabilité (PostgreSQL) |

---

## 📊 Affichage dans le dashboard

Une fois installé, le node apparaît dans l'interface :

| Nom             | Statut | Uptime | Actions         |
|------------------|--------|--------|------------------|
| client-X-eu1     | ✅ Actif | 3h     | 🔁 Restart 🗑 Delete |

---

## ✅ Avantages

| Besoin ESN               | Réponse apportée |
|--------------------------|------------------|
| Déploiement rapide       | Script `curl | bash` généré dynamiquement |
| Supervision centralisée  | Ping + métriques visibles en dashboard |
| Sécurité avancée         | JWT, mTLS, audit log |
| Extensibilité            | Compatible cloud ou on-prem |

---

## 🧠 Résumé

Cette méthode d’installation permet :
- Une intégration simple pour les clients
- Un contrôle total depuis le control-plane
- Une visibilité globale dans un tableau de bord multi-tenant sécurisé
