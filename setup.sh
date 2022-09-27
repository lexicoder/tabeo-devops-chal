#!/bin/bash
PROJECT_ID_DEFAULT="tabeo-devops-chal"
read -p "GCP Project ID [$PROJECT_ID_DEFAULT]: " PROJECT_ID
PROJECT_ID="${PROJECT_ID:-$PROJECT_ID_DEFAULT}"

RESOURCE_PREFIX_DEFAULT="lexicoder"
read -p "Prefix for resources to be created [$RESOURCE_PREFIX_DEFAULT]: " RESOURCE_PREFIX
RESOURCE_PREFIX="${RESOURCE_PREFIX:-$RESOURCE_PREFIX_DEFAULT}"

GITHUB_REPO_DEFAULT="lexicoder/tabeo-devops-chal"
read -p "Github repo to permit for workload identity [$GITHUB_REPO_DEFAULT]: " GITHUB_REPO
GITHUB_REPO="${GITHUB_REPO:-$GITHUB_REPO_DEFAULT}"

SERVICE_ACCOUNT_NAME="$RESOURCE_PREFIX"
WORKLOAD_IDENTITY_POOL_NAME="$RESOURCE_PREFIX"
WORKLOAD_IDENTITY_POOL_PROVIDER_NAME="$RESOURCE_PREFIX"

gcloud config set project $PROJECT_ID

BILLING_ENABLED=$(gcloud beta billing projects describe $PROJECT_ID --format=json | jq '.billingEnabled')
if [ $BILLING_ENABLED == 'falsee' ]
then
    echo "Project $PROJECT_ID does not have billing enabled!"
    exit 1
fi

gcloud services enable iamcredentials.googleapis.com --project "${PROJECT_ID}"
gcloud services enable container.googleapis.com --project "${PROJECT_ID}"
gcloud services enable sts.googleapis.com --project "${PROJECT_ID}"
gcloud services enable iamcredentials.googleapis.com --project "${PROJECT_ID}"
gcloud services enable cloudresourcemanager.googleapis.com --project "${PROJECT_ID}"
gcloud services enable sqladmin.googleapis.com  --project "${PROJECT_ID}"

gcloud iam service-accounts describe "$SERVICE_ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com" --project "${PROJECT_ID}"
if [ $? -eq 1 ]
then
    gcloud iam service-accounts create "$SERVICE_ACCOUNT_NAME" --project "${PROJECT_ID}"
fi

gcloud projects add-iam-policy-binding $PROJECT_ID --member="serviceAccount:$SERVICE_ACCOUNT_NAME@${PROJECT_ID}.iam.gserviceaccount.com" --role="roles/iam.workloadIdentityPoolAdmin"
gcloud projects add-iam-policy-binding $PROJECT_ID --member="serviceAccount:$SERVICE_ACCOUNT_NAME@${PROJECT_ID}.iam.gserviceaccount.com" --role="roles/iam.serviceAccountAdmin"
gcloud projects add-iam-policy-binding $PROJECT_ID --member="serviceAccount:$SERVICE_ACCOUNT_NAME@${PROJECT_ID}.iam.gserviceaccount.com" --role="roles/owner"

gcloud iam workload-identity-pools describe "$WORKLOAD_IDENTITY_POOL_NAME" --project="${PROJECT_ID}" --location="global" --format="value(name)"
if [ $? -eq 1 ]
then
    gcloud iam workload-identity-pools create "$WORKLOAD_IDENTITY_POOL_NAME" --project="${PROJECT_ID}" --location="global" --display-name="$WORKLOAD_IDENTITY_POOL_NAME pool"
fi

WORKLOAD_IDENTITY_POOL_ID=$(gcloud iam workload-identity-pools describe "$WORKLOAD_IDENTITY_POOL_NAME" --project="${PROJECT_ID}" --location="global" --format="value(name)")

gcloud iam workload-identity-pools providers describe "$WORKLOAD_IDENTITY_POOL_PROVIDER_NAME" --project="${PROJECT_ID}" --location="global" --workload-identity-pool="$WORKLOAD_IDENTITY_POOL_NAME" --format="value(name)"
if [ $? -eq 1 ]
then
    gcloud iam workload-identity-pools providers create-oidc "$WORKLOAD_IDENTITY_POOL_PROVIDER_NAME" --project="${PROJECT_ID}" --location="global" --workload-identity-pool="$WORKLOAD_IDENTITY_POOL_NAME" --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository" --issuer-uri="https://token.actions.githubusercontent.com"
fi

gcloud iam service-accounts add-iam-policy-binding "$SERVICE_ACCOUNT_NAME@${PROJECT_ID}.iam.gserviceaccount.com" --project="${PROJECT_ID}" --role="roles/iam.workloadIdentityUser" --member="principalSet://iam.googleapis.com/${WORKLOAD_IDENTITY_POOL_ID}/attribute.repository/${GITHUB_REPO}"

WORKLOAD_IDENTITY_PROVIDER=$(gcloud iam workload-identity-pools providers describe "$WORKLOAD_IDENTITY_POOL_PROVIDER_NAME" --project="${PROJECT_ID}" --location="global" --workload-identity-pool="$WORKLOAD_IDENTITY_POOL_NAME" --format="value(name)")

echo "WORKLOAD_IDENTITY_PROVIDER: $WORKLOAD_IDENTITY_PROVIDER"
echo "SERVICE_ACCOUNT: $SERVICE_ACCOUNT_NAME@${PROJECT_ID}.iam.gserviceaccount.com"

gh auth status
if [ $? -eq 1 ]
then
    gh auth login
fi

gcloud storage buckets describe gs://${PROJECT_ID}-terraform
if [ $? -eq 1 ]
then
    gcloud storage buckets create gs://tabeo-devops-chal-terraform --project="${PROJECT_ID}"
fi

gh secret set WORKLOAD_IDENTITY_PROVIDER --body "WORKLOAD_IDENTITY_PROVIDER" --app actions --repos $GITHUB_REPO
gh secret set SERVICE_ACCOUNT --body "SERVICE_ACCOUNT" --app actions --repos $GITHUB_REPO
