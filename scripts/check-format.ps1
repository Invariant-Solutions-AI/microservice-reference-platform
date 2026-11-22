# Copyright 2026 Invariant Solutions LLC
# SPDX-License-Identifier: Apache-2.0

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $root

$unformatted = @(gofmt -l shared services)
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
if ($unformatted.Count -gt 0) {
    Write-Error ("Go files need formatting:`n" + ($unformatted -join "`n"))
    exit 1
}
