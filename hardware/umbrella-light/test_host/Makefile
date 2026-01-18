# Makefile for Host-Based Unit Tests - Corrected Include Order

CC = gcc
CFLAGS = -Imocks -Imocks/freertos -I../src -I../include -Iunity/src -Wall -Wextra -g
LDFLAGS =

TARGET = run_tests

OBJS = main.o unity.o gpio.o freertos.o api_client.o esp_sdk_stubs.o test_main.o

.PHONY: all clean

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(LDFLAGS) -o $(TARGET) $(OBJS)
	./$(TARGET)

main.o:
	$(CC) $(CFLAGS) -c ../src/main.c -o main.o

unity.o:
	$(CC) $(CFLAGS) -c unity/src/unity.c -o unity.o

gpio.o:
	$(CC) $(CFLAGS) -c mocks/gpio.c -o gpio.o

freertos.o:
	$(CC) $(CFLAGS) -c mocks/freertos.c -o freertos.o

api_client.o:
	$(CC) $(CFLAGS) -c mocks/api_client.c -o api_client.o

esp_sdk_stubs.o:
	$(CC) $(CFLAGS) -c mocks/esp_sdk_stubs.c -o esp_sdk_stubs.o

test_main.o:
	$(CC) $(CFLAGS) -c test_main.c -o test_main.o

clean:
	rm -f *.o $(TARGET)
