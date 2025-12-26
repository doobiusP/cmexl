include(GenerateExportHeader)

if(NOT CMAKE_CXX_COMPILER)
    message(FATAL_ERROR "[||.UName||][ERROR] Cannot determine C++ compiler to use")
endif()

string(FIND ${CMAKE_CXX_COMPILER} "mingw" MINGW_INDEX)
set(||.UName||_USING_MINGW FALSE)
if(NOT MINGW_INDEX EQUAL -1)
    set(||.UName||_USING_MINGW TRUE)
endif()

if(CMAKE_SIZEOF_VOID_P EQUAL 8)
    set(||.UName||_64BIT TRUE)
elseif(CMAKE_SIZEOF_VOID_P EQUAL 4)
    set(||.UName||_32BIT TRUE)
else()
    message(FATAL_ERROR "[||.UName||][ERROR] Unable to determine sizeof(void*)")
endif()

# Only ever link to this target PRIVATEly. Avoid putting the CONFIG macros in header files;
# use them only for implementation files
add_library(||.LName||_build_settings INTERFACE)

# ===================== Common Compiler & Linker Flags ===================== #
# ======== Debug Options ======== #
set(||.UName||_MSVC_OPTIONS_DEBUG)
set(||.UName||_MSVC_OPTIONS_DEBUG_LINKING /DEBUG:FULL)
set(||.UName||_MSVC_OPTIONS_DEBUG_DIAGNOSTICS /W4)
set(||.UName||_MSVC_OPTIONS_DEBUG_COMPILING /Oy- /MP /permissive-) #/Od /Ob0 /Zi added automatically with Debug
# /permissive- automatically set by /std:c++20 but adding here just in case
set(||.UName||_GNU_OPTIONS_DEBUG)
set(||.UName||_GNU_OPTIONS_DEBUG_LINKING -rdynamic)
set(||.UName||_GNU_OPTIONS_DEBUG_DIAGNOSTICS -Wall -Wextra -Wpedantic)
set(||.UName||_GNU_OPTIONS_DEBUG_COMPILING -O0 -ggdb -pg -fno-omit-frame-pointer) # -g added automatically with Debug

if(||.UName||_USING_MINGW)
    list(REMOVE_ITEM ||.UName||_GNU_OPTIONS_DEBUG_LINKING -rdynamic) #TODO: Once DLLs are supported, what about mingw's -export-dynamic?
    list(REMOVE_ITEM ||.UName||_GNU_OPTIONS_DEBUG_COMPILING -pg)
endif()

list(APPEND ||.UName||_MSVC_OPTIONS_DEBUG
    ${||.UName||_MSVC_OPTIONS_DEBUG_COMPILING}
    ${||.UName||_MSVC_OPTIONS_DEBUG_DIAGNOSTICS}
)
list(APPEND ||.UName||_GNU_OPTIONS_DEBUG
    ${||.UName||_GNU_OPTIONS_DEBUG_COMPILING}
    ${||.UName||_GNU_OPTIONS_DEBUG_DIAGNOSTICS}
)
# ======== Release/RelWithDebInfo/MinSizeRel Variants ======== #
set(||.UName||_MSVC_OPTIONS_RELEASE)
set(||.UName||_MSVC_OPTIONS_RELEASE_LINKING)
set(||.UName||_MSVC_OPTIONS_RELEASE_DIAGNOSTICS)
set(||.UName||_MSVC_OPTIONS_RELEASE_COMPILING /MP /permissive-) #/O2 /Ob2 added automatically with Release
# /permissive- automatically set by /std:c++20 but adding here just in case

set(||.UName||_GNU_OPTIONS_RELEASE)
set(||.UName||_GNU_OPTIONS_RELEASE_LINKING -rdynamic)
set(||.UName||_GNU_OPTIONS_RELEASE_DIAGNOSTICS -Wall)
set(||.UName||_GNU_OPTIONS_RELEASE_COMPILING) # RELEASE enables -O3 which enables -finline-functions

if(||.UName||_32BIT)
    list(APPEND ||.UName||_MSVC_OPTIONS_RELEASE_COMPILING /Oy)
    list(APPEND ||.UName||_GNU_OPTIONS_RELEASE_COMPILING -fomit-frame-pointer)
