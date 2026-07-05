#include "engine/asset_book_manager.hpp"
#include "net/tcp_server.hpp"

#include <cstdint>
#include <cstdlib>

int main(int argc, char** argv) {
    uint16_t port = 9000;
    if (argc > 1) port = static_cast<uint16_t>(std::atoi(argv[1]));

    engine::AssetBookManager books;
    net::TcpServer server(port);
    server.run(books);
    return 0;
}
