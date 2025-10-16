#!/bin/bash

set -o pipefail

print_divider() {
  echo "--------------------------------"
}

run_test() {
  local label="$1"
  local package="$2"
  local cmd=(go test -v "$package")
  local tmpfile

  print_divider
  echo "$label"
  print_divider

  tmpfile=$(mktemp)

  if "${cmd[@]}" 2>&1 | tee "$tmpfile"; then
    echo "✅ ${label} tests passed"
    rm -f "$tmpfile"
  else
    echo "❌ ${label} tests failed"
    FAILED_TESTS+=("$label")
    FAILED_LOGS+=("$tmpfile")
  fi

  echo
}

cleanup() {
  for log_file in "${FAILED_LOGS[@]}"; do
    [ -f "$log_file" ] && rm -f "$log_file"
  done
}

trap cleanup EXIT

declare -a FAILED_TESTS=()
declare -a FAILED_LOGS=()

echo "Running tests..."

tests=(
  "auth|./auth/..."
  "cli|./cli/..."
  "jq|./jq/..."
  "config|./config/..."
  "registry|./registry/..."
  "server|./server/..."
  "main|."
)

for entry in "${tests[@]}"; do
  IFS='|' read -r label package <<< "$entry"
  run_test "$label" "$package"
done

if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
  print_divider
  echo "Failed test summary"
  print_divider

  for i in "${!FAILED_TESTS[@]}"; do
    label=${FAILED_TESTS[$i]}
    log_file=${FAILED_LOGS[$i]}

    echo "- ${label}"
    print_divider
    cat "$log_file"
    print_divider
    echo
  done

  exit 1
fi

print_divider
echo "🚀 All tests pass"
print_divider