else()
    if(NOT ||.UName||_64BIT)
        message(WARNING "[||.UName||] Can't determine CMAKE_SIZEOF_VOID_P")
    endif()
    list(APPEND ||.UName||_MSVC_OPTIONS_RELEASE_COMPILING /Oy-)
    list(APPEND ||.UName||_GNU_OPTIONS_RELEASE_COMPILING -fno-omit-frame-pointer)
endif()

if(||.UName||_USING_MINGW)
    list(REMOVE_ITEM ||.UName||_GNU_OPTIONS_RELEASE_LINKING "-rdynamic") #TODO: Once DLLs are supported, what about mingw's -export-dynamic?
endif()

list(APPEND ||.UName||_MSVC_OPTIONS_RELEASE
    ${||.UName||_MSVC_OPTIONS_RELEASE_COMPILING}
    ${||.UName||_MSVC_OPTIONS_RELEASE_DIAGNOSTICS}
)
list(APPEND ||.UName||_GNU_OPTIONS_RELEASE
    ${||.UName||_GNU_OPTIONS_RELEASE_COMPILING}
    ${||.UName||_GNU_OPTIONS_RELEASE_DIAGNOSTICS}
)

message(VERBOSE "[||.UName||] sizeof(void*): ${CMAKE_SIZEOF_VOID_P}")

message(VERBOSE "[||.UName||] Explicitly added flags:")
message(VERBOSE "[||.UName||][FLAGS] MSVC+Debug: ${||.UName||_MSVC_OPTIONS_DEBUG}")
message(VERBOSE "[||.UName||][FLAGS] GNU+Debug: ${||.UName||_GNU_OPTIONS_DEBUG}")
message(VERBOSE "[||.UName||][FLAGS] MSVC+Release: ${||.UName||_MSVC_OPTIONS_RELEASE}")
message(VERBOSE "[||.UName||][FLAGS] GNU+Release: ${||.UName||_GNU_OPTIONS_RELEASE}")

message(VERBOSE "[||.UName||][FLAGS] MSVC+Debug linker flags: ${||.UName||_MSVC_OPTIONS_DEBUG_LINKING}")
message(VERBOSE "[||.UName||][FLAGS] GNU+Debug linker flags: ${||.UName||_GNU_OPTIONS_DEBUG_LINKING}")
if(CMAKE_GENERATOR MATCHES "Visual Studio")
    message(VERBOSE "[||.UName||][Visual Studio] Platform Name: ${CMAKE_VS_PLATFORM_NAME}")
    message(VERBOSE "[||.UName||][Visual Studio] Platform Version: ${CMAKE_VS_WINDOWS_TARGET_PLATFORM_VERSION}")
    message(VERBOSE "[||.UName||][Visual Studio] Platform Toolset: ${CMAKE_VS_PLATFORM_TOOLSET}")
endif()

# ==================== Target Properties ==================== #
target_include_directories(||.LName||_build_settings
    INTERFACE
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/src>
        $<INSTALL_INTERFACE:${CMAKE_INSTALL_INCLUDEDIR}>
)

target_compile_definitions(||.LName||_build_settings
    INTERFACE
    ||.UName||_CONFIG_$<UPPER_CASE:$<CONFIG>>
)

target_compile_options(||.LName||_build_settings
    INTERFACE
    $<$<CONFIG:Debug>:$<IF:$<CXX_COMPILER_FRONTEND_VARIANT:MSVC>,${||.UName||_MSVC_OPTIONS_DEBUG},${||.UName||_GNU_OPTIONS_DEBUG}>>
    $<$<NOT:$<CONFIG:Debug>>:$<IF:$<CXX_COMPILER_FRONTEND_VARIANT:MSVC>,${||.UName||_MSVC_OPTIONS_RELEASE},${||.UName||_GNU_OPTIONS_RELEASE}>>
)

target_link_options(||.LName||_build_settings
    INTERFACE
    $<$<CONFIG:Debug>:$<IF:$<CXX_COMPILER_FRONTEND_VARIANT:MSVC>,${||.UName||_MSVC_OPTIONS_DEBUG_LINKING},${||.UName||_GNU_OPTIONS_DEBUG_LINKING}>>
)

install(
  TARGETS 
    ||.LName||_build_settings
  EXPORT 
    ||.LName||_build_settings
)

install(
  EXPORT
    ||.LName||_build_settings
  DESTINATION
    "${||.UName||_GLOBAL_EXPORT_DIR}"
  NAMESPACE 
    ${LIBRARY_ALIAS}::
  COMPONENT
    "export"
)
