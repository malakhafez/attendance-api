# Attendance API

> Restaurant Branch Management System — Attendance, Employee & Overtime API

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)
![Gin](https://img.shields.io/badge/Gin-v1.9+-00ADD8?style=flat)

---

## 📖 Overview

This service provides three core modules:

- **Attendance** — Check-in / Check-out with automatic lateness deduction.
- **Employee Management** — Full CRUD with Daily / Monthly payment types.
- **Overtime** — Create, approve, and reject extra-hour requests.

The API is designed to run **behind an API Gateway**.  
It trusts the identity headers injected by the gateway after JWT verification.

---

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 15+
- Port `9003` available

### 1. Create the Database

```sql
CREATE DATABASE restaurant_db;