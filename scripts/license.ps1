# Copyright 2026 Invariant Solutions LLC
# SPDX-License-Identifier: Apache-2.0

param(
    [switch]$Fix
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$copyright = "// Copyright 2026 Invariant Solutions LLC"
$spdx = "// SPDX-License-Identifier: Apache-2.0"

$extensions = @(".go", ".ps1", ".sh", ".yml", ".yaml")
$excludedDirectories = @(
    ".git", "vendor", "node_modules", "bin", "obj", "dist", "target",
    "coverage", "generated"
)

function Get-CommentHeader {
    param([string]$Path)

    $name = [IO.Path]::GetFileName($Path)
    if ($name -ieq "Dockerfile" -or $name -like "Dockerfile.*") {
        return "# Copyright 2026 Invariant Solutions LLC", "# SPDX-License-Identifier: Apache-2.0"
    }

    switch ([IO.Path]::GetExtension($Path).ToLowerInvariant()) {
        ".go" { return "// Copyright 2026 Invariant Solutions LLC", "// SPDX-License-Identifier: Apache-2.0" }
        default { return "# Copyright 2026 Invariant Solutions LLC", "# SPDX-License-Identifier: Apache-2.0" }
    }
}

function Is-Generated {
    param([string]$Path)

    $name = [IO.Path]::GetFileName($Path)
    return (
        $name.EndsWith(".pb.go", [StringComparison]::OrdinalIgnoreCase) -or
        $name.EndsWith("_generated.go", [StringComparison]::OrdinalIgnoreCase) -or
        $name.StartsWith("zz_generated", [StringComparison]::OrdinalIgnoreCase)
    )
}

$failed = [System.Collections.Generic.List[string]]::new()

Get-ChildItem -LiteralPath $root -Recurse -File -Force |
    Where-Object {
        $relative = $_.FullName.Substring($root.Length).TrimStart("\", "/")
        $parts = $relative -split '[\\/]'
        $excluded = $false

        foreach ($part in $parts) {
            if ($excludedDirectories -contains $part.ToLowerInvariant()) {
                $excluded = $true
                break
            }
        }

        if ($excluded) { return $false }
        if (Is-Generated $_.FullName) { return $false }

        $name = $_.Name
        $extension = [IO.Path]::GetExtension($name).ToLowerInvariant()
        return ($extensions -contains $extension -or $name -ieq "Dockerfile" -or $name -like "Dockerfile.*")
    } |
    ForEach-Object {
        $content = Get-Content -LiteralPath $_.FullName -Raw
        $copyrightHeader, $spdxHeader = Get-CommentHeader $_.FullName
        $hasExpectedHeaders =
            $content.StartsWith($copyrightHeader) -and
            $content.Contains($spdxHeader)

        if (-not $hasExpectedHeaders) {
            if ($Fix) {
                $newline = if ($content.Contains("`r`n")) { "`r`n" } else { "`n" }
                $lines = $content -split "`r?`n"
                if ($lines[0] -match '^(//|#) Copyright ') {
                    $lines = $lines[1..($lines.Length - 1)]
                }
                if ($lines.Length -gt 0 -and $lines[0] -match '^(//|#) SPDX-License-Identifier:') {
                    $lines = $lines[1..($lines.Length - 1)]
                }
                if ($lines.Length -gt 0 -and [string]::IsNullOrWhiteSpace($lines[0])) {
                    $lines = $lines[1..($lines.Length - 1)]
                }
                $normalized = ($lines -join $newline).TrimEnd()
                $withHeader = $copyrightHeader + $newline + $spdxHeader + $newline + $newline + $normalized
                Set-Content -LiteralPath $_.FullName -Value $withHeader -NoNewline
                Write-Host "Added license header: $($_.FullName.Substring($root.Length + 1))"
            } else {
                $failed.Add($_.FullName.Substring($root.Length + 1))
            }
        }
    }

if ($failed.Count -gt 0) {
    Write-Error ("License header check failed for:`n" + ($failed -join "`n"))
    exit 1
}