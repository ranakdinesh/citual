# Module: Identity (Citual)

## 1. Core Responsibilities
This module acts as the **Identity Provider (IdP)** for the Citual SaaS. It handles:
* **Authentication:** OAuth2 / OpenID Connect using **Ory Fosite v0.49**.
* **Authorization:** RBAC (Role-Based Access Control) + Tenant Isolation.
* **Tenancy:** Management of Ops vs. Customer tenants.
* **Audit:** Logging critical security events.

## 2. Strict Invariants (DO NOT VIOLATE)
1.  **Single Ops Tenant:** There can be only **one** tenant with `kind = 'ops'`.
2.  **Single Super Admin:** There can be only **one** user with `is_super_admin = true`.
3.  **Tenant Isolation:**
    * Every database query for tenant-scoped data MUST include `WHERE tenant_id = $1`.
    * API requests must validate that `JWT.tid` (Tenant ID) matches the requested resource's tenant (unless `is_super_admin`).
4.  **RBAC Strategy:**
    * **Tokens contain Roles, NOT Permissions.**
    * Permissions are resolved server-side.
    * Changing permissions increments `user.authz_version` to invalidate tokens/caches.

## 3. Database & SQL Standards
* **Split Strategy:** Keep migrations and queries separated by functional domain (e.g., `users.sql`, `fosite.sql`).
* **Foreign Keys:** Always use `ON DELETE RESTRICT` for critical relationships (don't delete a tenant if it has users).
* **Timestamps:** Use `timestamptz` (UTC).

## 4. JWT Token Claims Structure
* `sub`: User ID (UUID)
* `tid`: Tenant ID (UUID)
* `tk`: Tenant Kind (`ops` | `customer`)
* `roles`: Array of Role IDs `[]UUID`
* `sa`: Boolean (`is_super_admin`)
* `av`: Authorization Version (Integer)

internal/modules/identity/sql/
├── migrations/
│   ├── 0001_tenants.sql
│   ├── 0002_users.sql
│   ├── 0003_rbac.sql
│   ├── 0004_fosite.sql
│   └── 0005_audit.sql
└── queries/
├── tenants.sql
├── users.sql
├── rbac.sql
├── fosite.sql
└── audit.sql