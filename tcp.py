import socket
import threading
import time
import sys


def server(port):
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.bind(('localhost', port))
    server_socket.listen(1)
    print(f"Server is listening on localhost:{port}")
    sys.stdout.flush()

    conn, addr = server_socket.accept()
    print(f"Connected by {addr}")
    sys.stdout.flush()

    messages_received = 0
    start_time = time.time()

    while True:
        data = conn.recv(1024)
        if not data:
            break
        print(f"Server received: {data.decode()}")
        sys.stdout.flush()
        conn.sendall(f"Server received: {data.decode()}".encode())
        messages_received += 1

    end_time = time.time()
    print(f"Server received {messages_received} messages in {
          end_time - start_time:.2f} seconds")
    sys.stdout.flush()
    conn.close()


def client(port):
    time.sleep(1)
    client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    client_socket.connect(('localhost', port))
    print(f"Client connected to localhost:{port}")
    sys.stdout.flush()

    messages = ["Hello", "How are you?", "Goodbye"]
    for message in messages:
        client_socket.sendall(message.encode())
        data = client_socket.recv(1024)
        print(f"Client received: {data.decode()}")
        sys.stdout.flush()
        time.sleep(1)

    client_socket.close()


def main():
    if len(sys.argv) != 2:
        print("Usage: python tcp.py <port>")
        sys.exit(1)

    port = int(sys.argv[1])

    server_thread = threading.Thread(target=server, args=(port,))
    client_thread = threading.Thread(target=client, args=(port,))

    server_thread.start()
    client_thread.start()

    server_thread.join()
    client_thread.join()


if __name__ == "__main__":
    main()
