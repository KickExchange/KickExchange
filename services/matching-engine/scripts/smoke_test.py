import socket
import struct

MSG_TYPE = {
    "Hello": 0, "HelloAck": 1, "Heartbeat": 2,
    "NewLimit": 3, "NewMarket": 4, "NewStop": 5, "NewStopLimit": 6,
    "ModifyLimit": 7, "ModifyStop": 8, "ModifyStopLimit": 9, "CancelOrder": 10,
    "Accepted": 11, "Rejected": 12, "Executed": 13, "OrderDone": 14,
}
MSG_NAME = {v: k for k, v in MSG_TYPE.items()}

HEADER_FMT = "<BBHIQQI"  # version, msg_type, flags, seq, request_id, asset_id, payload_size
HEADER_SIZE = struct.calcsize(HEADER_FMT)


def build_header(msg_type, request_id, asset_id, payload_size):
    return struct.pack(HEADER_FMT, 1, MSG_TYPE[msg_type], 0, 0, request_id, asset_id, payload_size)


def send_msg(sock, msg_type, request_id, asset_id, payload):
    sock.sendall(build_header(msg_type, request_id, asset_id, len(payload)) + payload)


def recv_exact(sock, n):
    buf = b""
    while len(buf) < n:
        chunk = sock.recv(n - len(buf))
        if not chunk:
            raise ConnectionError("connection closed")
        buf += chunk
    return buf


PAYLOAD_FMT = {
    "HelloAck": "<IBI",
    "Accepted": "<QQq",
    "Rejected": "<QB",
    "Executed": "<QQQqqqQ",
    "OrderDone": "<QBq",
}


def recv_msg(sock):
    header = recv_exact(sock, HEADER_SIZE)
    version, msg_type, flags, seq, request_id, asset_id, payload_size = struct.unpack(HEADER_FMT, header)
    payload = recv_exact(sock, payload_size) if payload_size else b""
    name = MSG_NAME[msg_type]
    fields = struct.unpack(PAYLOAD_FMT[name], payload) if name in PAYLOAD_FMT else ()
    return name, request_id, asset_id, fields


def main():
    sock = socket.create_connection(("localhost", 9000))

    send_msg(sock, "Hello", 1, 0, struct.pack("<II", 1, 1))
    print(recv_msg(sock))

    # resting sell: order 1, sell 10 @ 100
    send_msg(sock, "NewLimit", 2, 1, struct.pack("<QBqq", 1, 1, 10, 100))
    for _ in range(2):
        print(recv_msg(sock))

    # crossing buy: order 2, buy 10 @ 100
    send_msg(sock, "NewLimit", 3, 1, struct.pack("<QBqq", 2, 0, 10, 100))
    for _ in range(5):
        print(recv_msg(sock))

    sock.close()


if __name__ == "__main__":
    main()
