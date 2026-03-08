# CI/CD Architecture: GitHub Actions + Azure

## Overview

This repository uses a fully **passwordless CI/CD pipeline** powered by:

- **GitHub Actions** – orchestrates build, test, and deploy
- **Azure Container Registry (ACR)** – stores Docker images
- **Azure Key Vault** – stores all secrets centrally
- **Azure Container Apps (ACA)** – runs the application

---

## Credential Flow

```
Developer
   │
   │ git push → main
   ▼
GitHub Actions Runner
   │
   │ OIDC Token (short-lived JWT, no password)
   ▼
Azure Active Directory (App Registration)
   │
   ├──► ACR          (AcrPush role)        → docker push image
   ├──► Key Vault    (KV Secrets User role) → fetch secrets
   └──► Container Apps (Contributor role)  → deploy new image

Azure Container Apps (Runtime)
   │
   │ Managed Identity (auto-rotated, no password)
   ├──► ACR       (AcrPull role) → pull image on startup
   └──► Key Vault (KV Secrets User role) → read secrets at runtime
```

---

## RBAC Role Assignments Summary

| From                  | To                    | Auth Method           | Role                    |
|-----------------------|-----------------------|-----------------------|-------------------------|
| GitHub Actions        | Azure AD              | OIDC Federated Token  | —                       |
| GitHub Actions        | ACR                   | OIDC → az acr login   | `AcrPush`               |
| GitHub Actions        | Key Vault             | OIDC → KV action      | `Key Vault Secrets User`|
| GitHub Actions        | Container Apps        | OIDC → az CLI         | `Contributor`           |
| Container Apps        | ACR                   | Managed Identity      | `AcrPull`               |
| Container Apps        | Key Vault             | Managed Identity      | `Key Vault Secrets User`|

---

## Setup Instructions

### Prerequisites
- Azure CLI installed and logged in
- Azure subscription with Owner/Contributor access
- GitHub repository admin access

### Step 1: Run the Setup Script

```bash
chmod +x scripts/setup-azure-credentials.sh
./scripts/setup-azure-credentials.sh
```

### Step 2: Add GitHub Secrets

Go to **Settings → Secrets and variables → Actions** and add:

| Secret Name             | Value                          |
|-------------------------|--------------------------------|
| `AZURE_CLIENT_ID`       | App Registration Client ID     |
| `AZURE_TENANT_ID`       | Azure AD Tenant ID             |
| `AZURE_SUBSCRIPTION_ID` | Azure Subscription ID          |

> ⚠️ No passwords or keys needed — OIDC handles authentication automatically.

### Step 3: Add Secrets to Key Vault

```bash
az keyvault secret set --vault-name simplebank-kv --name "DB-CONNECTION-STRING" --value "<your-db-connection-string>"
az keyvault secret set --vault-name simplebank-kv --name "DB-PASSWORD" --value "<your-db-password>"
az keyvault secret set --vault-name simplebank-kv --name "TOKEN-SYMMETRIC-KEY" --value "<your-token-symmetric-key>"
```

---

## Pipeline Steps

1. **Checkout** – clone the repository
2. **Azure Login (OIDC)** – authenticate to Azure using a short-lived federated token (no stored credentials)
3. **Get Secrets from Key Vault** – fetch runtime secrets from Azure Key Vault
4. **Set up Go** – install Go 1.22
5. **Run Tests** – execute `go test ./...`
6. **Build Docker Image** – build and tag with the commit SHA and `latest`
7. **Login & Push to ACR** – push the image to Azure Container Registry
8. **Deploy to Azure Container Apps** – update the running container app with the new image and secrets

---

## Security Considerations

- **No long-lived credentials** are stored in GitHub Secrets — only the three OIDC identifiers (`CLIENT_ID`, `TENANT_ID`, `SUBSCRIPTION_ID`).
- Secrets are stored in **Azure Key Vault** and fetched at deploy time; they are never written to disk or logs.
- The Container App uses a **System-Assigned Managed Identity** to pull images from ACR and read secrets from Key Vault at runtime, with no passwords involved.
- The GitHub Actions workflow only triggers on pushes to the `main` branch and runs in the `production` GitHub environment, enabling optional manual approval gates.
