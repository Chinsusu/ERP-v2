#!/usr/bin/env sh
set -eu

usage() {
  echo "Usage: $0 dev|staging"
}

environment="${1:-dev}"
case "$environment" in
  dev|staging) ;;
  *)
    usage
    exit 2
    ;;
esac

root_dir="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
compose_file="$root_dir/infra/compose/docker-compose.$environment.yml"
env_file="$root_dir/infra/env/$environment.env"
example_env_file="$root_dir/infra/env/$environment.env.example"

if [ ! -f "$compose_file" ]; then
  echo "Missing compose file: $compose_file" >&2
  exit 1
fi

if [ ! -f "$env_file" ]; then
  env_file="$example_env_file"
  echo "Using example env file: $env_file"
fi

if [ ! -f "$env_file" ]; then
  echo "Missing env file: $env_file" >&2
  exit 1
fi

compose() {
  docker compose --env-file "$env_file" -f "$compose_file" "$@"
}

echo "Deploying ERP $environment"
if ! compose pull api worker web; then
  compose build api worker web
fi
compose up -d --remove-orphans postgres redis minio minio-init mailhog
compose --profile tools run --rm migrate

if [ "$environment" = "dev" ]; then
  compose --profile tools run --rm seed
fi

compose up -d --remove-orphans api worker web reverse-proxy
"$root_dir/infra/scripts/smoke-dev-staging.sh" "$environment"

echo "ERP $environment deployment finished"
echo "Access logs: docker compose --env-file \"$env_file\" -f \"$compose_file\" logs reverse-proxy"
