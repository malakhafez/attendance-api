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
