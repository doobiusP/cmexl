set(VCPKG_CMAKE_SYSTEM_NAME Linux)
set(VCPKG_TARGET_ARCHITECTURE x86)
set(VCPKG_CRT_LINKAGE dynamic)
set(VCPKG_LIBRARY_LINKAGE static)
set(VCPKG_FIXUP_ELF_RPATH ON)
# x64 toolchain also works for x86 since it just sets the compiler
set(VCPKG_CHAINLOAD_TOOLCHAIN_FILE "${PROJECT_SOURCE_DIR}/toolchains/x86-linux-static-md.toolchain.cmake")