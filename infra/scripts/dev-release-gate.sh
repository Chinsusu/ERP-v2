#!/usr/bin/env sh
set -eu

usage() {
  echo "Usage: $0 [dev]" >&2
}

environment="${1:-dev}"
case "$environment" in
  dev) ;;
  *)
    usage
    exit 2
    ;;
esac

root_dir="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
go_image="${ERP_RELEASE_GATE_GO_IMAGE:-golang:1.23}"
node_image="${ERP_RELEASE_GATE_NODE_IMAGE:-node:22}"
pnpm_version="${ERP_RELEASE_GATE_PNPM_VERSION:-9.15.4}"
skip_deploy="${ERP_RELEASE_GATE_SKIP_DEPLOY:-0}"
allow_dirty="${ERP_RELEASE_GATE_ALLOW_DIRTY:-0}"
tmp_root="${ERP_RELEASE_GATE_TMP_ROOT:-/tmp}"

case "$tmp_root" in
  /*) ;;
  *)
    echo "ERP_RELEASE_GATE_TMP_ROOT must be an absolute path: $tmp_root" >&2
    exit 2
    ;;
esac

if [ "$tmp_root" = "/" ]; then
  echo "Refusing to use / as ERP_RELEASE_GATE_TMP_ROOT" >&2
  exit 2
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required for the dev release gate" >&2
  exit 1
fi

if ! command -v git >/dev/null 2>&1; then
  echo "git is required for the dev release gate" >&2
  exit 1
fi

if [ "$allow_dirty" != "1" ] && [ -n "$(git -C "$root_dir" status --porcelain)" ]; then
  echo "Refusing to run release gate with a dirty workspace. Commit or stash changes, or set ERP_RELEASE_GATE_ALLOW_DIRTY=1." >&2
  exit 1
fi

checks_dir=""
cleanup() {
  if [ "$checks_dir" != "" ]; then
    rm -rf -- "$checks_dir"
  fi
}
trap cleanup EXIT HUP INT TERM

step() {
  echo
  echo "==> $1"
}

go_run() {
  docker run --rm -v "$checks_dir/apps/api:/work" -w /work "$go_image" "$@"
}

node_run() {
  docker run --rm -v "$checks_dir:/work" -w /work -e CI=1 -e NEXT_TELEMETRY_DISABLED=1 "$node_image" "$@"
}

pnpm_run() {
  node_run npx "pnpm@$pnpm_version" "$@"
}

step "Disk preflight"
"$root_dir/infra/scripts/dev-verification-preflight.sh" preflight

step "Prepare clean release-gate snapshot"
checks_dir="$(mktemp -d "$tmp_root/erp-v2-s9-release-gate-XXXXXX")"
git -C "$root_dir" archive --format=tar HEAD | tar -xf - -C "$checks_dir"
echo "Snapshot: $(git -C "$root_dir" rev-parse --short HEAD)"

step "Backend checks"
go_run sh -c 'test -z "$(gofmt -l .)" && go vet ./... && go test ./... && go build ./cmd/api && go build ./cmd/worker'

step "OpenAPI checks"
pnpm_run --package=@redocly/cli dlx redocly lint packages/openapi/openapi.yaml
pnpm_run openapi:contract
pnpm_run dlx openapi-typescript packages/openapi/openapi.yaml -o /tmp/erp-openapi-schema.ts

step "Frontend checks"
pnpm_run install --filter web --frozen-lockfile=false --lockfile=false
pnpm_run --filter web typecheck
pnpm_run --filter web test
pnpm_run --filter web build

if [ "$skip_deploy" = "1" ]; then
  step "Deploy and smoke"
  echo "Skipped because ERP_RELEASE_GATE_SKIP_DEPLOY=1"
else
  step "Deploy dev and run smoke"
  "$root_dir/infra/scripts/deploy-dev-staging.sh" "$environment"

  step "Capture dev deploy evidence"
  "$root_dir/infra/scripts/dev-deploy-evidence.sh" "$environment"
fi

echo
echo "Dev release gate passed"
