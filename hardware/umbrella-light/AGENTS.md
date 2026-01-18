# Agent Instructions for Umbrella Light

This document provides instructions for building and testing the `umbrella-light` firmware.

## Building the Firmware

The firmware is built using the Espressif RTOS SDK toolchain.

### Prerequisites

Ensure you have the ESP8266 RTOS SDK toolchain installed and configured correctly. You can find instructions on the [Espressif documentation website](https://docs.espressif.com/projects/esp8266-rtos-sdk/en/latest/get-started/index.html).

### Build Steps

1.  **Navigate to the project directory:**
    ```bash
    cd hardware/umbrella-light
    ```

2.  **Configure the project (if needed):**
    ```bash
    idf.py menuconfig
    ```

3.  **Build the project:**
    ```bash
    idf.py build
    ```

4.  **Flash the firmware to the device:**
    ```bash
    idf.py -p /dev/ttyUSB0 flash
    ```
    (Replace `/dev/ttyUSB0` with the correct serial port for your device).

## Running Host-Based Unit Tests

The unit tests are designed to run on a host machine (e.g., your local development machine) without needing any ESP8266 hardware.

### Prerequisites

- A C compiler (e.g., `gcc`)
- `make`

### Test Steps

1.  **Navigate to the host testing directory:**
    ```bash
    cd hardware/umbrella-light/test_host
    ```

2.  **Compile and run the tests:**
    ```bash
    make
    ```
