
```
                        [ Interface Web (Next.js) ]
                                  |
                       REST (secure) via API Gateway
                                  |
                                  v
    ┌─────────────────────────────────────────────────────────────┐
    │                      Plateforme API                         │
    └─────────────────────────────────────────────────────────────┘
       │                        │                          │
       ▼                        ▼                          ▼
┌────────────────┐     ┌────────────────────┐     ┌───────────────────────┐
│ Control Plane   │     │ Kubernetes Manager │     │     Auth Service      │
│────────────────│     │────────────────────│     │───────────────────────│
│ - AuthN/AuthZ   │     │ - Gère clusters K8s│     │ - JWT, OAuth2         │
│ - Start Workflow│     │ - Provision, Scale │     │ - Validation tokens   │
│ - Manage Nodes  │     │ - Deploy/Upgrade    │     │                       │
└────────────────┘     └────────────────────┘     └───────────────────────┘
       │                        ▲                          ▲
       │                        │                          │
       ▼                        │                          │
  gRPC → Temporal Server        │                          │
       │                        │                          │
       ▼                        │                          │
┌───────────────────────────────┐                         │
│       Workflow Engine         │                         │
│───────────────────────────────│                         │
│ - Orchestrates operations     │                         │
│ - LogAuditActivity            │                         │
│ - ExecuteCommandActivity      │                         │
│ - NotifyFailureActivity       │                         │
│ - UpdateStatusActivity        │                         │
└───────────────────────────────┘                         │
       │
       ▼
┌───────────────────────────────┐
│         Worker Node           │
│───────────────────────────────│
│ - Poll Temporal               │
│ - Exécute commandes (bash...) │
│ - Renvoie les résultats       │
└───────────────────────────────┘

```