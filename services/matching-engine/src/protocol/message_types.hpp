#ifndef PROTOCOL_MESSAGE_TYPES_HPP
#define PROTOCOL_MESSAGE_TYPES_HPP

#include <cstdint>

namespace protocol {

// Protocol v1. Little-endian, x86_64 only: Trading Service and Matching
// Engine both run on x86_64, so this skips byte-swap cost with no downside.
constexpr uint8_t kProtocolVersion = 1;
constexpr uint32_t kEngineVersion = 0x00010000;  // engine build version, major<<16|minor<<8|patch

enum class Side : uint8_t {
    Buy = 0,
    Sell = 1,
};

enum class MsgType : uint8_t {
    Hello = 0,
    HelloAck = 1,
    Heartbeat = 2,
    NewLimit = 3,
    NewMarket = 4,
    NewStop = 5,
    NewStopLimit = 6,
    ModifyLimit = 7,
    ModifyStop = 8,
    ModifyStopLimit = 9,
    CancelOrder = 10,
    Accepted = 11,
    Rejected = 12,
    Executed = 13,
    OrderDone = 14,
};

enum class RejectReason : uint8_t {
    InvalidAsset = 0,
    UnknownOrderId = 1,
    Throttled = 2,
    InsufficientShares = 3,
    Malformed = 4,
    VersionMismatch = 5,
    DuplicateRequest = 6,
    InvalidPrice = 7,
    InvalidQuantity = 8,
    EngineOverloaded = 9,
};

enum class FinalStatus : uint8_t {
    Filled = 0,
    PartiallyFilledResting = 1,
    Resting = 2,
    Cancelled = 3,
    PartiallyFilledDropped = 4,  // market order remainder, no liquidity left, doesn't rest
};

// 28 bytes on the wire. request_id correlates Accepted/Rejected with the
// request that produced them; Executed/OrderDone are async pushes keyed by
// order_id and carry request_id = 0 when not a direct response.
// flags bits (all reserved/unused today): bit0 IOC, bit1 FOK, bit2 POST_ONLY, bit3 REDUCE_ONLY.
struct Header {
    uint8_t version = kProtocolVersion;
    MsgType msg_type{};
    uint16_t flags = 0;
    uint32_t sequence_number = 0;
    uint64_t request_id = 0;
    uint64_t asset_id = 0;
    uint32_t payload_size = 0;
};

struct HelloPayload {
    uint32_t min_version;
    uint32_t max_version;
};

struct HelloAckPayload {
    uint32_t accepted_version;
    uint8_t ok;
    uint32_t engine_version;
};

struct NewLimitPayload {
    uint64_t order_id;
    Side side;
    int64_t shares;
    int64_t price_ticks;
};

struct NewMarketPayload {
    uint64_t order_id;
    Side side;
    int64_t shares;
};

struct NewStopPayload {
    uint64_t order_id;
    Side side;
    int64_t shares;
    int64_t stop_price_ticks;
};

struct NewStopLimitPayload {
    uint64_t order_id;
    Side side;
    int64_t shares;
    int64_t limit_price_ticks;
    int64_t stop_price_ticks;
};

struct ModifyLimitPayload {
    uint64_t order_id;
    int64_t new_shares;
    int64_t new_price_ticks;
};

struct ModifyStopPayload {
    uint64_t order_id;
    int64_t new_shares;
    int64_t new_stop_price_ticks;
};

struct ModifyStopLimitPayload {
    uint64_t order_id;
    int64_t new_shares;
    int64_t new_limit_price_ticks;
    int64_t new_stop_price_ticks;
};

struct CancelOrderPayload {
    uint64_t order_id;
};

struct AcceptedPayload {
    uint64_t order_id;
    uint64_t timestamp;  // unix epoch milliseconds
    int64_t remaining_qty;
};

struct RejectedPayload {
    uint64_t order_id;
    RejectReason reason;
};

struct ExecutedPayload {
    uint64_t trade_id;
    uint64_t order_id;
    uint64_t matched_order_id;
    int64_t price_ticks;
    int64_t quantity;
    int64_t remaining_qty;
    uint64_t timestamp;  // unix epoch milliseconds
};

struct OrderDonePayload {
    uint64_t order_id;
    FinalStatus final_status;
    int64_t remaining_qty;
};

}  // namespace protocol

#endif
