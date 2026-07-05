cmake_minimum_required(VERSION 3.20)

project(MatchingEngine VERSION 0.1.0)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -O2")

enable_testing()

include(FetchContent)
FetchContent_Declare(
    googletest
    GIT_REPOSITORY https://github.com/google/googletest.git
    GIT_TAG        v1.15.2
)
FetchContent_MakeAvailable(googletest)

set(Headers
    src/limit_order_book/Book.hpp
    src/limit_order_book/Limit.hpp
    src/limit_order_book/Order.hpp
    src/limit_order_book/Throttle.hpp
    src/protocol/message_types.hpp
    src/protocol/codec.hpp
    src/engine/asset_book_manager.hpp
    src/engine/session.hpp
    src/net/tcp_server.hpp
)
set(Sources
    src/limit_order_book/Book.cpp
    src/limit_order_book/Limit.cpp
    src/limit_order_book/Order.cpp
    src/limit_order_book/Throttle.cpp
    src/protocol/codec.cpp
    src/engine/asset_book_manager.cpp
    src/engine/session.cpp
    src/net/tcp_server.cpp
)

add_library(${PROJECT_NAME}_lib STATIC ${Sources} ${Headers})
target_include_directories(${PROJECT_NAME}_lib PUBLIC src)

add_executable(${PROJECT_NAME} src/main.cpp)
target_link_libraries(${PROJECT_NAME} PRIVATE ${PROJECT_NAME}_lib)

add_subdirectory(test)
