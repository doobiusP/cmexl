include_guard()
# Note that these options should be set on the CLI or in a preset

option(||.UName||_BUILD_COMP_math "Build math target?" OFF)
option(||.UName||_BUILD_COMP_core "Build core target?" OFF)

option(||.UName||_BUILD_TESTS "Build tests?" OFF)
option(||.UName||_BUILD_EXAMPLES "Build example executables using the library?" OFF)
option(||.UName||_INSTALL_DOCS "Install documentation?" OFF)

option(BUILD_SHARED_LIBS "Build shared libraries?" OFF)

option(||.UName||_USE_ADDR2LINE "Use addr2line for Boost::stacktrace_addr2line?" OFF)
option(||.UName||_USE_BACKTRACE "Use libbacktrace for Boost::stacktrace_backtrace?" OFF)