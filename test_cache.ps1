# Test script for Cache functionality
Write-Host "Testing Backend Path API Cache functionality..." -ForegroundColor Green

# Test cache endpoint multiple times
Write-Host "`n1. Testing cache endpoint (first request - should be MISS)..." -ForegroundColor Yellow
try {
    $response1 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/test/cache" -Method GET
    Write-Host "First response timestamp: $($response1.timestamp)" -ForegroundColor Green
}
catch {
    Write-Host "Cache test failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n2. Testing cache endpoint (second request - should be HIT)..." -ForegroundColor Yellow
try {
    $response2 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/test/cache" -Method GET
    Write-Host "Second response timestamp: $($response2.timestamp)" -ForegroundColor Green
    
    if ($response1.timestamp -eq $response2.timestamp) {
        Write-Host "✅ CACHE HIT: Timestamps match - response was served from cache!" -ForegroundColor Green
    }
    else {
        Write-Host "❌ CACHE MISS: Timestamps differ - response was not cached" -ForegroundColor Red
    }
}
catch {
    Write-Host "Cache test failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n3. Testing cache endpoint (third request - should still be HIT)..." -ForegroundColor Yellow
try {
    $response3 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/test/cache" -Method GET
    Write-Host "Third response timestamp: $($response3.timestamp)" -ForegroundColor Green
    
    if ($response1.timestamp -eq $response3.timestamp) {
        Write-Host "✅ CACHE HIT: Timestamps match - response was served from cache!" -ForegroundColor Green
    }
    else {
        Write-Host "❌ CACHE MISS: Timestamps differ - response was not cached" -ForegroundColor Red
    }
}
catch {
    Write-Host "Cache test failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test non-cached endpoint for comparison
Write-Host "`n4. Testing non-cached endpoint (health) for comparison..." -ForegroundColor Yellow
try {
    $health1 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/test/health" -Method GET
    $health2 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/test/health" -Method GET
    Write-Host "Health endpoint timestamps: $($health1.timestamp) vs $($health2.timestamp)" -ForegroundColor Cyan
    
    if ($health1.timestamp -ne $health2.timestamp) {
        Write-Host "✅ Health endpoint is NOT cached (timestamps differ)" -ForegroundColor Green
    }
    else {
        Write-Host "❌ Health endpoint might be cached (timestamps match)" -ForegroundColor Yellow
    }
}
catch {
    Write-Host "Health test failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`nCache testing completed!" -ForegroundColor Green
Write-Host "`nCache behavior:" -ForegroundColor Cyan
Write-Host "- GET requests to /api/v1/test/cache should be cached for 5 minutes" -ForegroundColor White
Write-Host "- GET requests to /api/v1/test/health should NOT be cached" -ForegroundColor White
Write-Host "- Check X-Cache header in responses for HIT/MISS indication" -ForegroundColor White 