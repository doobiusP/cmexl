cmake_minimum_required(VERSION 3.18)

macro(_mmsg STR)
    message(VERBOSE "[||.UName||] ${STR}")
endmacro()

if(MSVC AND POLICY CMP0141)
    cmake_policy(SET CMP0141 NEW)
    set(CMAKE_MSVC_DEBUG_INFORMATION_FORMAT
        "$<$<CONFIG:Debug,RelWithDebInfo>:ProgramDatabase>"
        "$<$<CONFIG:Debug,RelWithDebInfo>:EditAndContinue>")
endif()

include("cmake/||.LName||_options.cmake")

file(TO_CMAKE_PATH "${CMAKE_TOOLCHAIN_FILE}" TOOLCHAIN_CMAKE_STYLE_PATH)
file(TO_CMAKE_PATH "${VCPKG_CHAINLOAD_TOOLCHAIN_FILE}" CHAINLOAD_CMAKE_STYLE_PATH)

_mmsg("Toolchain = ${TOOLCHAIN_CMAKE_STYLE_PATH}")
_mmsg("Triplet = ${VCPKG_TARGET_TRIPLET}")
_mmsg("Features = ${VCPKG_MANIFEST_FEATURES}")
_mmsg("Chainload Toolchain = ${CHAINLOAD_CMAKE_STYLE_PATH}")

# =============== Project Init =============== #
include("cmake/||.LName||_version.cmake")
project(||.UName|| VERSION "${||.UName||_VER}" LANGUAGES C CXX)
if(||.UName||_BUILD_TESTS)
    enable_testing()
endif()

if(CMAKE_SIZEOF_VOID_P EQUAL 8)
    set(||.UName||_64BIT TRUE)
elseif(CMAKE_SIZEOF_VOID_P EQUAL 4)
    set(||.UName||_32BIT TRUE)
else()
    message(FATAL_ERROR "[||.UName||][ERROR] Unable to determine sizeof(void*)")
endif()
message(VERBOSE "[||.UName||] sizeof(void*): ${CMAKE_SIZEOF_VOID_P}")

if(CMAKE_GENERATOR MATCHES "Visual Studio")
    message(VERBOSE "[||.UName||][Visual Studio] Platform Name: ${CMAKE_VS_PLATFORM_NAME}")
    message(VERBOSE "[||.UName||][Visual Studio] Platform Version: ${CMAKE_VS_WINDOWS_TARGET_PLATFORM_VERSION}")
    message(VERBOSE "[||.UName||][Visual Studio] Platform Toolset: ${CMAKE_VS_PLATFORM_TOOLSET}")
endif()

include(CMakePackageConfigHelpers)
include(GNUInstallDirs)
# =============== Config Check =============== #
get_property(USING_MULTI_CONFIG_GENERATOR GLOBAL PROPERTY GENERATOR_IS_MULTI_CONFIG)
if(NOT USING_MULTI_CONFIG_GENERATOR AND NOT CMAKE_BUILD_TYPE)
    message(WARNING "[||.UName||] CMAKE_BUILD_TYPE not explicitly set. Using Debug as default...")
    set(CMAKE_BUILD_TYPE Debug CACHE STRING "Build type" FORCE)
endif()

if(CMAKE_CONFIGURATION_TYPES)
    if(NOT "Profile" IN_LIST CMAKE_CONFIGURATION_TYPES)
        list(APPEND CMAKE_CONFIGURATION_TYPES Profile)
    endif()
endif()

set(ALLOWED_BUILD_TYPES Debug Release RelWithDebInfo MinSizeRel Profile)
if(NOT WIN32)
    set_property(CACHE CMAKE_BUILD_TYPE PROPERTY STRINGS "${ALLOWED_BUILD_TYPES}")
endif()
if(NOT USING_MULTI_CONFIG_GENERATOR AND NOT CMAKE_BUILD_TYPE IN_LIST ALLOWED_BUILD_TYPES)
	message(FATAL_ERROR
		"[||.UName||] Invalid build type: ${CMAKE_BUILD_TYPE}."
		"Allowed values are: ${ALLOWED_BUILD_TYPES}"
	)
