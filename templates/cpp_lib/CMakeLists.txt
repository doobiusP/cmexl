cmake_minimum_required(VERSION 3.18)

set(LIBRARY_ALIAS ||.SName||)

macro(_mmsg STR)
  message(VERBOSE "[||.UName||] ${STR}")
endmacro()

_mmsg(Now processing ${CMAKE_CURRENT_LIST_FILE})

# Enable Hot Reload for MSVC compilers if supported
if(MSVC AND POLICY CMP0141)
  cmake_policy(SET CMP0141 NEW)
  set(CMAKE_MSVC_DEBUG_INFORMATION_FORMAT
    "$<$<CONFIG:Debug,RelWithDebInfo>:ProgramDatabase>"
    "$<$<CONFIG:Debug,RelWithDebInfo>:EditAndContinue>")
endif()

include("cmake/||.LName||_options.cmake")

_mmsg("Triplet = ${VCPKG_TARGET_TRIPLET}")
_mmsg("Features = ${VCPKG_MANIFEST_FEATURES}")

file(TO_CMAKE_PATH "${CMAKE_TOOLCHAIN_FILE}" TOOLCHAIN_CMAKE_STYLE_PATH)
_mmsg("Toolchain = ${TOOLCHAIN_CMAKE_STYLE_PATH}}")
file(TO_CMAKE_PATH "${VCPKG_CHAINLOAD_TOOLCHAIN_FILE}" CHAINLOAD_CMAKE_STYLE_PATH)
_mmsg("Chainload Toolchain = ${CHAINLOAD_CMAKE_STYLE_PATH}")

# =============== Project Init =============== #
include("cmake/||.LName||_version.cmake")
project(||.Name|| VERSION "${||.UName||_VER}" LANGUAGES C CXX)
if(||.UName||_BUILD_TESTS)
  enable_testing()
endif()
# =============== Config Check =============== #
# TODO: Allow users to create their configs
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

include(CMakePackageConfigHelpers)
include(GNUInstallDirs)

set(||.UName||_GLOBAL_INCLUDE_DIR "${CMAKE_INSTALL_INCLUDEDIR}/||.LName||")
set(||.UName||_GLOBAL_EXPORT_DIR "${CMAKE_INSTALL_LIBDIR}/cmake/||.Name||")
include("cmake/||.LName||_defaults.cmake")
include("cmake/||.LName||_settings_target.cmake")
include("cmake/||.LName||_topo_sort.cmake")

set(||.UName||_BOOST_STACKTRACE_IMPORT stacktrace_basic)
set(||.UName||_NEED_LIBDL OFF)
set(||.UName||_NEED_DBGENG OFF)

function(_||.UName||_find_addr2line OUTVAR)
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
    _||.UName||_find_addr2line(||.UName||_ADDR2LINE_BINARY)
    set(||.UName||_BOOST_STACKTRACE_IMPORT stacktrace_addr2line)
  endif()
endif()

add_subdirectory(src)

write_basic_package_version_file(
  "${CMAKE_CURRENT_BINARY_DIR}/||.Name||ConfigVersion.cmake"
  VERSION "${||.UName||_VER}"
  COMPATIBILITY SameMajorVersion
)

configure_package_config_file(
  "${CMAKE_CURRENT_LIST_DIR}/cmake/||.LName||_config.in.cmake"
  "${CMAKE_CURRENT_BINARY_DIR}/||.Name||Config.cmake"
  INSTALL_DESTINATION "${||.UName||_GLOBAL_EXPORT_DIR} "
)

set(||.UName||_CONFIGURED_COMPS_DIR "${CMAKE_CURRENT_BINARY_DIR}/cmake/comps")

file(GLOB ||.UName||_COMP_TEMPLATES
    "${CMAKE_CURRENT_LIST_DIR}/cmake/comps/*.in.cmake"
)

file(MAKE_DIRECTORY "${||.UName||_CONFIGURED_COMPS_DIR}")

foreach(tmpl IN LISTS ||.UName||_COMP_TEMPLATES)
    get_filename_component(name "${tmpl}" NAME_WE)
    configure_file(
        "${tmpl}"
        "${||.UName||_CONFIGURED_COMPS_DIR}/${name}.cmake"
        @ONLY
    )
endforeach()

install(
  DIRECTORY
    "${||.UName||_CONFIGURED_COMPS_DIR}/"
  DESTINATION
    "${||.UName||_GLOBAL_EXPORT_DIR}/comps"
  COMPONENT
    "export"
)

install(FILES
  ${CMAKE_CURRENT_BINARY_DIR}/||.Name||ConfigVersion.cmake
  ${CMAKE_CURRENT_BINARY_DIR}/||.Name||Config.cmake
  ${CMAKE_CURRENT_LIST_DIR}/cmake/||.LName||_topo_sort.cmake
  DESTINATION 
    "${||.UName||_GLOBAL_EXPORT_DIR}"
  COMPONENT
    "export"
)

if(CMAKE_SOURCE_DIR STREQUAL CMAKE_CURRENT_SOURCE_DIR)
  add_subdirectory(packaging)   
endif()