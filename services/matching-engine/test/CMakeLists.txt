set(This LimitOrderBookTests)

set(Sources
    LimitOrderBookTests.cpp
)

add_executable(${This} ${Sources})
target_link_libraries(${This} PUBLIC
    gtest_main
    MatchingEngine_lib
)

add_test(
    NAME ${This}
    COMMAND ${This}
)
