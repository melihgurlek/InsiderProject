# Test script for the Concurrent Processing System
$baseUrl = "http://localhost:8080/api/v1"

Write-Host "=== Testing Concurrent Processing System ===" -ForegroundColor Green

# First, let's register a user and get a token
Write-Host "`n1. Registering a test user..." -ForegroundColor Yellow
$registerBody = @{
    username = "testuser"
    email    = "test@example.com"
    password = "password123"
} | ConvertTo-Json

$registerResponse = Invoke-RestMethod -Uri "$baseUrl/auth/register" -Method POST -Body $registerBody -ContentType "application/json"
Write-Host "User registered: $($registerResponse.message)" -ForegroundColor Green

# Login to get JWT token
Write-Host "`n2. Logging in to get JWT token..." -ForegroundColor Yellow
$loginBody = @{
    username = "testuser"
    password = "password123"
} | ConvertTo-Json

$loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
$token = $loginResponse.token
Write-Host "Login successful, token obtained" -ForegroundColor Green

# Set up headers for authenticated requests
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type"  = "application/json"
}

# Test worker health
Write-Host "`n3. Testing worker health..." -ForegroundColor Yellow
try {
    $healthResponse = Invoke-RestMethod -Uri "$baseUrl/worker/health" -Method GET -Headers $headers
    Write-Host "Worker Health: $($healthResponse.status) - $($healthResponse.message)" -ForegroundColor Green
}
catch {
    Write-Host "Worker health check failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test worker stats
Write-Host "`n4. Getting worker stats..." -ForegroundColor Yellow
try {
    $statsResponse = Invoke-RestMethod -Uri "$baseUrl/worker/stats" -Method GET -Headers $headers
    Write-Host "Worker Stats:" -ForegroundColor Green
    Write-Host "  Total Processed: $($statsResponse.total_processed)" -ForegroundColor Cyan
    Write-Host "  Successful Tasks: $($statsResponse.successful_tasks)" -ForegroundColor Cyan
    Write-Host "  Failed Tasks: $($statsResponse.failed_tasks)" -ForegroundColor Cyan
    Write-Host "  Queue Size: $($statsResponse.queue_size)" -ForegroundColor Cyan
    Write-Host "  Active Workers: $($statsResponse.active_workers)" -ForegroundColor Cyan
    Write-Host "  Avg Process Time: $($statsResponse.average_process_time_seconds) seconds" -ForegroundColor Cyan
}
catch {
    Write-Host "Worker stats failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test single task submission
Write-Host "`n5. Testing single task submission..." -ForegroundColor Yellow
$taskBody = @{
    type     = "credit"
    user_id  = 1
    amount   = 100.50
    priority = 5
} | ConvertTo-Json

try {
    $taskResponse = Invoke-RestMethod -Uri "$baseUrl/worker/tasks" -Method POST -Body $taskBody -Headers $headers
    Write-Host "Task submitted successfully:" -ForegroundColor Green
    Write-Host "  Task ID: $($taskResponse.task_id)" -ForegroundColor Cyan
    Write-Host "  Status: $($taskResponse.status)" -ForegroundColor Cyan
    Write-Host "  Message: $($taskResponse.message)" -ForegroundColor Cyan
}
catch {
    Write-Host "Task submission failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test batch task submission
Write-Host "`n6. Testing batch task submission..." -ForegroundColor Yellow
$batchBody = @{
    tasks = @(
        @{
            type     = "credit"
            user_id  = 1
            amount   = 50.25
            priority = 3
        },
        @{
            type     = "debit"
            user_id  = 1
            amount   = 25.10
            priority = 7
        },
        @{
            type     = "credit"
            user_id  = 1
            amount   = 75.00
            priority = 2
        }
    )
} | ConvertTo-Json

try {
    $batchResponse = Invoke-RestMethod -Uri "$baseUrl/worker/batch" -Method POST -Body $batchBody -Headers $headers
    Write-Host "Batch submitted successfully:" -ForegroundColor Green
    Write-Host "  Batch ID: $($batchResponse.batch_id)" -ForegroundColor Cyan
    Write-Host "  Status: $($batchResponse.status)" -ForegroundColor Cyan
    Write-Host "  Message: $($batchResponse.message)" -ForegroundColor Cyan
    Write-Host "  Task IDs: $($batchResponse.task_ids -join ', ')" -ForegroundColor Cyan
}
catch {
    Write-Host "Batch submission failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Wait a bit for processing
Write-Host "`n7. Waiting for tasks to process..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# Check stats again
Write-Host "`n8. Checking updated worker stats..." -ForegroundColor Yellow
try {
    $updatedStatsResponse = Invoke-RestMethod -Uri "$baseUrl/worker/stats" -Method GET -Headers $headers
    Write-Host "Updated Worker Stats:" -ForegroundColor Green
    Write-Host "  Total Processed: $($updatedStatsResponse.total_processed)" -ForegroundColor Cyan
    Write-Host "  Successful Tasks: $($updatedStatsResponse.successful_tasks)" -ForegroundColor Cyan
    Write-Host "  Failed Tasks: $($updatedStatsResponse.failed_tasks)" -ForegroundColor Cyan
    Write-Host "  Queue Size: $($updatedStatsResponse.queue_size)" -ForegroundColor Cyan
    Write-Host "  Active Workers: $($updatedStatsResponse.active_workers)" -ForegroundColor Cyan
}
catch {
    Write-Host "Updated worker stats failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test transfer task
Write-Host "`n9. Testing transfer task..." -ForegroundColor Yellow
$transferBody = @{
    type       = "transfer"
    user_id    = 1
    to_user_id = 2
    amount     = 10.00
    priority   = 8
} | ConvertTo-Json

try {
    $transferResponse = Invoke-RestMethod -Uri "$baseUrl/worker/tasks" -Method POST -Body $transferBody -Headers $headers
    Write-Host "Transfer task submitted successfully:" -ForegroundColor Green
    Write-Host "  Task ID: $($transferResponse.task_id)" -ForegroundColor Cyan
    Write-Host "  Status: $($transferResponse.status)" -ForegroundColor Cyan
}
catch {
    Write-Host "Transfer task submission failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== Concurrent Processing System Test Complete ===" -ForegroundColor Green 