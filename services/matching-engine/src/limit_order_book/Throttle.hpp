#ifndef THROTTLE_HPP
#define THROTTLE_HPP

#include <chrono>

class Throttle {
public:
    Throttle(int max_orders, std::chrono::milliseconds interval);

    bool can_process_order();

private:
    int max_orders_per_interval;
    std::chrono::milliseconds interval_duration;
    int orders_processed_in_interval;
    std::chrono::steady_clock::time_point interval_start_time;
};

#endif // THROTTLE_HPP

