#!/usr/bin/env bash
set -euo pipefail

DATASET="${1:-}"
PROFILE="${2:-all-quick}"
LABEL="${3:-baseline}"

if [[ -z "$DATASET" ]]; then
  echo "Usage: $0 <datasetDir> [profile=all-quick] [label=baseline]"
  exit 2
fi

CSV="$DATASET/comparisons/cross-test/kind-analysis/profile-$PROFILE/label-$LABEL/cross_test.csv"
if [[ ! -f "$CSV" ]]; then
  echo "Could not find: $CSV"
  exit 1
fi

./cmd/aggregate_projects.py "$CSV"

# ./cmd/aggregate_projects.sh ./Examples/GoBench/goker/blocking all-quick baseline > project_summary.csv
