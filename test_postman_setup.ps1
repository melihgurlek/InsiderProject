# Test Postman Setup Script
# This script creates a test user and logs in to get a token for Postman testing

Write-Host "Setting up test user for Postman..." -ForegroundColor Green

# Register a new test user
Write-Host "Registering test user..." -ForegroundColor Yellow
$registerBody = @{
    username = "postmantest"
    email    = "postmantest@example.com"
    password = "password123"
} | ConvertTo-Json

try {
    $registerResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/register" -Method POST -ContentType "application/json" -Body $registerBody
    Write-Host "User registered successfully!" -ForegroundColor Green
    Write-Host "User ID: $($registerResponse.user.id)" -ForegroundColor Cyan
}
catch {
    Write-Host "Registration failed or user already exists: $($_.Exception.Message)" -ForegroundColor Yellow
}

# Login to get token
Write-Host "Logging in..." -ForegroundColor Yellow
$loginBody = @{
    username = "postmantest"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -ContentType "application/json" -Body $loginBody
    Write-Host "Login successful!" -ForegroundColor Green
    Write-Host "Token: $($loginResponse.token)" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "=== POSTMAN SETUP INSTRUCTIONS ===" -ForegroundColor Magenta
    Write-Host "1. Set the 'token' variable in Postman to: $($loginResponse.token)" -ForegroundColor White
    Write-Host "2. Set the 'scheduleTime' variable to a future time (e.g., 2025-07-29T20:00:00Z)" -ForegroundColor White
    Write-Host "3. The user_id for testing is: $($loginResponse.user.id)" -ForegroundColor White
    Write-Host "4. You can now test all scheduled transaction endpoints!" -ForegroundColor White
}
catch {
    Write-Host "Login failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
}

Write-Host ""
Write-Host "Testing health endpoint..." -ForegroundColor Yellow
try {
    $healthResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/test/health" -Method GET
    Write-Host "Health check successful: $($healthResponse.status)" -ForegroundColor Green
}
catch {
    Write-Host "Health check failed: $($_.Exception.Message)" -ForegroundColor Red
} 