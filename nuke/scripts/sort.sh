#!/usr/bin/env bash
set -euo pipefail

INPUT="${1:-all.txt}"
PATTERNS="${2:-patterns.tsv}"

mkdir -p subtech
rm -f subtech/*.txt

while read -r name pattern; do
    [ -z "${name:-}" ] && continue
    [[ "$name" =~ ^# ]] && continue
    [ -z "${pattern:-}" ] && continue

    rg -i --no-heading --color never -e "$pattern" "$INPUT" \
      | sed 's/^\*\.//; s/^\.//' \
      | tr '[:upper:]' '[:lower:]' \
      | sort -u > "subtech/$name.txt"

    [ -s "subtech/$name.txt" ] || rm -f "subtech/$name.txt"
done < "$PATTERNS"