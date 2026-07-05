#include "Throttle.hpp"
#include <thread> // Required for std::this_thread

Throttle::Throttle(int max_orders, std::chrono::milliseconds interval)
    : max_orders_per_interval(max_orders),
      interval_duration(interval),
      orders_processed_in_interval(0),
      interval_start_time(std::chrono::steady_clock::now()) {}

bool Throttle::can_process_order() {
    auto now = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::milliseconds>(now - interval_start_time);
 
    if (elapsed >= interval_duration) { 
        interval_start_time = now;
        orders_processed_in_interval = 1;  
        return true;
    }
 
    if (orders_processed_in_interval < max_orders_per_interval) { 
        orders_processed_in_interval++;
        return true;
    }
 
    return false;
}