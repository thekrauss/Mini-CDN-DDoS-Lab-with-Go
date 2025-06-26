### 📡 Vue Système 

flowchart TD
    UI[Interface Web (Next.js)]
    APIGW[REST API Gateway sécurisée]
    CP[Control Plane]
    TEMPORAL[Temporal Server]
    WE[Workflow Engine]
    WN[Worker Node]
    KUBE[Kube Manager]
    AUTH[Auth Service]

    UI --> APIGW
    APIGW --> CP
    CP -->|Auth via JWT| AUTH
    CP -->|Lance workflow| TEMPORAL
    TEMPORAL -->|Exécute steps| WE
    WE -->|Dispatch Activity| WN
    WN -->|Résultat| TEMPORAL
    TEMPORAL -->|Log, notif, etc| WE
    CP -.->|optionnel| KUBE

    click TEMPORAL href "https://temporal.io" _blank
    click AUTH href "https://jwt.io" _blank

    classDef external fill:#f9f,stroke:#333,stroke-width:1px;
    classDef core fill:#bbf,stroke:#333,stroke-width:1px;
    classDef infra fill:#bfb,stroke:#333,stroke-width:1px;
    class TEMPORAL,AUTH,KUBE external;
    class CP,WE,WN core;
