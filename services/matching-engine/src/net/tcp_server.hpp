#ifndef NET_TCP_SERVER_HPP
#define NET_TCP_SERVER_HPP

#include "../engine/asset_book_manager.hpp"

#include <cstdint>

namespace net {

// Single persistent connection, one at a time (Protocol v1). Blocking,
// single-threaded read-dispatch-write loop per connection - no separate
// matching thread, avoids cross-thread handoff cost.
class TcpServer {
public:
    explicit TcpServer(uint16_t port);

    void run(engine::AssetBookManager& books);

private:
    uint16_t port_;
};

}  // namespace net

#endif
