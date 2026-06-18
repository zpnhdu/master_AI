$ErrorActionPreference = "Stop"

$projectRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$dataRoot = "D:\GopherAI-dev"
$composeFile = Join-Path $projectRoot "deploy\local\docker-compose.yml"

foreach ($dir in @("mysql", "redis", "rabbitmq")) {
    New-Item -ItemType Directory -Force -Path (Join-Path $dataRoot $dir) | Out-Null
}

docker compose -f $composeFile up -d
docker compose -f $composeFile ps

Write-Host ""
Write-Host "GopherAI infra is starting."
Write-Host "MySQL:    127.0.0.1:3307  root/123456  database=GopherAI"
Write-Host "Redis:    127.0.0.1:6379"
Write-Host "RabbitMQ: 127.0.0.1:5672   root/123456"
Write-Host "RabbitMQ management: http://localhost:15672"
