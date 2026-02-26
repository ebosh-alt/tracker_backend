#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SPEC_FILE="${1:-$ROOT_DIR/openapi.yaml}"

if [[ ! -f "$SPEC_FILE" ]]; then
  echo "OpenAPI file not found: $SPEC_FILE"
  exit 1
fi

if command -v docker >/dev/null 2>&1; then
  SPEC_DIR="$(cd "$(dirname "$SPEC_FILE")" && pwd)"
  SPEC_NAME="$(basename "$SPEC_FILE")"

  echo "Running OpenAPI validation via openapi-generator-cli (docker)..."
  docker run --rm \
    -v "$SPEC_DIR:/local" \
    openapitools/openapi-generator-cli:v7.6.0 \
    validate -i "/local/$SPEC_NAME"
  echo "OpenAPI validation passed: $SPEC_FILE"
  exit 0
fi

echo "Docker not found. Falling back to YAML syntax check (not full OpenAPI validation)."
ruby -ryaml -e "data = YAML.load_file(ARGV[0]); raise 'missing `openapi` field' unless data.is_a?(Hash) && data['openapi'];" "$SPEC_FILE"
echo "YAML syntax check passed: $SPEC_FILE"
