#ifndef PROTOCOL_CODEC_HPP
#define PROTOCOL_CODEC_HPP

#include "message_types.hpp"

#include <cstdint>
#include <cstddef>
#include <vector>

namespace protocol {

using Buffer = std::vector<uint8_t>;

constexpr size_t kHeaderSize = 1 + 1 + 2 + 4 + 8 + 8 + 4;  // 28 bytes

void encode_header(Buffer& out, const Header& h);
Header decode_header(const uint8_t* data, size_t len);

void encode_hello(Buffer& out, const HelloPayload& p);
HelloPayload decode_hello(const uint8_t* data, size_t len);

void encode_hello_ack(Buffer& out, const HelloAckPayload& p);
HelloAckPayload decode_hello_ack(const uint8_t* data, size_t len);

void encode_new_limit(Buffer& out, const NewLimitPayload& p);
NewLimitPayload decode_new_limit(const uint8_t* data, size_t len);

void encode_new_market(Buffer& out, const NewMarketPayload& p);
NewMarketPayload decode_new_market(const uint8_t* data, size_t len);

void encode_new_stop(Buffer& out, const NewStopPayload& p);
NewStopPayload decode_new_stop(const uint8_t* data, size_t len);

void encode_new_stop_limit(Buffer& out, const NewStopLimitPayload& p);
NewStopLimitPayload decode_new_stop_limit(const uint8_t* data, size_t len);

void encode_modify_limit(Buffer& out, const ModifyLimitPayload& p);
ModifyLimitPayload decode_modify_limit(const uint8_t* data, size_t len);

void encode_modify_stop(Buffer& out, const ModifyStopPayload& p);
ModifyStopPayload decode_modify_stop(const uint8_t* data, size_t len);

void encode_modify_stop_limit(Buffer& out, const ModifyStopLimitPayload& p);
ModifyStopLimitPayload decode_modify_stop_limit(const uint8_t* data, size_t len);

void encode_cancel_order(Buffer& out, const CancelOrderPayload& p);
CancelOrderPayload decode_cancel_order(const uint8_t* data, size_t len);

void encode_accepted(Buffer& out, const AcceptedPayload& p);
AcceptedPayload decode_accepted(const uint8_t* data, size_t len);

void encode_rejected(Buffer& out, const RejectedPayload& p);
RejectedPayload decode_rejected(const uint8_t* data, size_t len);

void encode_executed(Buffer& out, const ExecutedPayload& p);
ExecutedPayload decode_executed(const uint8_t* data, size_t len);

void encode_order_done(Buffer& out, const OrderDonePayload& p);
OrderDonePayload decode_order_done(const uint8_t* data, size_t len);

// Fixed payload size on the wire for a given request/response type.
size_t wire_size(MsgType type);

}  // namespace protocol

#endif
