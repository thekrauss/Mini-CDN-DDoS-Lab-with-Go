
```
+--------------------------------------------------------------------------------------+
|                                 [ Interface Web (Next.js) ]                         |
|                                                                                      |
|      - Dashboard multi-tenant (admin/client)                                         |
|      - Envoie les actions via REST à API Gateway sécurisée                          |
+---------------------------------------------+----------------------------------------+
                                              |
                                REST (via API Gateway sécurisée, JWT)
                                              |
                           +------------------v------------------+
                           |           Control Plane             |
                           |-------------------------------------|
                           | - Authentifie utilisateur via JWT   |
                           | - Valide les permissions (RBAC)     |
                           | - Gère base PostgreSQL & Redis      |
                           | - Lance workflows via Temporal      |
                           +------------------+------------------+
                                              |
                                      gRPC SDK (secure)
                                              |
                           +------------------v------------------+
                           |          Workflow Engine (Temporal) |
                           |-------------------------------------|
                           | - Orchestration fiable (retry, etc) |
                           | - LogAuditActivity                  |
                           | - ExecuteCommandActivity            |
                           | - UpdateStatusActivity              |
                           | - NotifyFailureActivity             |
                           +------------------+------------------+
                                              |
                            Activités dispatchées vers Workers (poll)
                                              |
                           +------------------v------------------+
                           |             Worker Node             |
                           |-------------------------------------|
                           | - Poll les activities Temporal       |
                           | - Exécute restart/update/scripts     |
                           | - Ping(), SendMetrics()              |
                           | - GetConfig() si mode dynamique      |
                           +-------------------------------------+
                                              |
                              (optionnel) appelle Kube Manager
                                              |
                           +------------------v------------------+
                           |          Kube Manager Service        |
                           |-------------------------------------|
                           | - Provisionne des clusters K8s       |
                           | - Déploie via Helm/Kustomize         |
                           | - Gère utilisateurs K8s              |
                           | - Upgrade, restart, scale            |
                           | - Surveille l'état des nœuds K8s     |
                           +-------------------------------------+
```