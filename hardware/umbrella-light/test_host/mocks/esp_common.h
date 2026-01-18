#ifndef MOCK_ESP_COMMON_H
#define MOCK_ESP_COMMON_H

#include <stdint.h>

// --- Type definitions ---

struct station_config {
    uint8_t ssid[32];
    uint8_t password[64];
};

// --- Enums ---

enum {
    STATION_IDLE = 0,
    STATION_CONNECTING,
    STATION_WRONG_PASSWORD,
    STATION_NO_AP_FOUND,
    STATION_CONNECT_FAIL,
    STATION_GOT_IP
};

// --- Function declarations ---

int wifi_station_get_connect_status(void);
void wifi_set_opmode(int mode);
int wifi_station_set_config(void *config);
void wifi_station_connect(void);
int system_get_flash_size_map(void);
void user_init(void);
unsigned int user_rf_cal_sector_set(void);


#endif // MOCK_ESP_COMMON_H
