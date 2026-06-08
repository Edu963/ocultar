# 🆘 Support & Operations Learning Path

Welcome to the Support & Operations Path. This track is designed for IT administrators, support engineers, and SREs responsible for deploying, maintaining, and troubleshooting OCULTAR.

## 🎯 Learning Objectives
By the end of this path, you will be able to:
- Deploy OCULTAR using Docker and Docker Compose.
- Manage secrets using Doppler.
- Troubleshoot common connectivity and performance issues.
- Perform critical maintenance tasks like key rotation and vault backups.
- Interpret audit logs for security investigations.

---

## 📚 Curriculum

### 1. [Deployment Guide](./deployment-guide.md)
Step-by-step instructions for local and production deployments.

### 2. [Troubleshooting & Diagnostics](./troubleshooting.md)
How to identify and fix common issues with the refinery, vault, and SLM engine.

### 3. [Operational Runbooks](../runbooks/README.md)
Standard operating procedures for critical tasks.
- [Key Rotation](../runbooks/key-rotation.md)
- [Vault Backup & Recovery](../runbooks/vault-recovery.md)
- [Scaling OCULTAR](../runbooks/scaling.md)

### 4. [Monitoring & Logging](../architecture/security-model.md#5-tamper-proof-audit-logs)
Understanding the audit log format and integrating with SIEM.

---

## 🛠️ Essential Tools
- **Docker Desktop**: The primary deployment environment.
- **Doppler CLI**: For secure secret injection.
- **cURL**: For testing API endpoints directly.
- **DuckDB CLI**: For inspecting the local vault during debugging.

---

## 🆘 Getting Help
If you encounter an issue not covered in this academy:
1.  Check the [Troubleshooting Guide](./troubleshooting.md).
2.  Search the [Architecture Docs](../architecture/README.md) for deeper context.
3.  Contact the **OCULTAR Engineering Team** via Slack or Email.
