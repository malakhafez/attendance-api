# Restaurant Branch Management System — API Contract

> **Single Source of Truth** for all backend services and the API Gateway.

**Version**: 1.0  
**Last Updated**: 2026-10-12  
**Maintained By**: Restaurant Backend Team

---

## 📖 Overview

The backend consists of **3 independent services** sitting behind a **single API Gateway**.

The Gateway is responsible for:
1. **JWT verification** (using the shared secret).
2. **Routing** requests to the correct service by path prefix.
3. **Injecting identity headers** for downstream services.

**Flow**: `Client → Gateway → (Auth / Branch / Attendance) → PostgreSQL`

---

## 🌐 Service Registry

| Service        | Port | Base URL                  | Repo                                                            |
|----------------|------|---------------------------|-----------------------------------------------------------------|
| Auth API       | 9001 | `http://localhost:9001`   | [auth-api](https://github.com/malakhafez/auth-api)              |
| Branch API     | 9002 | `http://localhost:9002`   | [branch-api](https://github.com/malakhafez/branch-api)          |
| Attendance API | 9003 | `http://localhost:9003`   | [attendance-api](https://github.com/malakhafez/attendance-api)  |

---

## 🔐 Authentication & Authorization

### JWT Secret (SHARED)

**All services** must trust the same JWT issued by Auth API.

**Algorithm**: HS256  
**Issuer**: `restaurant-backend`  
**Access Token TTL**: 15 minutes  
**Refresh Token TTL**: 30 days

### JWT Claims

| Claim       | Type   | Description                                |
|-------------|--------|--------------------------------------------|
| `sub`       | UUID   | User ID                                    |
| `role`      | String | `CEO` · `BRANCH_MANAGER` · `SUPPLY_MANAGER`|
| `branch_id` | UUID   | Branch ID (only for `BRANCH_MANAGER`)      |
| `mcp`       | Bool   | `true` if user must change password first  |

### Identity Headers (INJECTED BY GATEWAY)

After verifying the JWT, the Gateway **must** inject these headers:

| Header         | Source        | Required For                        |
|----------------|---------------|-------------------------------------|
| `X-User-ID`    | `sub`         | All protected endpoints             |
| `X-User-Role`  | `role`        | All protected endpoints             |
| `X-Branch-ID`  | `branch_id`   | Branch Managers only (else empty)   |

**⚠️ These headers MUST NEVER come from the client.**

### Public vs Protected

- **Public endpoints**: `/health`, `/api/v1/auth/login`, `/api/v1/auth/refresh`, `/api/v1/auth/logout`, Swagger UI.
- **Protected endpoints**: everything else — **must** have Gateway headers.

---

## 🛣️ Routing Table

### Auth API → `:9001`

| Method | Path                               | Auth            |
|--------|------------------------------------|-----------------|
| POST   | `/api/v1/auth/login`               | Public          |
| POST   | `/api/v1/auth/refresh`             | Public          |
| POST   | `/api/v1/auth/logout`              | Public          |
| GET    | `/api/v1/auth/me`                  | Authenticated   |
| POST   | `/api/v1/auth/change-password`     | Authenticated   |
| PATCH  | `/api/v1/auth/me/email`            | Authenticated   |
| POST   | `/api/v1/users`                    | CEO only        |
| GET    | `/api/v1/users`                    | CEO only        |
| PATCH  | `/api/v1/users/:id/status`         | CEO only        |
| POST   | `/api/v1/users/:id/reset-password` | CEO only        |
| POST   | `/api/v1/users/:id/unlock`         | CEO only        |

### Branch API → `:9002`

| Method | Path                     | Auth       |
|--------|--------------------------|------------|
| POST   | `/api/v1/branches`       | CEO only   |
| GET    | `/api/v1/branches`       | CEO + BM   |
| GET    | `/api/v1/branches/:id`   | CEO + BM   |
| PATCH  | `/api/v1/branches/:id`   | CEO only   |

### Attendance API → `:9003`

| Method | Path                                | Auth       |
|--------|-------------------------------------|------------|
| POST   | `/api/v1/attendance/check-in`       | BM + CEO   |
| POST   | `/api/v1/attendance/check-out`      | BM + CEO   |
| PATCH  | `/api/v1/attendance/deductions/:id` | BM + CEO   |
| POST   | `/api/v1/employees`                 | BM + CEO   |
| GET    | `/api/v1/employees`                 | BM + CEO   |
| GET    | `/api/v1/employees/:id`             | BM + CEO   |
| PATCH  | `/api/v1/employees/:id`             | BM + CEO   |
| PATCH  | `/api/v1/employees/:id/payment`     | BM + CEO   |
| POST   | `/api/v1/employees/:id/terminate`   | BM + CEO   |
| POST   | `/api/v1/overtime`                  | BM + CEO   |
| GET    | `/api/v1/overtime`                  | BM + CEO   |
| GET    | `/api/v1/overtime/:id`              | BM + CEO   |
| PATCH  | `/api/v1/overtime/:id/approve`      | BM + CEO   |
| PATCH  | `/api/v1/overtime/:id/reject`       | BM + CEO   |

---

## 🗄️ Shared Database

**Database**: `restaurant_db` (PostgreSQL)

### Table Ownership

| Table                | Owned By       | Depends On              |
|----------------------|----------------|-------------------------|
| `branches`           | Branch API     | —                       |
| `users`              | Auth API       | `branches`              |
| `refresh_tokens`     | Auth API       | `users`                 |
| `employees`          | Attendance API | `branches`              |
| `attendance_records` | Attendance API | `employees`, `branches` |
| `deductions`         | Attendance API | `attendance_records`, `employees`, `branches` |
| `overtime`           | Attendance API | `employees`, `branches` |