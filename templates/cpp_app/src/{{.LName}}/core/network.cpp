#include "||.LName||/core/network.h"
#include <cstdio>
#include <errno.h>
#include <stdlib.h>

void die(const char* log) {
    int err = errno;
    fprintf(stderr, "[%d] %s\n", err, log);
    abort();
}