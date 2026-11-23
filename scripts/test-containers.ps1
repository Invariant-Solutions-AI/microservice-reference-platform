# Copyright 2026 Invariant Solutions LLC
# SPDX-License-Identifier: Apache-2.0

$ErrorActionPreference = "Stop"

$edgePort = if ($env:EDGE_PORT) { $env:EDGE_PORT } else { "8080" }
$ordersPort = if ($env:ORDERS_PORT) { $env:ORDERS_PORT } else { "8081" }
$edgeUrl = "http://127.0.0.1:$edgePort"
$ordersUrl = "http://127.0.0.1:$ordersPort"

function Wait-Ready {
    param([string]$Url)

    for ($attempt = 0; $attempt -lt 30; $attempt++) {
        try {
            $response = Invoke-WebRequest -Uri $Url -TimeoutSec 2
            if ($response.StatusCode -eq 200) { return }
        } catch {
            Start-Sleep -Seconds 1
        }
    }
    throw "Service did not become ready: $Url"
}

Wait-Ready "$edgeUrl/ready"
Wait-Ready "$ordersUrl/ready"

$running = @(docker compose ps --status running --services)
if ($LASTEXITCODE -ne 0) { throw "Failed to inspect Compose services" }
foreach ($service in @("edge", "orders", "postgres", "otel-collector")) {
    if ($running -notcontains $service) { throw "Compose service is not running: $service" }
}

foreach ($service in @("edge", "orders")) {
    $container = docker compose ps -q $service
    if ($LASTEXITCODE -ne 0 -or -not $container) { throw "Container not found: $service" }
    $user = docker inspect --format '{{.Config.User}}' $container
    if ($LASTEXITCODE -ne 0 -or $user -ne "nonroot:nonroot") {
        throw "Container is not configured to run as nonroot: $service"
    }
}

$body = @{
    customerId = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
    items = @(@{ productId = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"; quantity = 2 })
} | ConvertTo-Json -Depth 4 -Compress
$createdResponse = Invoke-WebRequest -Uri "$edgeUrl/api/v1/orders" -Method Post -ContentType "application/json" -Body $body
if ($createdResponse.StatusCode -ne 201) { throw "Order creation did not return 201" }
$created = $createdResponse.Content | ConvertFrom-Json
if (-not $created.orderId -or $created.customerId -ne "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb" -or
    @($created.items).Count -ne 1 -or $created.items[0].productId -ne "cccccccc-cccc-4ccc-8ccc-cccccccccccc" -or
    $created.items[0].quantity -ne 2) {
    throw "Order creation returned unexpected data"
}

$orderUrl = "$edgeUrl/api/v1/orders/$($created.orderId)"
$before = Invoke-WebRequest -Uri $orderUrl
docker compose restart orders
if ($LASTEXITCODE -ne 0) { throw "Failed to restart Orders" }
Wait-Ready "$ordersUrl/ready"
$after = Invoke-WebRequest -Uri $orderUrl
if ($before.Content -cne $after.Content) { throw "Order changed after Orders restarted" }

Write-Host "Container smoke test passed, including order persistence across an Orders restart."
