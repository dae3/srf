#include "freertos/FreeRTOS.h"
#include <stdio.h>
#include <string.h>

// --- Mock state ---
static const char* last_created_task_name = NULL;
// ---

void vTaskDelay(const uint32_t ticks) {
    printf("MOCK: vTaskDelay(ticks=%d)\n", ticks);
}

void xTaskCreate(void (*task_fn)(void *), const char *name, uint16_t stack_size, void *params, uint32_t priority, void *task_handle) {
    printf("MOCK: xTaskCreate(name=%s)\n", name);
    last_created_task_name = name;
}

void vTaskDelete(void *task_handle) {
    printf("MOCK: vTaskDelete()\n");
}

// --- Test helpers for mock state ---
int was_task_created(const char* name) {
    if (last_created_task_name == NULL) {
        return 0;
    }
    return strcmp(last_created_task_name, name) == 0;
}

void reset_task_creation_state() {
    last_created_task_name = NULL;
}
