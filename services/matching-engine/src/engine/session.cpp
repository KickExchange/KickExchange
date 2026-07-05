#include "session.hpp"
#include "../limit_order_book/Order.hpp"

#include <chrono>

namespace engine {

using namespace protocol;

namespace {
bool is_buy(Side side) { return side == Side::Buy; }
}  // namespace

Session::Session(AssetBookManager& books) : books_(books) {}

uint64_t Session::now_millis() const {
    auto now = std::chrono::system_clock::now();
    return static_cast<uint64_t>(
        std::chrono::duration_cast<std::chrono::milliseconds>(now.time_since_epoch()).count());
}

OutgoingMessage Session::make(const Header& req, MsgType type, uint64_t request_id, Buffer payload) {
    Header h;
    h.version = kProtocolVersion;
    h.msg_type = type;
    h.flags = 0;
    h.sequence_number = next_sequence_number_++;
    h.request_id = request_id;
    h.asset_id = req.asset_id;
    h.payload_size = static_cast<uint32_t>(payload.size());
    return OutgoingMessage{h, std::move(payload)};
}

OutgoingMessage Session::make_rejected(const Header& req, uint64_t order_id, RejectReason reason) {
    RejectedPayload p{order_id, reason};
    Buffer buf;
    encode_rejected(buf, p);
    return make(req, MsgType::Rejected, req.request_id, std::move(buf));
}

template <typename PlaceFn>
std::vector<OutgoingMessage> Session::place_order(const Header& req, uint64_t order_id, int64_t requested_shares,
                                                   bool remainder_can_rest, OrderKind resting_kind, PlaceFn place) {
    std::vector<OutgoingMessage> out;
    Book& book = books_.get_or_create(req.asset_id);

    std::vector<TradeEvent> trades;
    book.onTrade = [&](const TradeEvent& t) { trades.push_back(t); };
    place(book);
    book.onTrade = nullptr;

    uint64_t ts = now_millis();

    AcceptedPayload accepted{order_id, ts, requested_shares};
    Buffer acceptedBuf;
    encode_accepted(acceptedBuf, accepted);
    out.push_back(make(req, MsgType::Accepted, req.request_id, std::move(acceptedBuf)));

    int64_t remaining = requested_shares;
    for (const auto& t : trades) {
        uint64_t trade_id = next_trade_id_++;
        remaining -= t.shares;

        ExecutedPayload incoming{trade_id, order_id, static_cast<uint64_t>(t.restingOrderId),
                                  t.price, t.shares, remaining, ts};
        Buffer incomingBuf;
        encode_executed(incomingBuf, incoming);
        out.push_back(make(req, MsgType::Executed, req.request_id, std::move(incomingBuf)));

        ExecutedPayload resting{trade_id, static_cast<uint64_t>(t.restingOrderId), order_id,
                                 t.price, t.shares, t.restingRemainingShares, ts};
        Buffer restingBuf;
        encode_executed(restingBuf, resting);
        out.push_back(make(req, MsgType::Executed, 0, std::move(restingBuf)));

        if (t.restingRemainingShares == 0) {
            OrderDonePayload restingDone{static_cast<uint64_t>(t.restingOrderId), FinalStatus::Filled, 0};
            Buffer doneBuf;
            encode_order_done(doneBuf, restingDone);
            out.push_back(make(req, MsgType::OrderDone, 0, std::move(doneBuf)));
            order_kind_by_id_.erase(static_cast<uint64_t>(t.restingOrderId));
        }
    }

    FinalStatus status;
    if (remaining == 0) {
        status = FinalStatus::Filled;
        order_kind_by_id_.erase(order_id);
    } else if (remainder_can_rest) {
        status = trades.empty() ? FinalStatus::Resting : FinalStatus::PartiallyFilledResting;
        order_kind_by_id_[order_id] = resting_kind;
    } else {
        status = FinalStatus::PartiallyFilledDropped;
        order_kind_by_id_.erase(order_id);
    }

    OrderDonePayload done{order_id, status, remaining};
    Buffer doneBuf;
    encode_order_done(doneBuf, done);
    out.push_back(make(req, MsgType::OrderDone, req.request_id, std::move(doneBuf)));

    return out;
}

std::vector<OutgoingMessage> Session::handle_new_limit(const Header& req, const NewLimitPayload& p) {
    if (req.asset_id == 0) return {make_rejected(req, p.order_id, RejectReason::InvalidAsset)};
    return place_order(req, p.order_id, p.shares, true, OrderKind::Limit, [&](Book& b) {
        b.addLimitOrder(static_cast<int>(p.order_id), is_buy(p.side), static_cast<int>(p.shares),
                         static_cast<int>(p.price_ticks));
    });
}

std::vector<OutgoingMessage> Session::handle_new_market(const Header& req, const NewMarketPayload& p) {
    if (req.asset_id == 0) return {make_rejected(req, p.order_id, RejectReason::InvalidAsset)};
    return place_order(req, p.order_id, p.shares, false, OrderKind::Limit, [&](Book& b) {
        b.marketOrder(static_cast<int>(p.order_id), is_buy(p.side), static_cast<int>(p.shares));
    });
}

std::vector<OutgoingMessage> Session::handle_new_stop(const Header& req, const NewStopPayload& p) {
    if (req.asset_id == 0) return {make_rejected(req, p.order_id, RejectReason::InvalidAsset)};
    return place_order(req, p.order_id, p.shares, true, OrderKind::Stop, [&](Book& b) {
        b.addStopOrder(static_cast<int>(p.order_id), is_buy(p.side), static_cast<int>(p.shares),
                        static_cast<int>(p.stop_price_ticks));
    });
}

std::vector<OutgoingMessage> Session::handle_new_stop_limit(const Header& req, const NewStopLimitPayload& p) {
    if (req.asset_id == 0) return {make_rejected(req, p.order_id, RejectReason::InvalidAsset)};
    return place_order(req, p.order_id, p.shares, true, OrderKind::StopLimit, [&](Book& b) {
        b.addStopLimitOrder(static_cast<int>(p.order_id), is_buy(p.side), static_cast<int>(p.shares),
                             static_cast<int>(p.limit_price_ticks), static_cast<int>(p.stop_price_ticks));
    });
}

std::vector<OutgoingMessage> Session::handle_modify_limit(const Header& req, const ModifyLimitPayload& p) {
    if (req.asset_id == 0) return {make_rejected(req, p.order_id, RejectReason::InvalidAsset)};
    auto it = order_kind_by_id_.find(p.order_id);
    if (it == order_kind_by_id_.end() || it->second != OrderKind::Limit) {
        return {make_rejected(req, p.order_id, RejectReason::UnknownOrderId)};
    }
    Book& book = books_.get_or_create(req.asset_id);
    book.modifyLimitOrder(static_cast<int>(p.order_id), static_cast<int>(p.new_shares),
                           static_cast<int>(p.new_price_ticks));
    AcceptedPayload accepted{p.order_id, now_millis(), p.new_shares};
    Buffer buf;
    encode_accepted(buf, accepted);
    return {make(req, MsgType::Accepted, req.request_id, std::move(buf))};
}

std::vector<OutgoingMessage> Session::handle_modify_stop(const Header& req, const ModifyStopPayload& p) {
    if (req.asset_id == 0) return {make_rejected(req, p.order_id, RejectReason::InvalidAsset)};
    auto it = order_kind_by_id_.find(p.order_id);
    if (it == order_kind_by_id_.end() || it->second != OrderKind::Stop) {
        return {make_rejected(req, p.order_id, RejectReason::UnknownOrderId)};
    }
    Book& book = books_.get_or_create(req.asset_id);
    book.modifyStopOrder(static_cast<int>(p.order_id), static_cast<int>(p.new_shares),
                          static_cast<int>(p.new_stop_price_ticks));
    AcceptedPayload accepted{p.order_id, now_millis(), p.new_shares};
    Buffer buf;
    encode_accepted(buf, accepted);
    return {make(req, MsgType::Accepted, req.request_id, std::move(buf))};
}

std::vector<OutgoingMessage> Session::handle_modify_stop_limit(const Header& req, const ModifyStopLimitPayload& p) {
    if (req.asset_id == 0) return {make_rejected(req, p.order_id, RejectReason::InvalidAsset)};
    auto it = order_kind_by_id_.find(p.order_id);
    if (it == order_kind_by_id_.end() || it->second != OrderKind::StopLimit) {
        return {make_rejected(req, p.order_id, RejectReason::UnknownOrderId)};
    }
    Book& book = books_.get_or_create(req.asset_id);
    book.modifyStopLimitOrder(static_cast<int>(p.order_id), static_cast<int>(p.new_shares),
                               static_cast<int>(p.new_limit_price_ticks), static_cast<int>(p.new_stop_price_ticks));
    AcceptedPayload accepted{p.order_id, now_millis(), p.new_shares};
    Buffer buf;
    encode_accepted(buf, accepted);
    return {make(req, MsgType::Accepted, req.request_id, std::move(buf))};
}

std::vector<OutgoingMessage> Session::handle_cancel_order(const Header& req, const CancelOrderPayload& p) {
    if (req.asset_id == 0) return {make_rejected(req, p.order_id, RejectReason::InvalidAsset)};
    auto it = order_kind_by_id_.find(p.order_id);
    if (it == order_kind_by_id_.end()) {
        return {make_rejected(req, p.order_id, RejectReason::UnknownOrderId)};
    }

    Book& book = books_.get_or_create(req.asset_id);
    int oid = static_cast<int>(p.order_id);
    Order* existing = book.searchOrderMap(oid);
    int64_t remaining = existing != nullptr ? existing->getShares() : 0;

    switch (it->second) {
        case OrderKind::Limit: book.cancelLimitOrder(oid); break;
        case OrderKind::Stop: book.cancelStopOrder(oid); break;
        case OrderKind::StopLimit: book.cancelStopLimitOrder(oid); break;
    }
    order_kind_by_id_.erase(it);

    OrderDonePayload done{p.order_id, FinalStatus::Cancelled, remaining};
    Buffer buf;
    encode_order_done(buf, done);
    return {make(req, MsgType::OrderDone, req.request_id, std::move(buf))};
}

}  // namespace engine
