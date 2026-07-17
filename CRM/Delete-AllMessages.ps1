#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Deletes ALL Google Messages conversations via ADB — fast.
.DESCRIPTION
    Uses a scroll-and-select strategy:
    1. Push delete_all.sh to the device
    2. Opens Google Messages
    3. Runs the scroll-and-select script — selects ~540 conversations,
       then deletes them all at once
    4. Repeats with delete_batch.sh for any remaining stragglers
.NOTES
    Requires: ADB in PATH, device connected with USB debugging enabled.
#>

$ErrorActionPreference = "Stop"
$ScriptsDir  = $PSScriptRoot
$DeviceTmp   = "/data/local/tmp"

function Write-Step($msg) {
    Write-Host "→ $msg" -ForegroundColor Cyan
}

function Write-OK {
    Write-Host "  ✓" -ForegroundColor Green
}

function Write-Fail {
    Write-Host "  ✖ FAILED" -ForegroundColor Red
}

# ── Step 0: Find ADB executable ──
function Find-Adb {
    $commonPaths = @(
        "adb",                                         # in PATH
        "C:\Android\SDK\platform-tools\adb.exe",
        "$env:LOCALAPPDATA\Android\Sdk\platform-tools\adb.exe",
        "$env:ANDROID_HOME\platform-tools\adb.exe",
        "$env:ANDROID_SDK_ROOT\platform-tools\adb.exe"
    )
    foreach ($path in $commonPaths) {
        $exe = if ($path -eq "adb") { Get-Command $path -ErrorAction SilentlyContinue } else { $path }
        if ($exe -and (Test-Path $exe)) { return $exe }
    }
    throw "ADB not found. Install Android SDK platform-tools or add it to your PATH."
}
$Adb = Find-Adb
Write-Host "→ Using ADB: $Adb" -ForegroundColor DarkGray

# ── Step 1: Verify ADB device ──
Write-Step "Checking ADB device..."
$devices = & $Adb devices -l 2>&1
if ($LASTEXITCODE -ne 0 -or $devices -notmatch '(?m)^[a-fA-F0-9]+\s+device') {
    Write-Fail
    throw "No authorised device connected. Check USB debugging and allow the prompt on your phone."
}
$serial = ($devices -split "`n" | Where-Object { $_ -match '^\S+\s+device' })[0].Split()[0]
Write-Host "  Device: $serial" -ForegroundColor Green

# ── Step 2: Push scripts to device ──
Write-Step "Pushing scripts to device..."
foreach ($script in @("delete_all.sh", "delete_batch.sh")) {
    $local = Join-Path $ScriptsDir $script
    if (-not (Test-Path $local)) { Write-Warning "  ⚠ $script not found, skipping"; continue }
    $null = & $Adb push "`"$local`"" "$DeviceTmp/$script" 2>&1
    if ($LASTEXITCODE -ne 0) { Write-Fail; throw "Failed to push $script" }
    $null = & $Adb shell "chmod 755 $DeviceTmp/$script"
}
Write-OK

# ── Step 3: Open Google Messages ──
Write-Step "Opening Google Messages..."
$null = & $Adb shell "input keyevent 3" 2>&1
Start-Sleep -Milliseconds 300
$null = & $Adb shell "am start -n com.google.android.apps.messaging/.ui.ConversationListActivity" 2>&1
Start-Sleep -Milliseconds 2000

# ── Step 4: Helper to check if conversations remain ──
function Test-ConversationsExist {
    $null = & $Adb shell "uiautomator dump 2>/dev/null"
    $xml = & $Adb shell "cat /sdcard/window_dump.xml" 2>&1
    if ([string]::IsNullOrEmpty($xml)) { return $true }  # assume yes on error
    return $xml -match 'swipeableContainer'
}

# ── Step 5: Run select-ALL-and-delete ──
Write-Step "Running select-all + delete (covers ~540 conversations)..."
Write-Host "  " -NoNewline
$null = & $Adb shell "sh $DeviceTmp/delete_all.sh" 2>&1
Write-OK
Start-Sleep -Seconds 3

# ── Step 6: Cleanup stragglers with batch delete ──
Write-Step "Cleaning up remaining conversations..."
$rounds = 0
for ($i = 0; $i -lt 15; $i++) {
    if (-not (Test-ConversationsExist)) {
        Write-Host "  ✓ No more conversations found!" -ForegroundColor Green
        break
    }
    $null = & $Adb shell "sh $DeviceTmp/delete_batch.sh" 2>&1
    $rounds++
    Write-Host "." -NoNewline -ForegroundColor Yellow
    Start-Sleep -Milliseconds 500

    # Quick check every 3 rounds
    if ($i % 3 -eq 2 -and -not (Test-ConversationsExist)) {
        break
    }
}

Write-Host ""
Write-Host ""
Write-Host "╔══════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  Done!                                    ║" -ForegroundColor Green
Write-Host "║  • Scroll-select-delete: 1 operation     ║" -ForegroundColor Green
Write-Host "║  • Batch cleanup rounds: $rounds          ║" -ForegroundColor Green
Write-Host "║  Check the device to verify.              ║" -ForegroundColor Green
Write-Host "╚══════════════════════════════════════════╝" -ForegroundColor Cyan
