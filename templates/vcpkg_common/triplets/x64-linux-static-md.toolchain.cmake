set(CMAKE_C_COMPILER   gcc) #TODO: Make these find_programs()
set(CMAKE_CXX_COMPILER g++) #TODO: Make these find_programs()
include("$ENV{VCPKG_ROOT}/scripts/toolchains/linux.cmake")