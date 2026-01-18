#ifndef MOCK_GPIO_H
#define MOCK_GPIO_H

#include <stdint.h>

// Define minimal types and macros to satisfy main.c compilation on host
#define PERIPHS_IO_MUX_GPIO2_U 0
#define FUNC_GPIO2 2

void PIN_FUNC_SELECT(uint32_t pin_name, uint32_t func);
void gpio_output_set(uint32_t set_mask, uint32_t clear_mask, uint32_t enable_mask, uint32_t disable_mask);

#endif // MOCK_GPIO_H
