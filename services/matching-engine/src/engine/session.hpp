#ifndef ENGINE_SESSION_HPP
#define ENGINE_SESSION_HPP

#include "asset_book_manager.hpp"
#include "../protocol/message_types.hpp"
#include "../protocol/codec.hpp"

#include <cstdint>
#include <unordered_map>
#include <vector>

namespace engine {

struct OutgoingMessage {
    protocol::Header header;
    protocol::Buffer payload;
};

// One Session per TCP connection. order_id/shares/price are narrowed from
// protocol's 64-bit to Book's 32-bit int - fine if Trading Service assigns
// ids from a 32-bit sequence, unsafe otherwise.
class Session {
public:
    explicit Session(AssetBookManager& books);

    std::vector<OutgoingMessage> handle_new_limit(const protocol::Header& req, const protocol::NewLimitPayload& p);
    std::vector<OutgoingMessage> handle_new_market(const protocol::Header& req, const protocol::NewMarketPayload& p);
    std::vector<OutgoingMessage> handle_new_stop(const protocol::Header& req, const protocol::NewStopPayload& p);
    std::vector<OutgoingMessage> handle_new_stop_limit(const protocol::Header& req, const protocol::NewStopLimitPayload& p);
    std::vector<OutgoingMessage> handle_modify_limit(const protocol::Header& req, const protocol::ModifyLimitPayload& p);
    std::vector<OutgoingMessage> handle_modify_stop(const protocol::Header& req, const protocol::ModifyStopPayload& p);
    std::vector<OutgoingMessage> handle_modify_stop_limit(const protocol::Header& req, const protocol::ModifyStopLimitPayload& p);
    std::vector<OutgoingMessage> handle_cancel_order(const protocol::Header& req, const protocol::CancelOrderPayload& p);

private:
    enum class OrderKind { Limit, Stop, StopLimit };

    AssetBookManager& books_;
    std::unordered_map<uint64_t, OrderKind> order_kind_by_id_;
    uint64_t next_trade_id_ = 1;
    uint32_t next_sequence_number_ = 0;

    uint64_t now_millis() const;
    OutgoingMessage make(const protocol::Header& req, protocol::MsgType type, uint64_t request_id, protocol::Buffer payload);
    OutgoingMessage make_rejected(const protocol::Header& req, uint64_t order_id, protocol::RejectReason reason);

    template <typename PlaceFn>
    std::vector<OutgoingMessage> place_order(const protocol::Header& req, uint64_t order_id, int64_t requested_shares,
                                              bool remainder_can_rest, OrderKind resting_kind, PlaceFn place);
};

}  // namespace engine

#endif
