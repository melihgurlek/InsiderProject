# Test Business Metrics Script
Write-Host "Testing Business Metrics Implementation..." -ForegroundColor Green

# Test user registration
Write-Host "`n1. Testing User Registration..." -ForegroundColor Yellow
$registerResponse = curl -s -X POST http://localhost:8080/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{"username":"testuser1","email":"test1@example.com","password":"password123"}'
Write-Host "Register Response: $registerResponse"

# Test user login
Write-Host "`n2. Testing User Login..." -ForegroundColor Yellow
$loginResponse = curl -s -X POST http://localhost:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"username":"testuser1","password":"password123"}'
Write-Host "Login Response: $loginResponse"

# Extract JWT token
$loginData = $loginResponse | ConvertFrom-Json
$token = $loginData.token

if ($token) {
    Write-Host "`n3. Testing Transaction Operations..." -ForegroundColor Yellow
    
    # Test credit transaction
    Write-Host "   - Testing Credit Transaction..."
    $creditResponse = curl -s -X POST http://localhost:8080/api/v1/transactions/credit `
      -H "Content-Type: application/json" `
      -H "Authorization: Bearer $token" `
      -d '{"user_id":1,"amount":100.50}'
    Write-Host "   Credit Response: $creditResponse"
    
    # Test debit transaction
    Write-Host "   - Testing Debit Transaction..."
    $debitResponse = curl -s -X POST http://localhost:8080/api/v1/transactions/debit `
      -H "Content-Type: application/json" `
      -H "Authorization: Bearer $token" `
      -d '{"user_id":1,"amount":25.00}'
    Write-Host "   Debit Response: $debitResponse"
    
    # Test balance check
    Write-Host "   - Testing Balance Check..."
    $balanceResponse = curl -s -X GET http://localhost:8080/api/v1/balances/current `
      -H "Authorization: Bearer $token"
    Write-Host "   Balance Response: $balanceResponse"
}

# Test business metrics API
Write-Host "`n4. Testing Business Metrics API..." -ForegroundColor Yellow
$metricsResponse = curl -s http://localhost:8080/api/v1/metrics/summary
Write-Host "Metrics Summary: $metricsResponse"

$kpisResponse = curl -s http://localhost:8080/api/v1/metrics/kpis
Write-Host "KPIs: $kpisResponse"

# Test Prometheus metrics
Write-Host "`n5. Checking Prometheus Metrics..." -ForegroundColor Yellow
$prometheusMetrics = curl -s http://localhost:8080/metrics
$userRegCount = ($prometheusMetrics | Select-String "user_registration_total").Count
$activeUsersCount = ($prometheusMetrics | Select-String "active_users").Count
$transactionCount = ($prometheusMetrics | Select-String "transaction_count_total").Count

Write-Host "   User Registration Metrics: $userRegCount"
Write-Host "   Active Users Metrics: $activeUsersCount"
Write-Host "   Transaction Metrics: $transactionCount"

Write-Host "`nBusiness Metrics Test Complete!" -ForegroundColor Green
Write-Host "Check Grafana at http://localhost:3000 (admin/admin) to see the dashboards." -ForegroundColor Cyan 