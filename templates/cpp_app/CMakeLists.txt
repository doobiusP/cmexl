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
project(||.Name|| VERSION "${||.UName||_VER}" LANGUAGES C CXX)
if(||.LName||_BUILD_TESTS)
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

# =============== Config Check =============== #
get_property(USING_MULTI_CONFIG_GENERATOR GLOBAL PROPERTY GENERATOR_IS_MULTI_CONFIG)
if(NOT USING_MULTI_CONFIG_GENERATOR AND NOT CMAKE_BUILD_TYPE)
  message(WARNING "[||.UName||] CMAKE_BUILD_TYPE not explicitly set. Using Debug as default...")
  set(CMAKE_BUILD_TYPE Debug CACHE STRING "Build type" FORCE)
endif()

set(ALLOWED_BUILD_TYPES Debug Release RelWithDebInfo MinSizeRel)
if(NOT USING_MULTI_CONFIG_GENERATOR AND NOT CMAKE_BUILD_TYPE IN_LIST ALLOWED_BUILD_TYPES)
  message(FATAL_ERROR
    "[||.UName||] Invalid build type: ${CMAKE_BUILD_TYPE}."
    "Allowed values are: ${ALLOWED_BUILD_TYPES}"
  )
endif()

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