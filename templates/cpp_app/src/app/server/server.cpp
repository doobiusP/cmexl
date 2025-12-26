#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#if defined(_WIN32)
    #include <winsock2.h>
    #include <ws2tcpip.h>
    using socket_t = SOCKET;
    using socket_len_t = int;
#else
    #include <errno.h>
    #include <unistd.h>
    #include <arpa/inet.h>
    #include <sys/socket.h>
    #include <netinet/ip.h>
    using socket_t = int;
    using socket_len_t = socklen_t;
    #define INVALID_SOCKET (-1)
#endif

static void msg(const char *m) {
    fprintf(stderr, "%s\n", m);
}

static void die(const char *m) {
#if defined(_WIN32)
    int err = WSAGetLastError();
#else
    int err = errno;
#endif
    fprintf(stderr, "[%d] %s\n", err, m);
    abort();
}

// Changed parameter to socket_t to prevent truncation warnings
static void do_something(socket_t connfd) {
    char rbuf[64] = {};

#if defined(_WIN32)
    int n = recv(connfd, rbuf, sizeof(rbuf) - 1, 0);
#else
    ssize_t n = read(connfd, rbuf, sizeof(rbuf) - 1);
#endif

    if (n < 0) {
        msg("read() error");
        return;
    }

    fprintf(stderr, "client says: %s\n", rbuf);

    const char wbuf[] = "world";

#if defined(_WIN32)
    send(connfd, wbuf, (int)strlen(wbuf), 0);
#else
    write(connfd, wbuf, strlen(wbuf));
#endif
}

int main() {
#if defined(_WIN32)
    WSADATA wsa;
    if (WSAStartup(MAKEWORD(2, 2), &wsa) != 0) {
        die("WSAStartup");
    }
#endif

    // Use socket_t and check against INVALID_SOCKET
    socket_t fd = socket(AF_INET, SOCK_STREAM, 0);
    if (fd == INVALID_SOCKET) {
        die("socket()");
    }

    int val = 1;
    setsockopt(fd, SOL_SOCKET, SO_REUSEADDR,
               (const char *)&val, sizeof(val));

    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(1234);
    addr.sin_addr.s_addr = htonl(INADDR_ANY);

    int rv = bind(fd, (const sockaddr *)&addr, sizeof(addr));
    if (rv != 0) {
        die("bind()");
    }

    rv = listen(fd, SOMAXCONN);
    if (rv != 0) {
        die("listen()");
    }

    while (true) {
        sockaddr_in client_addr{};
        socket_len_t addrlen = sizeof(client_addr);

        // Accept returns a socket_t
        socket_t connfd = accept(fd,
                                (sockaddr *)&client_addr,
                                &addrlen);
        
        if (connfd == INVALID_SOCKET) {
            continue;
        }

        do_something(connfd);

#if defined(_WIN32)
        closesocket(connfd);
#else
        close(connfd);
#endif
    }

#if defined(_WIN32)
    closesocket(fd);
    WSACleanup();
#else
    close(fd);
#endif

    return 0;
}