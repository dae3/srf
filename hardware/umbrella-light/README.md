# Umbrella Light — ESP8266 RTOS SDK migration

This small project was migrated from the esp8266 non-OS SDK to the esp8266-rtos-sdk.
The example blinks the onboard LED on a NodeMCU (ESP-12E) and runs as a FreeRTOS task.

Build & flash (PlatformIO):

```bash
# build
platformio run

# flash (auto-detect serial port or set upload_port in platformio.ini)
platformio run --target upload
```

Notes about the migration
- `platformio.ini` now sets `framework = esp8266-rtos-sdk`.
- `src/main.c` was rewritten to use a FreeRTOS task (`app_main` / `blink_task`).
- GPIO access uses conditional compilation: if the native driver header is
  available it will use `gpio_set_level`/`gpio_config`, otherwise it falls
  back to the legacy macros (`PIN_FUNC_SELECT` / `GPIO_OUTPUT_SET`) for
  maximum compatibility across SDK versions.
- `user_rf_cal_sector_set()` is implemented to choose the RF calibration
  sector based on the flash size map. For production code review the
  sector mapping for your device/partition layout.

Next steps you might want:
- Replace compatibility GPIO code with the native driver API if your
  target SDK exposes it (this repo already attempts to use it via
  conditional compilation).
- Verify `user_rf_cal_sector_set()` logic against your flash layout and
  adjust as needed (especially if using OTA or custom partitioning).
- Add CI checks that run `platformio run` to ensure the project builds on
  push.

If you'd like, I can add a small CI workflow file (GitHub Actions) that
runs PlatformIO build on pushes and PRs.
