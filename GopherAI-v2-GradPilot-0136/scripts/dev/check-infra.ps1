$ErrorActionPreference = "Continue"

$projectRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$composeFile = Join-Path $projectRoot "deploy\local\docker-compose.yml"

docker compose -f $composeFile ps

Write-Host ""
Write-Host "Port check:"
foreach ($port in @(3307, 6379, 5672, 15672)) {
    $conn = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
    if ($conn) {
        Write-Host "  $port OK"
    } else {
        Write-Host "  $port not listening"
    }
}

Write-Host ""
Write-Host "Service probes:"
docker exec gopherai-mysql mysqladmin -uroot -p123456 ping
docker exec gopherai-redis redis-cli ping
docker exec gopherai-rabbitmq rabbitmq-diagnostics -q ping
