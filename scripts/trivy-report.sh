#!/usr/bin/env bash
# Prints every finding of a Trivy JSON report as one run annotation, since a step
# can only raise ten. Annotations stay readable without the logs.
#   scripts/trivy-report.sh <report.json>
set -euo pipefail

findings="$(jq -r '.Results[]? | .Target as $t | (
    (.Vulnerabilities // [])[] | "\(.Severity) \($t) \(.PkgName) \(.InstalledVersion) fixed in \(.FixedVersion) \(.VulnerabilityID)"
  ), (
    (.Misconfigurations // [])[] | select(.Status == "FAIL") | "\(.Severity) \($t):\(.CauseMetadata.StartLine) \(.ID) \(.Title)"
  ), (
    (.Secrets // [])[] | "\(.Severity) \($t):\(.StartLine) \(.RuleID)"
  )' "$1" | sort -u)"

[ -z "$findings" ] && { echo "no findings in $1"; exit 0; }
echo "$findings"
echo "::error title=Trivy $1::$(printf '%s' "$findings" | awk 'BEGIN{ORS="%0A"} {print}')"