endif()

string(PREPEND CMAKE_CXX_FLAGS_PROFILE "${CMAKE_CXX_FLAGS_RELWITHDEBINFO} ")
string(PREPEND CMAKE_C_FLAGS_PROFILE   "${CMAKE_C_FLAGS_RELWITHDEBINFO} ")
string(PREPEND CMAKE_EXE_LINKER_FLAGS_PROFILE    "${CMAKE_EXE_LINKER_FLAGS_RELWITHDEBINFO} ")
string(PREPEND CMAKE_SHARED_LINKER_FLAGS_PROFILE "${CMAKE_SHARED_LINKER_FLAGS_RELWITHDEBINFO} ")
string(PREPEND CMAKE_MODULE_LINKER_FLAGS_PROFILE "${CMAKE_MODULE_LINKER_FLAGS_RELWITHDEBINFO} ")
string(PREPEND CMAKE_STATIC_LINKER_FLAGS_PROFILE "${CMAKE_STATIC_LINKER_FLAGS_RELWITHDEBINFO} ")

include("cmake/||.LName||_defaults.cmake")

function(_||.LName||_find_addr2line OUTVAR)
    if(||.UName||_USING_MINGW)
        set(_names x86_64-w64-mingw32-addr2line)
        if(||.UName||_32BIT)
            list(APPEND _names i686-w64-mingw32-addr2line)
        endif()
        list(APPEND _names addr2line)
    else()
        set(_names addr2line)
    endif()

    find_program(_bin NAMES ${_names})
    if(NOT _bin)
        message(FATAL_ERROR "[||.UName||] Cannot find addr2line. Re-run CMake with --debug-find to diagnose.")
    endif()
    message(STATUS "[||.UName||] Using addr2line: ${_bin}")

    set(${OUTVAR} "${_bin}" PARENT_SCOPE)
endfunction()

set(||.UName||_BOOST_STACKTRACE_IMPORT stacktrace_basic)
set(||.UName||_NEED_LIBDL OFF)
set(||.UName||_NEED_DBGENG OFF)

if(WIN32)
    set(||.UName||_BOOST_STACKTRACE_IMPORT stacktrace_windbg)
    set(||.UName||_NEED_DBGENG ON)
    if(||.UName||_USE_BACKTRACE OR ||.UName||_USE_ADDR2LINE)
        message(WARNING "[||.UName||] Forcefully disabling backtrace and/or addr2line on windows")
    endif()
    set(||.UName||_USE_BACKTRACE OFF)
    set(||.UName||_USE_ADDR2LINE OFF)
else()
    set(||.UName||_NEED_LIBDL ON)

    if(||.UName||_USE_BACKTRACE)
        set(||.UName||_BOOST_STACKTRACE_IMPORT stacktrace_backtrace)
    elseif(||.UName||_USE_ADDR2LINE)
        _||.LName||_find_addr2line(||.UName||_ADDR2LINE_BINARY)
        set(||.UName||_BOOST_STACKTRACE_IMPORT stacktrace_addr2line)
    endif()
endif()


_mmsg("COMPILER CXX All Flags: ${CMAKE_CXX_FLAGS}")
_mmsg("COMPILER CXX D Flags: ${CMAKE_CXX_FLAGS_DEBUG}")
_mmsg("COMPILER CXX R Flags: ${CMAKE_CXX_FLAGS_RELEASE}")
_mmsg("COMPILER CXX RWDI Flags: ${CMAKE_CXX_FLAGS_RELWITHDEBINFO}")
_mmsg("COMPILER CXX MSR Flags: ${CMAKE_CXX_FLAGS_MINSIZEREL}")
_mmsg("COMPILER CXX PROFILE Flags: ${CMAKE_CXX_FLAGS_PROFILE}")

