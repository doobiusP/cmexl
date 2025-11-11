include_guard()
option(DOOBIUS_VERBOSE "Enable verbose output from configure for doobius targets" OFF)

option(DOOBIUS_USE_VCPKG "Use vcpkg as package manager?" OFF)
option(DOOBIUS_COPY_VCPKG_INCLUDE "Copy includes of vcpkg_installed to a pre-defined location?" OFF) #TODO: Allow specifying where
option(DOOBIUS_USE_CONAN "Use conan as package manager?" OFF)

option(DOOBIUS_BUILD_TESTS "Build tests for the doobius library?" OFF)
option(DOOBIUS_CHECK_PKG_MANAGER "Force check the correctness of package manager setup? Note: overrides building any other feature" OFF)

option(DOOBIUS_USE_ADDR2LINE "Use addr2line for Boost::stacktrace_addr2line?" OFF)
option(DOOBIUS_USE_BACKTRACE "Use libbacktrace for Boost::stacktrace_backtrace?" OFF)

option(BUILD_SHARED_LIBS "Build using shared libraries" OFF)

if(DOOBIUS_USE_CONAN AND DOOBIUS_USE_VCPKG)
    message(FATAL_ERROR "[DOOBIUS] Can't use both Conan2 and vcpkg")
endif()
if((NOT DOOBIUS_USE_CONAN) AND (NOT DOOBIUS_USE_VCPKG)) #TODO: Remove this once pkg-config is supported
    message(FATAL_ERROR "[DOOBIUS][TEMPORARY] Have to use either conan or vcpkg")
endif()