# ER Diagram

## Current Schema (Implemented)
```mermaid
erDiagram
  users ||--o{ projects : owns
  users ||--o{ incidents : reports
  users ||--o{ comments : writes
  users ||--o{ audit_logs : performs

  projects ||--o{ incidents : contains
  incidents ||--o{ comments : has

  projects ||--o{ audit_logs : project_scope
  incidents ||--o{ audit_logs : incident_scope

  users {
    uint id PK
    string name
    string email UNIQUE
    string password_hash
    timestamp created_at
    timestamp updated_at
  }

  projects {
    uuid id PK
    string name
    string slug UNIQUE
    string description
    string environment
    string health
    uint owner_id FK
    int members
    int services
    timestamp created_at
    timestamp updated_at
    timestamp deleted_at
  }

  incidents {
    uint id PK
    string title
    string description
    string severity
    string status
    uuid project_id FK
    uint user_id FK
    timestamp created_at
    timestamp updated_at
  }

  comments {
    uint id PK
    string content
    uint incident_id FK
    uint user_id FK
    timestamp created_at
    timestamp updated_at
  }

  audit_logs {
    uint id PK
    uint user_id FK
    uuid project_id FK
    uint incident_id FK
    string entity_type
    string entity_id
    string action
    string field_name
    string old_value
    string new_value
    timestamp created_at
  }
```

## Future Schema Additions (Future Architecture)
- organizations, teams, memberships.
- resources and resource_relationships.
- monitoring signals and alert tables.
- AI artifact tables (embeddings metadata, retrieval traces).

## Related
- `docs/database/DATABASE.md`
- `docs/database/DATA_MODEL.md`
