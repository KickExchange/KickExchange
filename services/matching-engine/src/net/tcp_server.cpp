#include "tcp_server.hpp"
#include "../engine/session.hpp"
#include "../protocol/codec.hpp"

#include <arpa/inet.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <unistd.h>

#include <iostream>
#include <vector>

namespace net {

using namespace protocol;

namespace {

constexpr int kHeartbeatTimeoutSeconds = 15;

bool recv_exact(int sock, uint8_t* buf, size_t len) {
    size_t got = 0;
    while (got < len) {
        ssize_t n = recv(sock, buf + got, len - got, 0);
        if (n <= 0) return false;
        got += static_cast<size_t>(n);
    }
    return true;
}

bool send_all(int sock, const uint8_t* buf, size_t len) {
    size_t sent = 0;
    while (sent < len) {
        ssize_t n = send(sock, buf + sent, len - sent, 0);
        if (n <= 0) return false;
        sent += static_cast<size_t>(n);
    }
    return true;
}

bool send_message(int sock, const engine::OutgoingMessage& msg) {
    Buffer buf;
    encode_header(buf, msg.header);
    buf.insert(buf.end(), msg.payload.begin(), msg.payload.end());
    return send_all(sock, buf.data(), buf.size());
}

bool do_handshake(int sock) {
    uint8_t headerBuf[kHeaderSize];
    if (!recv_exact(sock, headerBuf, kHeaderSize)) return false;
    Header h = decode_header(headerBuf, kHeaderSize);
    if (h.msg_type != MsgType::Hello) return false;

    std::vector<uint8_t> payloadBuf(wire_size(MsgType::Hello));
    if (!recv_exact(sock, payloadBuf.data(), payloadBuf.size())) return false;
    HelloPayload hello = decode_hello(payloadBuf.data(), payloadBuf.size());

    bool ok = hello.min_version <= kProtocolVersion && kProtocolVersion <= hello.max_version;

    Header ackHeader;
    ackHeader.version = kProtocolVersion;
    ackHeader.msg_type = MsgType::HelloAck;
    ackHeader.sequence_number = 0;
    ackHeader.request_id = h.request_id;
    ackHeader.asset_id = 0;
    HelloAckPayload ack{kProtocolVersion, static_cast<uint8_t>(ok ? 1 : 0), kEngineVersion};
    Buffer ackPayload;
    encode_hello_ack(ackPayload, ack);
    ackHeader.payload_size = static_cast<uint32_t>(ackPayload.size());

    if (!send_message(sock, engine::OutgoingMessage{ackHeader, ackPayload})) return false;
    return ok;
}

void serve_connection(int sock, engine::AssetBookManager& books) {
    if (!do_handshake(sock)) {
        close(sock);
        return;
    }

    engine::Session session(books);

    while (true) {
        uint8_t headerBuf[kHeaderSize];
        if (!recv_exact(sock, headerBuf, kHeaderSize)) break;
        Header h = decode_header(headerBuf, kHeaderSize);

        std::vector<uint8_t> payloadBuf(wire_size(h.msg_type));
        if (!payloadBuf.empty() && !recv_exact(sock, payloadBuf.data(), payloadBuf.size())) break;

        std::vector<engine::OutgoingMessage> responses;
        switch (h.msg_type) {
            case MsgType::Heartbeat:
                continue;
            case MsgType::NewLimit:
                responses = session.handle_new_limit(h, decode_new_limit(payloadBuf.data(), payloadBuf.size()));
                break;
            case MsgType::NewMarket:
                responses = session.handle_new_market(h, decode_new_market(payloadBuf.data(), payloadBuf.size()));
                break;
            case MsgType::NewStop:
                responses = session.handle_new_stop(h, decode_new_stop(payloadBuf.data(), payloadBuf.size()));
                break;
            case MsgType::NewStopLimit:
                responses = session.handle_new_stop_limit(h, decode_new_stop_limit(payloadBuf.data(), payloadBuf.size()));
                break;
            case MsgType::ModifyLimit:
                responses = session.handle_modify_limit(h, decode_modify_limit(payloadBuf.data(), payloadBuf.size()));
                break;
            case MsgType::ModifyStop:
                responses = session.handle_modify_stop(h, decode_modify_stop(payloadBuf.data(), payloadBuf.size()));
                break;
            case MsgType::ModifyStopLimit:
                responses = session.handle_modify_stop_limit(h, decode_modify_stop_limit(payloadBuf.data(), payloadBuf.size()));
                break;
            case MsgType::CancelOrder:
                responses = session.handle_cancel_order(h, decode_cancel_order(payloadBuf.data(), payloadBuf.size()));
                break;
            default:
                break;
        }

        for (const auto& msg : responses) {
            if (!send_message(sock, msg)) return;
        }
    }

    close(sock);
}

}  // namespace

TcpServer::TcpServer(uint16_t port) : port_(port) {}

void TcpServer::run(engine::AssetBookManager& books) {
    int listener = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);

    int reuse = 1;
    setsockopt(listener, SOL_SOCKET, SO_REUSEADDR, &reuse, sizeof(reuse));

    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = INADDR_ANY;
    addr.sin_port = htons(port_);

    bind(listener, reinterpret_cast<sockaddr*>(&addr), sizeof(addr));
    listen(listener, 1);

    std::cout << "matching engine listening on port " << port_ << std::endl;

    while (true) {
        int client = accept(listener, nullptr, nullptr);
        if (client < 0) continue;

        timeval timeout{kHeartbeatTimeoutSeconds, 0};
        setsockopt(client, SOL_SOCKET, SO_RCVTIMEO, &timeout, sizeof(timeout));

        std::cout << "trading service connected" << std::endl;
        serve_connection(client, books);
        std::cout << "trading service disconnected" << std::endl;
    }
}

}  // namespace net
