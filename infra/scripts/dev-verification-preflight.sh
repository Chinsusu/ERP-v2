#!/usr/bin/env sh
set -eu

usage() {
  echo "Usage: $0 report|cleanup|preflight" >&2
}

mode="${1:-preflight}"
case "$mode" in
  report|cleanup|preflight) ;;
  *)
    usage
    exit 2
    ;;
esac

tmp_root="${ERP_VERIFY_TMP_ROOT:-/tmp}"
min_free_mb="${ERP_VERIFY_MIN_FREE_MB:-2048}"
dry_run="${ERP_VERIFY_DRY_RUN:-0}"

case "$tmp_root" in
  /*) ;;
  *)
    echo "ERP_VERIFY_TMP_ROOT must be an absolute path: $tmp_root" >&2
    exit 2
    ;;
esac

if [ "$tmp_root" = "/" ]; then
  echo "Refusing to use / as ERP_VERIFY_TMP_ROOT" >&2
  exit 2
fi

report_disk() {
  echo "Disk state:"
  df -h / "$tmp_root" 2>/dev/null | awk 'NR == 1 || !seen[$1]++'
}

cleanup_path() {
  path="$1"
  case "$path" in
    "$tmp_root"/erp-v2-verify-*|"$tmp_root"/erp-v2-s9-*|"$tmp_root"/pnpm-store-erp-v2-*|"$tmp_root"/pnpm-store-s9*) ;;
    *)
      echo "Refusing to remove unexpected path: $path" >&2
      exit 2
      ;;
  esac

  if [ "$dry_run" = "1" ]; then
    echo "Would remove $path"
  else
    echo "Removing $path"
    rm -rf -- "$path"
  fi
}

cleanup_verification_tmp() {
  found=0
  for pattern in erp-v2-verify-* erp-v2-s9-* pnpm-store-erp-v2-* pnpm-store-s9*; do
    for path in "$tmp_root"/$pattern; do
      [ -e "$path" ] || continue
      found=1
      cleanup_path "$path"
    done
  done

  if [ "$found" = "0" ]; then
    echo "No task-local verification temp paths found under $tmp_root"
  fi
}

assert_free_space() {
  available_kb="$(df -Pk "$tmp_root" | awk 'NR == 2 { print $4 }')"
  required_kb="$((min_free_mb * 1024))"

  if [ "$available_kb" -lt "$required_kb" ]; then
    echo "Insufficient free space on $tmp_root: ${available_kb}KB available, ${required_kb}KB required" >&2
    exit 1
  fi

  echo "Free space check passed on $tmp_root: ${available_kb}KB available"
}

case "$mode" in
  report)
    report_disk
    ;;
  cleanup)
    report_disk
    cleanup_verification_tmp
    report_disk
    ;;
  preflight)
    report_disk
    cleanup_verification_tmp
    report_disk
    assert_free_space
    ;;
esac
