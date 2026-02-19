#!/usr/bin/env bash
set -euo pipefail

ROOT_INPUT="${1:-}"
if [[ -z "${ROOT_INPUT}" ]]; then
  echo "Usage: $0 <ExamplesPath>"
  exit 2
fi

# Resolve absolute path
resolve_path() {
  local p="$1"
  if command -v realpath >/dev/null 2>&1; then
    realpath -m "$p"
    return 0
  fi
  if command -v readlink >/dev/null 2>&1; then
    if [[ -e "$p" ]]; then
      readlink -f "$p"
      return 0
    fi
  fi
  echo "$p"
}

RESOLVED="$(resolve_path "$ROOT_INPUT")"

if [[ ! -d "$RESOLVED" ]]; then
  echo "Error: not a directory: ${ROOT_INPUT}"
  echo "Resolved: ${RESOLVED}"
  echo
  echo "Tip: from repo root, this should work:"
  echo "  ./cmd/cleanup.sh ./Examples/Simple/"
  exit 1
fi

ROOT="$RESOLVED"

# Guardrails
if [[ "$ROOT" == "/" ]]; then
  echo "Refusing to run on /"
  exit 1
fi

echo "Cleaning results/, comparisons/, advocateResult/, advocateTrace/ under:"
echo "  ${ROOT}"
echo

# Find matching directories and prune them 
mapfile -t TARGETS < <(
  find "$ROOT" -type d \( \
    -name results -o \
    -name comparisons -o \
    -name advocateResult -o \
    -name advocateTrace \
  \) -prune -print
)

if [[ ${#TARGETS[@]} -eq 0 ]]; then
  echo "No matching directories found."
  exit 0
fi

# Confirm before deleting
echo "Will remove:"
for d in "${TARGETS[@]}"; do
  echo "  ${d}"
done
echo

read -r -p "Proceed? (y/N): " ans
if [[ "${ans}" != "y" && "${ans}" != "Y" ]]; then
  echo "Aborted."
  exit 0
fi

for d in "${TARGETS[@]}"; do
  d="${d//$'\r'/}"
  rm -rf -- "${d}"
done

echo "Done."
