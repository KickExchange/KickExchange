// Package engineclient ports the Matching Engine's Protocol v1 wire format.
// Canonical source of truth: services/matching-engine/src/protocol/{message_types.hpp,codec.cpp}.
package engineclient

import (
	"encoding/binary"
	"fmt"
)

const (
	ProtocolVersion = 1
	HeaderSize      = 28
)

type MsgType uint8

const (
	MsgHello           MsgType = 0
	MsgHelloAck        MsgType = 1
	MsgHeartbeat       MsgType = 2
	MsgNewLimit        MsgType = 3
	MsgNewMarket       MsgType = 4
	MsgNewStop         MsgType = 5
	MsgNewStopLimit    MsgType = 6
	MsgModifyLimit     MsgType = 7
	MsgModifyStop      MsgType = 8
	MsgModifyStopLimit MsgType = 9
	MsgCancelOrder     MsgType = 10
	MsgAccepted        MsgType = 11
	MsgRejected        MsgType = 12
	MsgExecuted        MsgType = 13
	MsgOrderDone       MsgType = 14
)

type Side uint8

const (
	SideBuy  Side = 0
	SideSell Side = 1
)

type RejectReason uint8

const (
	RejectInvalidAsset       RejectReason = 0
	RejectUnknownOrderId     RejectReason = 1
	RejectThrottled          RejectReason = 2
	RejectInsufficientShares RejectReason = 3
	RejectMalformed          RejectReason = 4
	RejectVersionMismatch    RejectReason = 5
	RejectDuplicateRequest   RejectReason = 6
	RejectInvalidPrice       RejectReason = 7
	RejectInvalidQuantity    RejectReason = 8
	RejectEngineOverloaded   RejectReason = 9
)

type FinalStatus uint8

const (
	StatusFilled                 FinalStatus = 0
	StatusPartiallyFilledResting FinalStatus = 1
	StatusResting                FinalStatus = 2
	StatusCancelled              FinalStatus = 3
	StatusPartiallyFilledDropped FinalStatus = 4
)

// wireSize returns the fixed payload size for a given message type. Mirrors codec.cpp's wire_size.
func wireSize(t MsgType) (int, error) {
	switch t {
	case MsgHello:
		return 8, nil
	case MsgHelloAck:
		return 9, nil
	case MsgHeartbeat:
		return 0, nil
	case MsgNewLimit:
		return 25, nil
	case MsgNewMarket:
		return 17, nil
	case MsgNewStop:
		return 25, nil
	case MsgNewStopLimit:
		return 33, nil
	case MsgModifyLimit:
		return 24, nil
	case MsgModifyStop:
		return 24, nil
	case MsgModifyStopLimit:
		return 32, nil
	case MsgCancelOrder:
		return 8, nil
	case MsgAccepted:
		return 24, nil
	case MsgRejected:
		return 9, nil
	case MsgExecuted:
		return 56, nil
	case MsgOrderDone:
		return 17, nil
	}
	return 0, fmt.Errorf("engineclient: unknown msg_type %d", t)
}

type Header struct {
	Version        uint8
	MsgType        MsgType
	Flags          uint16
	SequenceNumber uint32
	RequestID      uint64
	AssetID        uint64
	PayloadSize    uint32
}

func encodeHeader(h Header) []byte {
	buf := make([]byte, HeaderSize)
	buf[0] = h.Version
	buf[1] = uint8(h.MsgType)
	binary.LittleEndian.PutUint16(buf[2:4], h.Flags)
	binary.LittleEndian.PutUint32(buf[4:8], h.SequenceNumber)
	binary.LittleEndian.PutUint64(buf[8:16], h.RequestID)
	binary.LittleEndian.PutUint64(buf[16:24], h.AssetID)
	binary.LittleEndian.PutUint32(buf[24:28], h.PayloadSize)
	return buf
}

func decodeHeader(buf []byte) Header {
	return Header{
		Version:        buf[0],
		MsgType:        MsgType(buf[1]),
		Flags:          binary.LittleEndian.Uint16(buf[2:4]),
		SequenceNumber: binary.LittleEndian.Uint32(buf[4:8]),
		RequestID:      binary.LittleEndian.Uint64(buf[8:16]),
		AssetID:        binary.LittleEndian.Uint64(buf[16:24]),
		PayloadSize:    binary.LittleEndian.Uint32(buf[24:28]),
	}
}

func encodeMessage(h Header, payload []byte) []byte {
	h.PayloadSize = uint32(len(payload))
	out := encodeHeader(h)
	return append(out, payload...)
}

type HelloPayload struct {
	MinVersion uint32
	MaxVersion uint32
}

func encodeHello(p HelloPayload) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], p.MinVersion)
	binary.LittleEndian.PutUint32(buf[4:8], p.MaxVersion)
	return buf
}

type HelloAckPayload struct {
	AcceptedVersion uint32
	Ok              uint8
	EngineVersion   uint32
}

func decodeHelloAck(buf []byte) HelloAckPayload {
	return HelloAckPayload{
		AcceptedVersion: binary.LittleEndian.Uint32(buf[0:4]),
		Ok:              buf[4],
		EngineVersion:   binary.LittleEndian.Uint32(buf[5:9]),
	}
}

type NewLimitPayload struct {
	OrderID    uint64
	Side       Side
	Shares     int64
	PriceTicks int64
}

func encodeNewLimit(p NewLimitPayload) []byte {
	buf := make([]byte, 25)
	binary.LittleEndian.PutUint64(buf[0:8], p.OrderID)
	buf[8] = uint8(p.Side)
	binary.LittleEndian.PutUint64(buf[9:17], uint64(p.Shares))
	binary.LittleEndian.PutUint64(buf[17:25], uint64(p.PriceTicks))
	return buf
}

