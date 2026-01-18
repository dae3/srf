/*
	Blink example for ESP8266 using the ESP8266 RTOS SDK (FreeRTOS)

	- Creates a FreeRTOS task that toggles the onboard LED (GPIO2 / D4)
	- Uses RTOS APIs (tasks, vTaskDelay) instead of non-OS timers
	- Builds under PlatformIO with `framework = esp8266-rtos-sdk`
*/

#include <stdio.h>
#include "esp_common.h"
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "gpio.h"
#include <string.h>

// (Use SDK headers from the framework; avoid local re-declarations that
// conflict with the framework's headers.)

/*
 * WiFi credentials handling (compile-time/backed by ignored config file):
 * - Prefer compile-time macros `WIFI_SSID` and `WIFI_PASS` supplied via
 *   build flags (e.g. `-D WIFI_SSID=\"MyNet\"`) or by creating a
 *   local `include/wifi_config.h` (see `include/wifi_config.h.example`).
 * - If neither is provided the macros fall back to empty strings.
 */

#include "wifi_config.h"
#include "api_client.h"


#ifndef WIFI_SSID
#define WIFI_SSID ""
#endif
#ifndef WIFI_PASS
#define WIFI_PASS ""
#endif

// Blink task: toggles the LED every 500ms and prints status to the console
static void blink_task(void *pvParameters) {
	(void)pvParameters;
	int led_on = 0;
	/* Ensure the pin is muxed to GPIO function and enable it as output.
	 * Use the mask-based gpio_output_set API so we explicitly enable the
	 * pin as output and drive it. This avoids ambiguity between different
	 * SDK header variants and macros.
	 */
	PIN_FUNC_SELECT(PERIPHS_IO_MUX_GPIO2_U, FUNC_GPIO2);
	/* Enable GPIO2 as output and set it HIGH (LED off on many boards).
	 * gpio_output_set(set_mask, clear_mask, enable_mask, disable_mask)
	 */


	for (;;) {
		led_on = !led_on;
	    gpio_output_set((1 << 2), 0, (1 << 2), 0);
		vTaskDelay((1000 * configTICK_RATE_HZ) / 1000);
	}

	vTaskDelete(NULL);
}

void app_main(void) {
	xTaskCreate(&blink_task, "blink_task", 256, NULL, 5, NULL);
}

// --- Runtime WiFi connect/monitoring -------------------------------------
// Uses the compile-time `WIFI_SSID` / `WIFI_PASS` macros (or empty strings).
// This will configure the chip as a WiFi station and attempt to connect.

static const char *wifi_status_name(int st) {
	switch (st) {
		case STATION_IDLE: return "IDLE";
		case STATION_CONNECTING: return "CONNECTING";
		case STATION_WRONG_PASSWORD: return "WRONG_PASSWORD";
		case STATION_NO_AP_FOUND: return "NO_AP_FOUND";
		case STATION_CONNECT_FAIL: return "CONNECT_FAIL";
		case STATION_GOT_IP: return "GOT_IP";
		default: return "UNKNOWN";
	}
}

static void wifi_monitor_task(void *pvParameters) {
	(void)pvParameters;
	static int api_called = 0;
	for (;;) {
		int status = wifi_station_get_connect_status();
		if (status == STATION_GOT_IP && !api_called) {
			/* Call the HTTPS-capable API helper once. This uses the SDK's
			 * esp_http_client which supports TLS. The helper will print
			 * the response and attempt a minimal JSON extraction of fields.
			 */
			api_client_get_https("https://sydney-umbrella.fly.dev/api");
			api_called = 1;
		}
		vTaskDelay((2000 * configTICK_RATE_HZ) / 1000);
	}
}

static void start_wifi_connect(void) {
	if (WIFI_SSID[0] == '\0') {
		// no SSID configured; skip starting WiFi
		return;
	}

	// Set station mode (1 is station mode in the SDK).
	wifi_set_opmode(1);

	struct station_config sta_conf;
	memset(&sta_conf, 0, sizeof(sta_conf));
#if defined(WIFI_SSID)
	strncpy((char *)sta_conf.ssid, WIFI_SSID, sizeof(sta_conf.ssid) - 1);
#endif
#if defined(WIFI_PASS)
	strncpy((char *)sta_conf.password, WIFI_PASS, sizeof(sta_conf.password) - 1);
#endif

	if (wifi_station_set_config(&sta_conf) != 0) {
		// failed to set config; abort
		return;
	}

	// Try to connect (non-blocking) and start a monitor task to log progress
	wifi_station_connect();
	xTaskCreate(&wifi_monitor_task, "wifi_monitor", 512, NULL, 3, NULL);
}

// Call start_wifi_connect from user_init so it runs at startup
void user_init(void) {
	app_main();
	start_wifi_connect();
}

// Provide a minimal RF calibration sector selector. A real application
// should compute this based on the flash size map. Returning 0 is a
// conservative placeholder for basic builds; adjust if the SDK complains
// or if you use OTA/flash areas explicitly.
unsigned int user_rf_cal_sector_set(void) {
	// Determine RF calibration sector based on flash size map.
	// Avoid including non-portable headers; system_get_flash_size_map()
	// returns a small integer code. Use numeric cases to remain portable
	// across SDK header layout differences.
	int size_map = system_get_flash_size_map();
	unsigned int rf_cal_sec = 0;

	switch (size_map) {
		case 0: // FLASH_SIZE_4M_MAP_256_256
			rf_cal_sec = 128 - 5;
			break;
		case 1: // FLASH_SIZE_8M_MAP_512_512
			rf_cal_sec = 256 - 5;
			break;
		case 2: // FLASH_SIZE_16M_MAP_1024_1024
			rf_cal_sec = 512 - 5;
			break;
		case 3: // FLASH_SIZE_32M_MAP_512_512_1024_1024
			rf_cal_sec = 1024 - 5;
			break;
		default:
			// Unknown map; fall back to a safe default
			rf_cal_sec = 0;
			break;
	}

	return rf_cal_sec;
}
