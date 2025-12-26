#include <stdint.h>
#include <string.h>
#include <stdio.h>
#include <||.LName||/core/network.h>

#if defined(_WIN32)
    #include <winsock2.h>
    #include <ws2tcpip.h>
    using socket_len_t = int;
    using socket_t = SOCKET;  // Windows uses SOCKET (UINT_PTR)
#else
    #include <unistd.h>
    #include <arpa/inet.h>
    #include <sys/socket.h>
    #include <netinet/ip.h>
    using socket_len_t = socklen_t;
    using socket_t = int;     // POSIX uses int
    #define INVALID_SOCKET (-1)
#endif

int main() {
#if defined(_WIN32)
    WSADATA wsa;
    if (WSAStartup(MAKEWORD(2, 2), &wsa) != 0) {
        die("WSAStartup");
    }
#endif

    // 1. Changed 'int' to 'socket_t'
    // 2. Changed 'fd < 0' to 'fd == INVALID_SOCKET'
    socket_t fd = socket(AF_INET, SOCK_STREAM, 0);
    if (fd == INVALID_SOCKET) {
        die("socket()");
    }

    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(1234);
    addr.sin_addr.s_addr = htonl(INADDR_LOOPBACK);

    // Windows connect() takes 'int' for the length parameter
    int rv = connect(fd, reinterpret_cast<const sockaddr*>(&addr),
                      sizeof(addr));
    if (rv != 0) {
        die("connect");
    }

    const char msg[] = "hello";

#if defined(_WIN32)
    send(fd, msg, (int)strlen(msg), 0);
#else
    write(fd, msg, strlen(msg));
#endif

    char rbuf[64]{};

#if defined(_WIN32)
    int n = recv(fd, rbuf, sizeof(rbuf) - 1, 0);
#else
    ssize_t n = read(fd, rbuf, sizeof(rbuf) - 1);
#endif

    if (n < 0) {
        die("read");
    }

    printf("server says: %s\n", rbuf);

#if defined(_WIN32)
    closesocket(fd);
    WSACleanup();
#else
    close(fd);
#endif

    return 0;
}