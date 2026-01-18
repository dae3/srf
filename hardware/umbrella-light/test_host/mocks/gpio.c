#include "gpio.h"
#include <stdio.h>

void PIN_FUNC_SELECT(uint32_t pin_name, uint32_t func) {
    printf("MOCK: PIN_FUNC_SELECT(pin_name=%d, func=%d)\n", pin_name, func);
}

void gpio_output_set(uint32_t set_mask, uint32_t clear_mask, uint32_t enable_mask, uint32_t disable_mask) {
    printf("MOCK: gpio_output_set(set_mask=%d, clear_mask=%d, enable_mask=%d, disable_mask=%d)\n", set_mask, clear_mask, enable_mask, disable_mask);
}
