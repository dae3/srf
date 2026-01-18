#include <stdio.h>
#include <stdint.h>

// Mock implementations for various ESP SDK functions

int wifi_station_get_connect_status(void) {
    printf("MOCK: wifi_station_get_connect_status()\n");
    return 0; // STATION_IDLE
}

void wifi_set_opmode(int mode) {
    printf("MOCK: wifi_set_opmode(mode=%d)\n", mode);
}

int wifi_station_set_config(void *config) {
    printf("MOCK: wifi_station_set_config()\n");
    return 0;
}

void wifi_station_connect(void) {
    printf("MOCK: wifi_station_connect()\n");
}

int system_get_flash_size_map(void) {
    printf("MOCK: system_get_flash_size_map()\n");
    return 0;
}
