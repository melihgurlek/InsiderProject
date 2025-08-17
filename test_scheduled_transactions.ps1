# Test Scheduled Transactions API
# This script tests the scheduled transaction functionality

$baseUrl = "http://localhost:8080"

# Get JWT token first
Write-Host "1. Getting JWT token..." -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "adminpass"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.token
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type"  = "application/json"
    }
    Write-Host "Login successful, token obtained" -ForegroundColor Green
}
catch {
    Write-Host "Login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 1: Create a one-time scheduled credit transaction (1 minute from now)
Write-Host "`n2. Creating one-time scheduled credit transaction..." -ForegroundColor Yellow
$scheduleTime = (Get-Date).AddMinutes(1).ToString("yyyy-MM-ddTHH:mm:ssZ")
$creditBody = @{
    user_id     = 127
    amount      = 100.50
    type        = "credit"
    schedule_at = $scheduleTime
    recurring   = $false
    description = "Test one-time credit"
} | ConvertTo-Json

try {
    $creditResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/scheduled-transactions" -Method POST -Body $creditBody -Headers $headers
    Write-Host "One-time credit scheduled successfully:" -ForegroundColor Green
    Write-Host "  ID: $($creditResponse.id)" -ForegroundColor Cyan
    Write-Host "  Amount: $($creditResponse.amount)" -ForegroundColor Cyan
    Write-Host "  Schedule At: $($creditResponse.schedule_at)" -ForegroundColor Cyan
    Write-Host "  Status: $($creditResponse.status)" -ForegroundColor Cyan
    $creditId = $creditResponse.id
}
catch {
    Write-Host "Failed to create one-time credit: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Create a recurring daily debit transaction
Write-Host "`n3. Creating recurring daily debit transaction..." -ForegroundColor Yellow
$debitBody = @{
    user_id     = 127
    amount      = 25.00
    type        = "debit"
    schedule_at = $scheduleTime
    recurring   = $true
    recurrence  = "daily"
    max_runs    = 5
    description = "Test recurring daily debit"
} | ConvertTo-Json

try {
    $debitResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/scheduled-transactions" -Method POST -Body $debitBody -Headers $headers
    Write-Host "Recurring debit scheduled successfully:" -ForegroundColor Green
    Write-Host "  ID: $($debitResponse.id)" -ForegroundColor Cyan
    Write-Host "  Amount: $($debitResponse.amount)" -ForegroundColor Cyan
    Write-Host "  Recurrence: $($debitResponse.recurrence)" -ForegroundColor Cyan
    Write-Host "  Max Runs: $($debitResponse.max_runs)" -ForegroundColor Cyan
    Write-Host "  Next Run: $($debitResponse.next_run_at)" -ForegroundColor Cyan
    $debitId = $debitResponse.id
}
catch {
    Write-Host "Failed to create recurring debit: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Create a weekly transfer transaction
Write-Host "`n4. Creating weekly transfer transaction..." -ForegroundColor Yellow
$transferBody = @{
    user_id     = 127
    to_user_id  = 126
    amount      = 50.00
    type        = "transfer"
    schedule_at = $scheduleTime
    recurring   = $true
    recurrence  = "weekly"
    description = "Test weekly transfer"
} | ConvertTo-Json

try {
    $transferResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/scheduled-transactions" -Method POST -Body $transferBody -Headers $headers
    Write-Host "Weekly transfer scheduled successfully:" -ForegroundColor Green
    Write-Host "  ID: $($transferResponse.id)" -ForegroundColor Cyan
    Write-Host "  From User: $($transferResponse.user_id)" -ForegroundColor Cyan
    Write-Host "  To User: $($transferResponse.to_user_id)" -ForegroundColor Cyan
    Write-Host "  Amount: $($transferResponse.amount)" -ForegroundColor Cyan
    Write-Host "  Recurrence: $($transferResponse.recurrence)" -ForegroundColor Cyan
    $transferId = $transferResponse.id
    Write-Host "Transfer ID: $transferId" -ForegroundColor Cyan
}
catch {
    Write-Host "Failed to create weekly transfer: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Get scheduled transaction statistics
Write-Host "`n5. Getting scheduled transaction statistics..." -ForegroundColor Yellow
try {
    $statsResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/scheduled-transactions/stats" -Method GET -Headers $headers
    Write-Host "Scheduled Transaction Statistics:" -ForegroundColor Green
    Write-Host "  Total Scheduled: $($statsResponse.total_scheduled)" -ForegroundColor Cyan
    Write-Host "  Pending: $($statsResponse.pending_count)" -ForegroundColor Cyan
    Write-Host "  Completed: $($statsResponse.completed_count)" -ForegroundColor Cyan
    Write-Host "  Failed: $($statsResponse.failed_count)" -ForegroundColor Cyan
    Write-Host "  Cancelled: $($statsResponse.cancelled_count)" -ForegroundColor Cyan
    Write-Host "  Recurring: $($statsResponse.recurring_count)" -ForegroundColor Cyan
    Write-Host "  One-time: $($statsResponse.one_time_count)" -ForegroundColor Cyan
    if ($statsResponse.next_execution_time) {
        Write-Host "  Next Execution: $($statsResponse.next_execution_time)" -ForegroundColor Cyan
    }
}
catch {
    Write-Host "Failed to get statistics: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 5: List user's scheduled transactions
Write-Host "`n6. Listing user's scheduled transactions..." -ForegroundColor Yellow
try {
    $listResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/scheduled-transactions?user_id=127" -Method GET -Headers $headers
    Write-Host "User's Scheduled Transactions:" -ForegroundColor Green
    foreach ($tx in $listResponse) {
        Write-Host "  ID: $($tx.id), Type: $($tx.type), Amount: $($tx.amount), Status: $($tx.status), Recurring: $($tx.recurring)" -ForegroundColor Cyan
    }
}
catch {
    Write-Host "Failed to list transactions: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 6: Get a specific scheduled transaction
Write-Host "`n7. Getting specific scheduled transaction..." -ForegroundColor Yellow
if ($creditId) {
    try {
        $getResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/scheduled-transactions/$creditId" -Method GET -Headers $headers
        Write-Host "Scheduled Transaction Details:" -ForegroundColor Green
        Write-Host "  ID: $($getResponse.id)" -ForegroundColor Cyan
        Write-Host "  Type: $($getResponse.type)" -ForegroundColor Cyan
        Write-Host "  Amount: $($getResponse.amount)" -ForegroundColor Cyan
        Write-Host "  Status: $($getResponse.status)" -ForegroundColor Cyan
        Write-Host "  Schedule At: $($getResponse.schedule_at)" -ForegroundColor Cyan
        Write-Host "  Recurring: $($getResponse.recurring)" -ForegroundColor Cyan
        Write-Host "  Description: $($getResponse.description)" -ForegroundColor Cyan
    }
    catch {
        Write-Host "Failed to get transaction: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 7: Update a scheduled transaction
Write-Host "`n8. Updating scheduled transaction..." -ForegroundColor Yellow
if ($creditId) {
    $updateBody = @{
        amount      = 150.75
        description = "Updated test credit"
    } | ConvertTo-Json

    try {
        $updateResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/scheduled-transactions/$creditId" -Method PUT -Body $updateBody -Headers $headers
        Write-Host "Transaction updated successfully:" -ForegroundColor Green
        Write-Host "  New Amount: $($updateResponse.amount)" -ForegroundColor Cyan
        Write-Host "  New Description: $($updateResponse.description)" -ForegroundColor Cyan
    }
    catch {
        Write-Host "Failed to update transaction: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 8: Manual execution of scheduled transactions
Write-Host "`n9. Manually executing scheduled transactions..." -ForegroundColor Yellow
try {
    $executeResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/scheduled-transactions/execute" -Method POST -Headers $headers
    Write-Host "Manual execution completed:" -ForegroundColor Green
    Write-Host "  Message: $($executeResponse.message)" -ForegroundColor Cyan
    Write-Host "  Status: $($executeResponse.status)" -ForegroundColor Cyan
}
catch {
    Write-Host "Failed to execute transactions: $($_.Exception.Message)" -ForegroundColor Red
}

# Wait a bit for execution to complete
Write-Host "`n10. Waiting for execution to complete..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Test 9: Check updated statistics after execution
Write-Host "`n11. Checking updated statistics..." -ForegroundColor Yellow
try {
    $updatedStatsResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/scheduled-transactions/stats" -Method GET -Headers $headers
    Write-Host "Updated Statistics:" -ForegroundColor Green
    Write-Host "  Total Scheduled: $($updatedStatsResponse.total_scheduled)" -ForegroundColor Cyan
    Write-Host "  Pending: $($updatedStatsResponse.pending_count)" -ForegroundColor Cyan
    Write-Host "  Completed: $($updatedStatsResponse.completed_count)" -ForegroundColor Cyan
    Write-Host "  Failed: $($updatedStatsResponse.failed_count)" -ForegroundColor Cyan
}
catch {
    Write-Host "Failed to get updated statistics: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 10: Cancel a scheduled transaction
Write-Host "`n12. Cancelling a scheduled transaction..." -ForegroundColor Yellow
if ($debitId) {
    try {
        Invoke-RestMethod -Uri "$baseUrl/api/v1/scheduled-transactions/$debitId" -Method DELETE -Headers $headers
        Write-Host "Transaction cancelled successfully" -ForegroundColor Green
    }
    catch {
        Write-Host "Failed to cancel transaction: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 11: Check final statistics
Write-Host "`n13. Final statistics check..." -ForegroundColor Yellow
try {
    $finalStatsResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/scheduled-transactions/stats" -Method GET -Headers $headers
    Write-Host "Final Statistics:" -ForegroundColor Green
    Write-Host "  Total Scheduled: $($finalStatsResponse.total_scheduled)" -ForegroundColor Cyan
    Write-Host "  Pending: $($finalStatsResponse.pending_count)" -ForegroundColor Cyan
    Write-Host "  Completed: $($finalStatsResponse.completed_count)" -ForegroundColor Cyan
    Write-Host "  Failed: $($finalStatsResponse.failed_count)" -ForegroundColor Cyan
    Write-Host "  Cancelled: $($finalStatsResponse.cancelled_count)" -ForegroundColor Cyan
}
catch {
    Write-Host "Failed to get final statistics: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`nScheduled Transactions API testing completed!" -ForegroundColor Green 