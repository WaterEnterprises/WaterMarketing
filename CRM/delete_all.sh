#!/system/bin/sh
# ╔══════════════════════════════════════════════════════╗
# ║  Google Messages - Select ALL & Delete                ║
# ║                                                      ║
# ║  Per round:                                          ║
# ║  1. Long-press → selection mode (1 selected)         ║
# ║  2. Tap 8 more → 9 selected                         ║
# ║  3. Full-page GENTLE scroll → brings 9 new into view ║
# ║  4. Tap all 9 positions → 9 MORE selected            ║
# ║  5. Delete 18 at once                                ║
# ║  6. Repeat                                           ║
# ║                                                      ║
# ║  KEY: Full-page scroll (1400px) ensures ALL old      ║
# ║  items scroll off, so tapping 9 only hits new ones.  ║
# ║  Scroll speed: 1167 px/s = gentle, no fling.         ║
# ╚══════════════════════════════════════════════════════╝
#
# Screen: 720x1600, items 160px tall
# Positions: 292, 452, 612, 772, 932, 1092, 1252, 1412, 1546
# Trash: (592, 132)  |  Confirm: (493, 944)

input keyevent 224
sleep 0.5

MAX_ROUNDS=20

for round in $(seq 1 $MAX_ROUNDS); do
  # Check every 3 rounds
  if [ $((round % 3)) -eq 1 ]; then
    uiautomator dump /sdcard/check.xml 2>/dev/null
    grep -q 'swipeableContainer' /sdcard/check.xml 2>/dev/null || break
  fi

  # ── Step 1: Long-press → selection mode ──
  input swipe 360 292 360 292 1200
  sleep 0.3
  input swipe 360 312 360 312 800
  sleep 0.4

  # ── Step 2: Tap 8 more (9 total selected) ──
  input tap 360 452
  input tap 360 612
  input tap 360 772
  input tap 360 932
  input tap 360 1092
  input tap 360 1252
  input tap 360 1412
  input tap 360 1546
  sleep 0.3

  # ── Step 3: Full-page gentle scroll ──
  # 1400px over 1200ms = 1167 px/s (scrolls past ~9 items)
  input swipe 360 1500 360 100 1200
  sleep 0.8

  # ── Step 4: Tap all 9 (all NEW, old ones scrolled off) ──
  input tap 360 292
  input tap 360 452
  input tap 360 612
  input tap 360 772
  input tap 360 932
  input tap 360 1092
  input tap 360 1252
  input tap 360 1412
  input tap 360 1546
  sleep 0.3

  # ── Step 5: Delete ALL selected (~18) ──
  input tap 592 132
  sleep 0.5
  input tap 493 944
  sleep 3.0
done
