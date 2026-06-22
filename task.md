# Kap-App Project Architecture Decisions

All tasks have been successfully completed. Below are the core architectural decisions implemented to connect the Go backend to Supabase:

## 1. Port 5432 Selection (Session Port)
- **Decision:** We used port `5432` for `SUPABASE_DB_URL` instead of `6543`.
- **Rationale:** Port `6543` connects to PgBouncer in transaction pooling mode, which prevents using PostgreSQL session-level features or variables, and can cause connection issues. Port `5432` ensures direct session connections, allowing stable connections and preventing PgBouncer variable-mismatch errors.

## 2. SECURITY DEFINER Triggers and Functions
- **Decision:** All database-level user sync triggers and RLS utility functions (`get_current_user_tenant_ids`, `get_current_user_role_in_tenant`) are configured with `SECURITY DEFINER` and have their search path locked using `SET search_path = public`.
- **Rationale:** `SECURITY DEFINER` lets the function run with the privileges of the creator (bypass RLS for internal logic), which prevents recursive deadlocks during RLS checks when querying user roles/memberships. Locking the search path is a critical security measure to prevent search-path injection attacks.

## 3. Zero-Dependency Env Parser with SplitN
- **Decision:** Hand-crafted `.env` file parser inside `main.go` using `strings.SplitN(line, "=", 2)`.
- **Rationale:** Keeps the project stable without pulling external libraries or modifying `go.mod`. Using `SplitN(..., 2)` ensures that environment variables containing `=` characters (such as complex passwords, URLs, or secrets) are correctly parsed without splitting the password itself.

---

### Migration & Code Files Linked:
- Consolidated SQL Schema: [001_complete_schema.sql](file:///c:/Users/yigit/OneDrive/Desktop/kapp-app-master/kap-app-backend/migrations/001_complete_schema.sql)
- Backend Entrypoint: [main.go](file:///c:/Users/yigit/OneDrive/Desktop/kapp-app-master/kap-app-backend/cmd/api/main.go)
- Environment Config File: [.env](file:///c:/Users/yigit/OneDrive/Desktop/kapp-app-master/kap-app-backend/.env)
