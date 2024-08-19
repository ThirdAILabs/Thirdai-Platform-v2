#!/bin/bash

set -e


docker run -d -p 5000:5000 --name registry registry:2

DOCKER_DIRECTORIES=("deployment_job" "llm_generation_job" "train_job" "thirdai_platform")
VERSION_TAG="v1.0.0"
REGISTRY="testthirdaiplatform.azurecr.io"
BRANCH_NAME="release-ci"

for DIR in "${DOCKER_DIRECTORIES[@]}"; do
  IMAGE_NAME="${DIR}_${BRANCH_NAME}"
  TAR_FILE="${IMAGE_NAME}.tar"
  
  if [ -f "$TAR_FILE" ]; then
    docker load -i "$TAR_FILE"
    docker tag "$REGISTRY/$IMAGE_NAME:$VERSION_TAG" "localhost:5000/$IMAGE_NAME:$VERSION_TAG"
    docker push "localhost:5000/$IMAGE_NAME:$VERSION_TAG"
  else
    echo "Warning: $TAR_FILE does not exist, skipping load and push commands for $IMAGE_NAME"
  fi
done

cat <<EOF > .env_config
DATABASE_URI="postgresql://postgres:newpassword@localhost:5432/model_bazaar"
PRIVATE_MODEL_BAZAAR_ENDPOINT="http://localhost:80/"
PUBLIC_MODEL_BAZAAR_ENDPOINT="http://localhost:80/"
NOMAD_ENDPOINT="http://localhost:4646/"
LICENSE_PATH="/Users/pratikqpranav/Downloads/ndb_enterprise_license.json"
GENAI_KEY=""
SHARE_DIR="/Users/pratikqpranav/ThirdAI/share"
JWT_SECRET="CsnCr3lebs9eJQ"
ADMIN_USERNAME="admin"
ADMIN_MAIL="admin@mail.com"
ADMIN_PASSWORD="password"
HASHICORP_VAULT_ENDPOINT="http://127.0.0.1:8200"
HASHICORP_VAULT_TOKEN="00000000-0000-0000-0000-000000000000"
TEST_ENVIRONMENT=True
RECOVERY_BUCKET_NAME="thirdai-recovery-1"
TRAIN_IMAGE_NAME="train_job_${BRANCH_NAME}"
DEPLOY_IMAGE_NAME="deployment_job_${BRANCH_NAME}"
GENERATION_IMAGE_NAME="llm_generation_job_${BRANCH_NAME}"
THIRDAI_PLATFORM_IMAGE_NAME="thirdai_platform_${BRANCH_NAME}"
TAG="${VERSION_TAG}"
EOF

echo ".env_config file has been created successfully from $CONFIG_FILE"
