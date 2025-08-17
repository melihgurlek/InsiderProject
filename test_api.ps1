# Test script for Backend Path API
Write-Host "Testing Backend Path API endpoints..." -ForegroundColor Green

# Test health endpoint
Write-Host "`n1. Testing health endpoint..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/test/health" -Method GET
    Write-Host "Health check: $($health.status)" -ForegroundColor Green
}
catch {
    Write-Host "Health check failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test echo endpoint
Write-Host "`n2. Testing echo endpoint..." -ForegroundColor Yellow
try {
    $echo = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/test/echo" -Method POST -ContentType "application/json" -Body '{"message":"Hello from PowerShell","number":123}'
    Write-Host "Echo response: $($echo.message), Number: $($echo.number), Echoed: $($echo.echoed)" -ForegroundColor Green
}
catch {
    Write-Host "Echo test failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test slow endpoint
Write-Host "`n3. Testing slow endpoint..." -ForegroundColor Yellow
try {
    $slow = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/test/slow" -Method GET
    Write-Host "Slow response: $($slow.message), Delay: $($slow.delay)" -ForegroundColor Green
}
catch {
    Write-Host "Slow test failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test metrics endpoint
Write-Host "`n4. Testing metrics endpoint..." -ForegroundColor Yellow
try {
    $metrics = Invoke-RestMethod -Uri "http://localhost:8080/metrics" -Method GET
    Write-Host "Metrics endpoint accessible" -ForegroundColor Green
}
catch {
    Write-Host "Metrics test failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`nAPI testing completed!" -ForegroundColor Green
Write-Host "`nAccess URLs:" -ForegroundColor Cyan
Write-Host "- API: http://localhost:8080" -ForegroundColor White
Write-Host "- Prometheus: http://localhost:9090" -ForegroundColor White
Write-Host "- Grafana: http://localhost:3000 (admin/admin)" -ForegroundColor White
Write-Host "- Jaeger: http://localhost:16686" -ForegroundColor White 