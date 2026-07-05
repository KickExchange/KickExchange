#include "codec.hpp"

#include <stdexcept>

namespace protocol {

namespace {

void put_u8(Buffer& out, uint8_t v) { out.push_back(v); }

void put_u16(Buffer& out, uint16_t v) {
    out.push_back(static_cast<uint8_t>(v));
    out.push_back(static_cast<uint8_t>(v >> 8));
}

void put_u32(Buffer& out, uint32_t v) {
    for (int i = 0; i < 4; ++i) out.push_back(static_cast<uint8_t>(v >> (8 * i)));
}

void put_u64(Buffer& out, uint64_t v) {
    for (int i = 0; i < 8; ++i) out.push_back(static_cast<uint8_t>(v >> (8 * i)));
}

void put_i64(Buffer& out, int64_t v) { put_u64(out, static_cast<uint64_t>(v)); }

// Reads a fixed-width field from `data` at `offset`, advancing offset.
// Throws std::runtime_error if fewer than `width` bytes remain.
void check_len(size_t len, size_t offset, size_t width) {
    if (offset + width > len) {
        throw std::runtime_error("protocol codec: buffer too short");
    }
}

uint8_t get_u8(const uint8_t* data, size_t len, size_t& offset) {
    check_len(len, offset, 1);
    uint8_t v = data[offset];
    offset += 1;
    return v;
}

uint16_t get_u16(const uint8_t* data, size_t len, size_t& offset) {
    check_len(len, offset, 2);
    uint16_t v = static_cast<uint16_t>(data[offset]) | (static_cast<uint16_t>(data[offset + 1]) << 8);
    offset += 2;
    return v;
}

uint32_t get_u32(const uint8_t* data, size_t len, size_t& offset) {
    check_len(len, offset, 4);
    uint32_t v = 0;
    for (int i = 0; i < 4; ++i) v |= static_cast<uint32_t>(data[offset + i]) << (8 * i);
    offset += 4;
    return v;
}

uint64_t get_u64(const uint8_t* data, size_t len, size_t& offset) {
    check_len(len, offset, 8);
    uint64_t v = 0;
    for (int i = 0; i < 8; ++i) v |= static_cast<uint64_t>(data[offset + i]) << (8 * i);
    offset += 8;
    return v;
}

int64_t get_i64(const uint8_t* data, size_t len, size_t& offset) {
    return static_cast<int64_t>(get_u64(data, len, offset));
}

}  // namespace

void encode_header(Buffer& out, const Header& h) {
    put_u8(out, h.version);
    put_u8(out, static_cast<uint8_t>(h.msg_type));
    put_u16(out, h.flags);
    put_u32(out, h.sequence_number);
    put_u64(out, h.request_id);
    put_u64(out, h.asset_id);
    put_u32(out, h.payload_size);
}

Header decode_header(const uint8_t* data, size_t len) {
    size_t offset = 0;
    Header h;
    h.version = get_u8(data, len, offset);
    h.msg_type = static_cast<MsgType>(get_u8(data, len, offset));
    h.flags = get_u16(data, len, offset);
    h.sequence_number = get_u32(data, len, offset);
    h.request_id = get_u64(data, len, offset);
    h.asset_id = get_u64(data, len, offset);
    h.payload_size = get_u32(data, len, offset);
    return h;
}

void encode_hello(Buffer& out, const HelloPayload& p) {
    put_u32(out, p.min_version);
    put_u32(out, p.max_version);
}

HelloPayload decode_hello(const uint8_t* data, size_t len) {
    size_t offset = 0;
    HelloPayload p;
    p.min_version = get_u32(data, len, offset);
    p.max_version = get_u32(data, len, offset);
    return p;
}

void encode_hello_ack(Buffer& out, const HelloAckPayload& p) {
    put_u32(out, p.accepted_version);
    put_u8(out, p.ok);
    put_u32(out, p.engine_version);
}

HelloAckPayload decode_hello_ack(const uint8_t* data, size_t len) {
    size_t offset = 0;
    HelloAckPayload p;
    p.accepted_version = get_u32(data, len, offset);
    p.ok = get_u8(data, len, offset);
    p.engine_version = get_u32(data, len, offset);
    return p;
}

void encode_new_limit(Buffer& out, const NewLimitPayload& p) {
    put_u64(out, p.order_id);
    put_u8(out, static_cast<uint8_t>(p.side));
    put_i64(out, p.shares);
    put_i64(out, p.price_ticks);
}

NewLimitPayload decode_new_limit(const uint8_t* data, size_t len) {
    size_t offset = 0;
    NewLimitPayload p;
    p.order_id = get_u64(data, len, offset);
    p.side = static_cast<Side>(get_u8(data, len, offset));
    p.shares = get_i64(data, len, offset);
    p.price_ticks = get_i64(data, len, offset);
    return p;
}

void encode_new_market(Buffer& out, const NewMarketPayload& p) {
    put_u64(out, p.order_id);
    put_u8(out, static_cast<uint8_t>(p.side));
    put_i64(out, p.shares);
}

NewMarketPayload decode_new_market(const uint8_t* data, size_t len) {
    size_t offset = 0;
    NewMarketPayload p;
    p.order_id = get_u64(data, len, offset);
    p.side = static_cast<Side>(get_u8(data, len, offset));
    p.shares = get_i64(data, len, offset);
    return p;
}

void encode_new_stop(Buffer& out, const NewStopPayload& p) {
    put_u64(out, p.order_id);
    put_u8(out, static_cast<uint8_t>(p.side));
    put_i64(out, p.shares);
    put_i64(out, p.stop_price_ticks);
}

NewStopPayload decode_new_stop(const uint8_t* data, size_t len) {
    size_t offset = 0;
    NewStopPayload p;
    p.order_id = get_u64(data, len, offset);
    p.side = static_cast<Side>(get_u8(data, len, offset));
    p.shares = get_i64(data, len, offset);
    p.stop_price_ticks = get_i64(data, len, offset);
    return p;
}

void encode_new_stop_limit(Buffer& out, const NewStopLimitPayload& p) {
    put_u64(out, p.order_id);
    put_u8(out, static_cast<uint8_t>(p.side));
    put_i64(out, p.shares);
    put_i64(out, p.limit_price_ticks);
    put_i64(out, p.stop_price_ticks);
}

NewStopLimitPayload decode_new_stop_limit(const uint8_t* data, size_t len) {
    size_t offset = 0;
    NewStopLimitPayload p;
    p.order_id = get_u64(data, len, offset);
    p.side = static_cast<Side>(get_u8(data, len, offset));
    p.shares = get_i64(data, len, offset);
    p.limit_price_ticks = get_i64(data, len, offset);
    p.stop_price_ticks = get_i64(data, len, offset);
    return p;
}

void encode_modify_limit(Buffer& out, const ModifyLimitPayload& p) {
    put_u64(out, p.order_id);
    put_i64(out, p.new_shares);
    put_i64(out, p.new_price_ticks);
}

ModifyLimitPayload decode_modify_limit(const uint8_t* data, size_t len) {
    size_t offset = 0;
    ModifyLimitPayload p;
    p.order_id = get_u64(data, len, offset);
    p.new_shares = get_i64(data, len, offset);
    p.new_price_ticks = get_i64(data, len, offset);
    return p;
}

void encode_modify_stop(Buffer& out, const ModifyStopPayload& p) {
    put_u64(out, p.order_id);
    put_i64(out, p.new_shares);
    put_i64(out, p.new_stop_price_ticks);
}

ModifyStopPayload decode_modify_stop(const uint8_t* data, size_t len) {
    size_t offset = 0;
    ModifyStopPayload p;
    p.order_id = get_u64(data, len, offset);
    p.new_shares = get_i64(data, len, offset);
    p.new_stop_price_ticks = get_i64(data, len, offset);
    return p;
}

void encode_modify_stop_limit(Buffer& out, const ModifyStopLimitPayload& p) {
    put_u64(out, p.order_id);
    put_i64(out, p.new_shares);
    put_i64(out, p.new_limit_price_ticks);
    put_i64(out, p.new_stop_price_ticks);
}

ModifyStopLimitPayload decode_modify_stop_limit(const uint8_t* data, size_t len) {
    size_t offset = 0;
    ModifyStopLimitPayload p;
    p.order_id = get_u64(data, len, offset);
    p.new_shares = get_i64(data, len, offset);
    p.new_limit_price_ticks = get_i64(data, len, offset);
    p.new_stop_price_ticks = get_i64(data, len, offset);
    return p;
}

void encode_cancel_order(Buffer& out, const CancelOrderPayload& p) { put_u64(out, p.order_id); }

CancelOrderPayload decode_cancel_order(const uint8_t* data, size_t len) {
    size_t offset = 0;
    CancelOrderPayload p;
    p.order_id = get_u64(data, len, offset);
    return p;
}

void encode_accepted(Buffer& out, const AcceptedPayload& p) {
    put_u64(out, p.order_id);
    put_u64(out, p.timestamp);
    put_i64(out, p.remaining_qty);
}

AcceptedPayload decode_accepted(const uint8_t* data, size_t len) {
    size_t offset = 0;
    AcceptedPayload p;
    p.order_id = get_u64(data, len, offset);
    p.timestamp = get_u64(data, len, offset);
    p.remaining_qty = get_i64(data, len, offset);
    return p;
}

void encode_rejected(Buffer& out, const RejectedPayload& p) {
    put_u64(out, p.order_id);
    put_u8(out, static_cast<uint8_t>(p.reason));
}

RejectedPayload decode_rejected(const uint8_t* data, size_t len) {
    size_t offset = 0;
    RejectedPayload p;
    p.order_id = get_u64(data, len, offset);
    p.reason = static_cast<RejectReason>(get_u8(data, len, offset));
    return p;
}

void encode_executed(Buffer& out, const ExecutedPayload& p) {
    put_u64(out, p.trade_id);
    put_u64(out, p.order_id);
    put_u64(out, p.matched_order_id);
    put_i64(out, p.price_ticks);
    put_i64(out, p.quantity);
    put_i64(out, p.remaining_qty);
    put_u64(out, p.timestamp);
}

ExecutedPayload decode_executed(const uint8_t* data, size_t len) {
    size_t offset = 0;
    ExecutedPayload p;
    p.trade_id = get_u64(data, len, offset);
    p.order_id = get_u64(data, len, offset);
    p.matched_order_id = get_u64(data, len, offset);
    p.price_ticks = get_i64(data, len, offset);
    p.quantity = get_i64(data, len, offset);
    p.remaining_qty = get_i64(data, len, offset);
    p.timestamp = get_u64(data, len, offset);
    return p;
}

void encode_order_done(Buffer& out, const OrderDonePayload& p) {
    put_u64(out, p.order_id);
    put_u8(out, static_cast<uint8_t>(p.final_status));
    put_i64(out, p.remaining_qty);
}

OrderDonePayload decode_order_done(const uint8_t* data, size_t len) {
    size_t offset = 0;
    OrderDonePayload p;
    p.order_id = get_u64(data, len, offset);
    p.final_status = static_cast<FinalStatus>(get_u8(data, len, offset));
    p.remaining_qty = get_i64(data, len, offset);
    return p;
}

size_t wire_size(MsgType type) {
    switch (type) {
        case MsgType::Hello: return 8;
        case MsgType::HelloAck: return 9;
        case MsgType::Heartbeat: return 0;
        case MsgType::NewLimit: return 25;
        case MsgType::NewMarket: return 17;
        case MsgType::NewStop: return 25;
        case MsgType::NewStopLimit: return 33;
        case MsgType::ModifyLimit: return 24;
        case MsgType::ModifyStop: return 24;
        case MsgType::ModifyStopLimit: return 32;
        case MsgType::CancelOrder: return 8;
        case MsgType::Accepted: return 24;
        case MsgType::Rejected: return 9;
        case MsgType::Executed: return 56;
        case MsgType::OrderDone: return 17;
    }
    throw std::runtime_error("wire_size: unknown msg_type");
}

}  // namespace protocol
