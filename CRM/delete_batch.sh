#!/system/bin/sh
# Run ON the device — deletes one batch of visible conversations.
# Called in a loop by the PowerShell wrapper.
#
# Coordinates (720x1600 screen):
#   Conversation list starts at Y=212, each item 160px tall.
#   9 items visible. First selected by long-press, remaining 8 tapped.

# ---- Step 1: Long-press first conversation (center Y=292) ----
input swipe 360 292 360 292 1200
sleep 0.5

# ---- Step 1b: Retry at a slight offset in case first missed ----
input swipe 360 312 360 312 800
sleep 0.4

# ---- Step 2: Tap remaining 8 conversations ----
input tap 360 452
input tap 360 612
input tap 360 772
input tap 360 932
input tap 360 1092
input tap 360 1252
input tap 360 1412
input tap 360 1546

# ---- Step 3: Tap Trash button ----
sleep 0.3
input tap 592 132

# ---- Step 4: Confirm "Move to trash" ----
sleep 0.5
input tap 493 944

# ---- Step 5: Wait for deletion to process ----
sleep 1.5
