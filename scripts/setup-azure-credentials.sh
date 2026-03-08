#!/bin/bash
# =============================================================
# setup-azure-credentials.sh
# Sets up passwordless auth between:
#   - GitHub Actions → Azure (OIDC)
#   - GitHub Actions → ACR (AcrPush)
#   - GitHub Actions → Key Vault (Key Vault Secrets User)
#   - GitHub Actions → Container Apps (Contributor)
#   - Container Apps → ACR (Managed Identity, AcrPull)
#   - Container Apps → Key Vault (Managed Identity, Key Vault Secrets User)
# =============================================================

set -e

# ── Configuration (edit these) ────────────────────────────────
APP_NAME="github-actions-oidc-simplebank"
GITHUB_OWNER="longth077"
GITHUB_REPO="simplebank"
GITHUB_BRANCH="main"
RESOURCE_GROUP="simplebank-rg"
ACR_NAME="mysimplebankacr"
KEY_VAULT_NAME="simplebank-kv"
CONTAINER_APP_NAME="simplebank-app"
LOCATION="eastasia"
# ─────────────────────────────────────────────────────────────

echo "==> [1/8] Creating App Registration in Azure AD..."
APP_ID=$(az ad app create --display-name "$APP_NAME" --query appId -o tsv)
echo "    App ID: $APP_ID"

echo "==> [2/8] Creating Service Principal..."
az ad sp create --id "$APP_ID" > /dev/null

echo "==> [3/8] Creating Federated Identity Credential (OIDC)..."
az ad app federated-credential create \
  --id "$APP_ID" \
  --parameters "{
    \"name\": \"github-actions-main\",
    \"issuer\": \"https://token.actions.githubusercontent.com\",
    \"subject\": \"repo:${GITHUB_OWNER}/${GITHUB_REPO}:ref:refs/heads/${GITHUB_BRANCH}\",
    \"audiences\": [\"api://AzureADTokenExchange\"]
  }"
echo "    Federated credential created."

SUB_ID=$(az account show --query id -o tsv)
TENANT_ID=$(az account show --query tenantId -o tsv)

echo "==> [4/8] Assigning AcrPush role to Service Principal on ACR..."
az role assignment create \
  --assignee "$APP_ID" \
  --scope "/subscriptions/${SUB_ID}/resourceGroups/${RESOURCE_GROUP}/providers/Microsoft.ContainerRegistry/registries/${ACR_NAME}" \
  --role AcrPush

echo "==> [5/8] Assigning Key Vault Secrets User role to Service Principal..."
az keyvault update \
  --name "$KEY_VAULT_NAME" \
  --resource-group "$RESOURCE_GROUP" \
  --enable-rbac-authorization true

az role assignment create \
  --assignee "$APP_ID" \
  --scope "/subscriptions/${SUB_ID}/resourceGroups/${RESOURCE_GROUP}/providers/Microsoft.KeyVault/vaults/${KEY_VAULT_NAME}" \
  --role "Key Vault Secrets User"

echo "==> [6/8] Assigning Contributor role to Service Principal on Container App..."
az role assignment create \
  --assignee "$APP_ID" \
  --scope "/subscriptions/${SUB_ID}/resourceGroups/${RESOURCE_GROUP}/providers/Microsoft.App/containerApps/${CONTAINER_APP_NAME}" \
  --role Contributor

echo "==> [7/8] Enabling Managed Identity on Container App..."
az containerapp identity assign \
  --name "$CONTAINER_APP_NAME" \
  --resource-group "$RESOURCE_GROUP" \
  --system-assigned

PRINCIPAL_ID=$(az containerapp identity show \
  --name "$CONTAINER_APP_NAME" \
  --resource-group "$RESOURCE_GROUP" \
  --query principalId -o tsv)
echo "    Managed Identity Principal ID: $PRINCIPAL_ID"

echo "==> [7b/8] Granting AcrPull to Managed Identity..."
az role assignment create \
  --assignee "$PRINCIPAL_ID" \
  --scope "/subscriptions/${SUB_ID}/resourceGroups/${RESOURCE_GROUP}/providers/Microsoft.ContainerRegistry/registries/${ACR_NAME}" \
  --role AcrPull

echo "==> [7c/8] Granting Key Vault Secrets User to Managed Identity..."
az role assignment create \
  --assignee "$PRINCIPAL_ID" \
  --scope "/subscriptions/${SUB_ID}/resourceGroups/${RESOURCE_GROUP}/providers/Microsoft.KeyVault/vaults/${KEY_VAULT_NAME}" \
  --role "Key Vault Secrets User"

echo "==> [8/8] Configuring ACA to pull from ACR using Managed Identity..."
az containerapp registry set \
  --name "$CONTAINER_APP_NAME" \
  --resource-group "$RESOURCE_GROUP" \
  --server "${ACR_NAME}.azurecr.io" \
  --identity system

echo ""
echo "✅ Setup complete! Add these as GitHub Secrets (no passwords needed):"
echo "   AZURE_CLIENT_ID     = $APP_ID"
echo "   AZURE_TENANT_ID     = $TENANT_ID"
echo "   AZURE_SUBSCRIPTION_ID = $SUB_ID"