type NewMarketPayload struct {
	OrderID uint64
	Side    Side
	Shares  int64
}

func encodeNewMarket(p NewMarketPayload) []byte {
	buf := make([]byte, 17)
	binary.LittleEndian.PutUint64(buf[0:8], p.OrderID)
	buf[8] = uint8(p.Side)
	binary.LittleEndian.PutUint64(buf[9:17], uint64(p.Shares))
	return buf
}

type NewStopPayload struct {
	OrderID        uint64
	Side           Side
	Shares         int64
	StopPriceTicks int64
}

func encodeNewStop(p NewStopPayload) []byte {
	buf := make([]byte, 25)
	binary.LittleEndian.PutUint64(buf[0:8], p.OrderID)
	buf[8] = uint8(p.Side)
	binary.LittleEndian.PutUint64(buf[9:17], uint64(p.Shares))
	binary.LittleEndian.PutUint64(buf[17:25], uint64(p.StopPriceTicks))
	return buf
}

type NewStopLimitPayload struct {
	OrderID         uint64
	Side            Side
	Shares          int64
	LimitPriceTicks int64
	StopPriceTicks  int64
}

func encodeNewStopLimit(p NewStopLimitPayload) []byte {
	buf := make([]byte, 33)
	binary.LittleEndian.PutUint64(buf[0:8], p.OrderID)
	buf[8] = uint8(p.Side)
	binary.LittleEndian.PutUint64(buf[9:17], uint64(p.Shares))
	binary.LittleEndian.PutUint64(buf[17:25], uint64(p.LimitPriceTicks))
	binary.LittleEndian.PutUint64(buf[25:33], uint64(p.StopPriceTicks))
	return buf
}

type ModifyLimitPayload struct {
	OrderID       uint64
	NewShares     int64
	NewPriceTicks int64
}

func encodeModifyLimit(p ModifyLimitPayload) []byte {
	buf := make([]byte, 24)
	binary.LittleEndian.PutUint64(buf[0:8], p.OrderID)
	binary.LittleEndian.PutUint64(buf[8:16], uint64(p.NewShares))
	binary.LittleEndian.PutUint64(buf[16:24], uint64(p.NewPriceTicks))
	return buf
}

type ModifyStopPayload struct {
	OrderID           uint64
	NewShares         int64
	NewStopPriceTicks int64
}

func encodeModifyStop(p ModifyStopPayload) []byte {
	buf := make([]byte, 24)
	binary.LittleEndian.PutUint64(buf[0:8], p.OrderID)
	binary.LittleEndian.PutUint64(buf[8:16], uint64(p.NewShares))
	binary.LittleEndian.PutUint64(buf[16:24], uint64(p.NewStopPriceTicks))
	return buf
}

type ModifyStopLimitPayload struct {
	OrderID            uint64
	NewShares          int64
	NewLimitPriceTicks int64
	NewStopPriceTicks  int64
}

func encodeModifyStopLimit(p ModifyStopLimitPayload) []byte {
	buf := make([]byte, 32)
	binary.LittleEndian.PutUint64(buf[0:8], p.OrderID)
	binary.LittleEndian.PutUint64(buf[8:16], uint64(p.NewShares))
	binary.LittleEndian.PutUint64(buf[16:24], uint64(p.NewLimitPriceTicks))
	binary.LittleEndian.PutUint64(buf[24:32], uint64(p.NewStopPriceTicks))
	return buf
}

type CancelOrderPayload struct {
	OrderID uint64
}

func encodeCancelOrder(p CancelOrderPayload) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf[0:8], p.OrderID)
	return buf
}

type AcceptedPayload struct {
	OrderID      uint64
	Timestamp    uint64
	RemainingQty int64
}

func decodeAccepted(buf []byte) AcceptedPayload {
	return AcceptedPayload{
		OrderID:      binary.LittleEndian.Uint64(buf[0:8]),
		Timestamp:    binary.LittleEndian.Uint64(buf[8:16]),
		RemainingQty: int64(binary.LittleEndian.Uint64(buf[16:24])),
	}
}

type RejectedPayload struct {
	OrderID uint64
	Reason  RejectReason
}

func decodeRejected(buf []byte) RejectedPayload {
	return RejectedPayload{
		OrderID: binary.LittleEndian.Uint64(buf[0:8]),
		Reason:  RejectReason(buf[8]),
	}
}

type ExecutedPayload struct {
	TradeID        uint64
	OrderID        uint64
	MatchedOrderID uint64
	PriceTicks     int64
	Quantity       int64
	RemainingQty   int64
	Timestamp      uint64
}

func decodeExecuted(buf []byte) ExecutedPayload {
	return ExecutedPayload{
		TradeID:        binary.LittleEndian.Uint64(buf[0:8]),
		OrderID:        binary.LittleEndian.Uint64(buf[8:16]),
		MatchedOrderID: binary.LittleEndian.Uint64(buf[16:24]),
		PriceTicks:     int64(binary.LittleEndian.Uint64(buf[24:32])),
		Quantity:       int64(binary.LittleEndian.Uint64(buf[32:40])),
		RemainingQty:   int64(binary.LittleEndian.Uint64(buf[40:48])),
		Timestamp:      binary.LittleEndian.Uint64(buf[48:56]),
	}
}

type OrderDonePayload struct {
	OrderID      uint64
	FinalStatus  FinalStatus
	RemainingQty int64
}

func decodeOrderDone(buf []byte) OrderDonePayload {
	return OrderDonePayload{
		OrderID:      binary.LittleEndian.Uint64(buf[0:8]),
		FinalStatus:  FinalStatus(buf[8]),
		RemainingQty: int64(binary.LittleEndian.Uint64(buf[9:17])),
	}
}
