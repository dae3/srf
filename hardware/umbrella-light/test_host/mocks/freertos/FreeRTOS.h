#ifndef MOCK_FREERTOS_H
#define MOCK_FREERTOS_H

#include <stdint.h>

// Define minimal types and macros to satisfy main.c compilation on host
#define configTICK_RATE_HZ 1000

void vTaskDelay(const uint32_t ticks);
void xTaskCreate(void (*task_fn)(void *), const char *name, uint16_t stack_size, void *params, uint32_t priority, void *task_handle);
void vTaskDelete(void *task_handle);

// Test helpers
int was_task_created(const char* name);
void reset_task_creation_state();

#endif // MOCK_FREERTOS_H
