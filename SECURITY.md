# 🔐 SECURITY.md — Sécurité et mTLS dans Mini-CDN-DDoS-Lab-with-Go

## 📌 Objectif

Ce document décrit les mécanismes de **sécurité réseau** mis en œuvre dans le projet `Mini-CDN-DDoS-Lab-with-Go`, en particulier l’utilisation du **mTLS (mutual TLS)** pour sécuriser les communications entre les services `control-plane`, `auth-service` et `worker-node`.

---

## 🛡️ 1. mTLS : Mutual TLS Authentication

### 🔍 Description

Le **mTLS** permet une **authentification bilatérale** :
- Le client (ex: `worker-node`) vérifie que le `control-plane` est bien authentique.
- Le serveur (`control-plane`) exige que le client possède un **certificat client valide**.

Cela évite les connexions anonymes ou usurpées.

### ✅ Avantages
- 🔐 Authentification mutuelle forte
- 🚫 Protection contre les connexions non autorisées
- 🌍 Déploiement sécurisé dans des environnements distribués (VPS, cloud)

---

## 🧱 2. Infrastructure TLS

Tous les certificats sont signés par une **Autorité de Certification (CA)** interne :

```bash
certs/
├── ca.crt             # Autorité de Certification (à distribuer)
├── control-plane.crt  # Certificat serveur
├── control-plane.key
├── worker.crt         # Certificat client
├── worker.key
```

---

## ⚙️ 3. Configuration côté `control-plane`

```go
creds := credentials.NewTLS(&tls.Config{
    Certificates: []tls.Certificate{cert},
    ClientCAs:    certPool,
    ClientAuth:   tls.RequireAndVerifyClientCert,
})
server := grpc.NewServer(grpc.Creds(creds))
```

Cela exige que tout `worker-node` s’authentifie avec un certificat signé par la CA.

---

## 📡 4. Configuration côté `worker-node`

```go
creds := credentials.NewTLS(&tls.Config{
    Certificates: []tls.Certificate{workerCert},
    RootCAs:      certPool, // contenant ca.crt
})
conn, err := grpc.Dial("control-plane:50051", grpc.WithTransportCredentials(creds))
```

Le `worker-node` vérifie que le `control-plane` utilise un certificat authentique.

---

## 🔑 5. Distribution des certificats

Chaque agent installé utilise :
- Un certificat pré-signé
- Ou un certificat généré par un script lors de l’installation (`install.sh`)

Des outils comme **Vault** ou **cfssl** peuvent être utilisés pour automatiser la gestion PKI à terme.

---

## 🔐 6. Sécurité supplémentaire

- 🔒 JWT pour authentifier les utilisateurs humains et services REST
- 📊 Logs sécurisés avec IP, User-Agent et contrôle des accès
- 📁 Configuration centralisée dans `config.yaml`
- 🎯 Séparation des rôles : `admin`, `viewer`, `superadmin`

---

## 📍 Conclusion

L’adoption de `mTLS` garantit que **chaque composant réseau est identifié de manière fiable**, offrant un **socle sécurisé** pour le développement d’une plateforme SaaS distribuée et multi-tenant.

Pour toute question de sécurité, contacter l’équipe projet.