_mmsg("COMPILER C All Flags: ${CMAKE_C_FLAGS}")
_mmsg("COMPILER C D Flags: ${CMAKE_C_FLAGS_DEBUG}")
_mmsg("COMPILER C R Flags: ${CMAKE_C_FLAGS_RELEASE}")
_mmsg("COMPILER C RWDI Flags: ${CMAKE_C_FLAGS_RELWITHDEBINFO}")
_mmsg("COMPILER C MSR Flags: ${CMAKE_C_FLAGS_MINSIZEREL}")
_mmsg("COMPILER C PROFILE Flags: ${CMAKE_C_FLAGS_PROFILE}")

_mmsg("EXE Linker Flags: ${CMAKE_EXE_LINKER_FLAGS}")
_mmsg("EXE Linker D Flags: ${CMAKE_EXE_LINKER_FLAGS_DEBUG}")
_mmsg("EXE Linker R Flags: ${CMAKE_EXE_LINKER_FLAGS_RELEASE}")
_mmsg("EXE Linker RWDI Flags: ${CMAKE_EXE_LINKER_FLAGS_RELWITHDEBINFO}")
_mmsg("EXE Linker MSR Flags: ${CMAKE_EXE_LINKER_FLAGS_MINSIZEREL}")
_mmsg("EXE Linker PROFILE Flags: ${CMAKE_EXE_LINKER_FLAGS_PROFILE}")

_mmsg("SHARED Linker Flags: ${CMAKE_SHARED_LINKER_FLAGS}")
_mmsg("SHARED Linker D Flags: ${CMAKE_SHARED_LINKER_FLAGS_DEBUG}")
_mmsg("SHARED Linker R Flags: ${CMAKE_SHARED_LINKER_FLAGS_RELEASE}")
_mmsg("SHARED Linker RWDI Flags: ${CMAKE_SHARED_LINKER_FLAGS_RELWITHDEBINFO}")
_mmsg("SHARED Linker MSR Flags: ${CMAKE_SHARED_LINKER_FLAGS_MINSIZEREL}")
_mmsg("SHARED Linker PROFILE Flags: ${CMAKE_SHARED_LINKER_FLAGS_PROFILE}")

_mmsg("MODULE Linker Flags: ${CMAKE_MODULE_LINKER_FLAGS}")
_mmsg("MODULE Linker D Flags: ${CMAKE_MODULE_LINKER_FLAGS_DEBUG}")
_mmsg("MODULE Linker R Flags: ${CMAKE_MODULE_LINKER_FLAGS_RELEASE}")
_mmsg("MODULE Linker RWDI Flags: ${CMAKE_MODULE_LINKER_FLAGS_RELWITHDEBINFO}")
_mmsg("MODULE Linker MSR Flags: ${CMAKE_MODULE_LINKER_FLAGS_MINSIZEREL}")
_mmsg("MODULE Linker PROFILE Flags: ${CMAKE_MODULE_LINKER_FLAGS_PROFILE}")

_mmsg("STATIC Linker Flags: ${CMAKE_STATIC_LINKER_FLAGS}")
add_subdirectory(src)
if(CMAKE_EXPORT_COMPILE_COMMANDS AND CMAKE_SOURCE_DIR STREQUAL CMAKE_CURRENT_SOURCE_DIR)
    set(COMPILE_COMMANDS_SRC "${CMAKE_BINARY_DIR}/compile_commands.json")
    set(COMPILE_COMMANDS_DST "${CMAKE_SOURCE_DIR}/compile_commands.json")

    add_custom_target(copy_compile_commands ALL
        COMMAND ${CMAKE_COMMAND} -E copy_if_different
        ${COMPILE_COMMANDS_SRC}
        ${COMPILE_COMMANDS_DST}
        COMMENT "Copying compile_commands.json to source root"
        VERBATIM
    )
endif()

if(CMAKE_SOURCE_DIR STREQUAL CMAKE_CURRENT_SOURCE_DIR)
    add_subdirectory(packaging)
endif()