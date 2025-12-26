set(CMAKE_C_FLAGS -m32)
set(CMAKE_CXX_FLAGS -m32)
set(CMAKE_EXE_LINKER_FLAGS -m32)
set(CMAKE_SHARED_LINKER_FLAGS -m32)
set(CMAKE_C_COMPILER   gcc)
set(CMAKE_CXX_COMPILER g++)

include("$ENV{VCPKG_ROOT}/scripts/toolchains/linux.cmake")
