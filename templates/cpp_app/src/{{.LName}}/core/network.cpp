#include "||.LName||/core/network.h"
#include <cmath>
#include <cstdio>
#include <errno.h>
#include <iostream>
#include <stdlib.h>
#include <vector>


void die(const char *log) {
  int err = errno;
  fprintf(stderr, "[%d] %s\n", err, log);
  abort();
}

void prof_test() {
  const int size = 1000;
  std::cout << "Starting heavy computation (" << size << "x" << size
            << " matrix)...\n";

  // Use vectors to avoid stack overflow for large sizes
  std::vector<std::vector<double>> A(size, std::vector<double>(size, 1.1));
  std::vector<std::vector<double>> B(size, std::vector<double>(size, 2.2));
  std::vector<std::vector<double>> C(size, std::vector<double>(size, 0.0));

  // O(n^3) Matrix Multiplication
  for (int i = 0; i < size; ++i) {
    for (int j = 0; j < size; ++j) {
      for (int k = 0; k < size; ++k) {
        C[i][j] += A[i][k] * B[k][j];
      }
    }
    if (i % 100 == 0)
      std::cout << "Progress: " << (i / 10) << "%" << '\n';
  }

  // A secondary heavy math task: Sine/Square root wave sum
  volatile double trash_collector = 0;
  for (int i = 0; i < 10000000; ++i) {
    trash_collector += std::sin(i) * std::sqrt(i);
  }

  std::cout << "Computation complete. Final corner value: "
            << C[size - 1][size - 1] << '\n';
  std::cout << "Trash collector: " << trash_collector << '\n';
}