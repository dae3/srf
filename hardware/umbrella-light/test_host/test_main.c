#include "unity.h"
#include "freertos/FreeRTOS.h"

// Forward declarations for functions under test
void app_main(void);

void setUp(void) {
    // Reset mock state before each test
    reset_task_creation_state();
}

void tearDown(void) {
    // Clean up code, if any
}

void test_app_main_creates_blink_task(void) {
    // Sanity check that the task wasn't already created
    TEST_ASSERT_FALSE(was_task_created("blink_task"));

    // Call the function under test
    app_main();

    // Assert that the blink_task was created by app_main
    TEST_ASSERT_TRUE(was_task_created("blink_task"));
}

int main(void) {
    UNITY_BEGIN();
    RUN_TEST(test_app_main_creates_blink_task);
    return UNITY_END();
}
