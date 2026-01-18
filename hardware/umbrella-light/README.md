# Umbrella Light — ESP8266 RTOS SDK

This project is a small application for the ESP8266 that demonstrates a FreeRTOS task. The original example blinked the onboard LED; it is intended to be expanded to call the umbrella API and indicate if rain is expected.

## Building and Flashing

This project is built using the Espressif RTOS SDK toolchain.

### Prerequisites

- ESP8266 RTOS SDK toolchain
- A C compiler (`gcc`) and `make` for running host-based tests

### Build and Flash Steps

1.  **Configure the project:**
    ```bash
    idf.py menuconfig
    ```

2.  **Build the project:**
    ```bash
    idf.py build
    ```

3.  **Flash the firmware:**
    ```bash
    idf.py -p /dev/ttyUSB0 flash
    ```
    (Replace `/dev/ttyUSB0` with the correct serial port for your device).

## Host-Based Unit Testing

This project includes a unit testing setup that runs on the host machine, without requiring any ESP8266 hardware.

### Running the Tests

1.  **Navigate to the test directory:**
    ```bash
    cd hardware/umbrella-light/test_host
    ```

2.  **Compile and run the tests:**
    ```bash
    make
    ```
