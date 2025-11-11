if(NOT CMAKE_CXX_COMPILER)
    message(FATAL_ERROR "[DOOBIUS] Cannot determine C++ compiler to use")
endif()
string(FIND ${CMAKE_CXX_COMPILER} "mingw" MINGW_INDEX)
set(DOOBIUS_USING_MINGW FALSE)
if(NOT MINGW_INDEX EQUAL -1)
    set(DOOBIUS_USING_MINGW TRUE)
endif()

if(CMAKE_SIZEOF_VOID_P EQUAL 8)
    set(DOOBIUS_64BIT TRUE)
elseif(CMAKE_SIZEOF_VOID_P EQUAL 4)
    set(DOOBIUS_32BIT TRUE)
else()
    message(FATAL_ERROR "[DOOBIUS] Unable to determine sizeof(void*)")
endif()

set(CMAKE_POSITION_INDEPENDENT_CODE ON)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)

add_library(DoobiusCompileAndLink INTERFACE)
add_library(Doobius::proj_build ALIAS DoobiusCompileAndLink)

if(DOOBIUS_USING_MINGW)
    function(copy_mingw64_dependency PATH_TO_DEP NAME_OF_DEP)
        execute_process(
            COMMAND 
            ${CMAKE_CXX_COMPILER} --print-file-name=${NAME_OF_DEP}
            OUTPUT_VARIABLE 
            _PATH_TO_DEP
            OUTPUT_STRIP_TRAILING_WHITESPACE
            ECHO_OUTPUT_VARIABLE
            COMMAND_ECHO STDOUT
        )
        if(${_PATH_TO_DEP} STREQUAL ${NAME_OF_DEP})
            message(FATAL_ERROR "[DOOBIUS][MINGW64] Unable to find ${NAME_OF_DEP} for mingw64")
        endif()
        message(STATUS "[DOOBIUS][MINGW64] Found ${NAME_OF_DEP} at ${_PATH_TO_DEP}")
        set(${PATH_TO_DEP} ${_PATH_TO_DEP} PARENT_SCOPE)
        configure_file(${_PATH_TO_DEP} ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/${NAME_OF_DEP} COPYONLY)
        message(STATUS "[DOOBIUS][MINGW64] Finished copying over ${NAME_OF_DEP} to ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}...")
    endfunction()
    copy_mingw64_dependency(LIBSTDCXX-6_DLL libstdc++-6.dll)
    copy_mingw64_dependency(LIBWINPTHREAD-1_DLL libwinpthread-1.dll)
    if(CMAKE_SIZEOF_VOID_P EQUAL 4)
        copy_mingw64_dependency(LIBGCC_S_DW2-1_DLL libgcc_s_dw2-1.dll)
    else()
        copy_mingw64_dependency(LIBGCC_S_SEH-1_DLL libgcc_s_seh-1.dll) 
    endif()
endif()

# ===================== Common Compiler & Linker Flags ===================== #
# ======== Debug Options ======== #
set(DOOBIUS_MSVC_OPTIONS_DEBUG)
set(DOOBIUS_MSVC_OPTIONS_DEBUG_LINKING /DEBUG:FULL)
set(DOOBIUS_MSVC_OPTIONS_DEBUG_DIAGNOSTICS /W4)
set(DOOBIUS_MSVC_OPTIONS_DEBUG_COMPILING /Oy- /MP /permissive-)                #/Od /Ob0 /Zi added automatically with Debug
                                                                               # /permissive- automatically set by /std:c++20 but adding here just in case
set(DOOBIUS_GNU_OPTIONS_DEBUG)
set(DOOBIUS_GNU_OPTIONS_DEBUG_LINKING -rdynamic)
set(DOOBIUS_GNU_OPTIONS_DEBUG_DIAGNOSTICS -Wall -Wextra -Wpedantic)
set(DOOBIUS_GNU_OPTIONS_DEBUG_COMPILING -O0 -ggdb -pg -fno-omit-frame-pointer) # -g added automatically with Debug

if(DOOBIUS_USING_MINGW)
    list(REMOVE_ITEM DOOBIUS_GNU_OPTIONS_DEBUG_LINKING -rdynamic) #TODO: Once DLLs are supported for DoobiusGameLibrary, what about mingw's -export-dynamic?
    list(REMOVE_ITEM DOOBIUS_GNU_OPTIONS_DEBUG_COMPILING -pg)
endif()

list(APPEND DOOBIUS_MSVC_OPTIONS_DEBUG 
    ${DOOBIUS_MSVC_OPTIONS_DEBUG_COMPILING}
    ${DOOBIUS_MSVC_OPTIONS_DEBUG_DIAGNOSTICS}
)
list(APPEND DOOBIUS_GNU_OPTIONS_DEBUG 
    ${DOOBIUS_GNU_OPTIONS_DEBUG_COMPILING}
    ${DOOBIUS_GNU_OPTIONS_DEBUG_DIAGNOSTICS}
)
# ======== Release/RelWithDebInfo/MinSizeRel Variants ======== #
set(DOOBIUS_MSVC_OPTIONS_RELEASE)
set(DOOBIUS_MSVC_OPTIONS_RELEASE_LINKING)
set(DOOBIUS_MSVC_OPTIONS_RELEASE_DIAGNOSTICS)
set(DOOBIUS_MSVC_OPTIONS_RELEASE_COMPILING /MP /permissive-) #/O2 /Ob2 added automatically with Release
                                                             # /permissive- automatically set by /std:c++20 but adding here just in case

set(DOOBIUS_GNU_OPTIONS_RELEASE)
set(DOOBIUS_GNU_OPTIONS_RELEASE_LINKING -rdynamic)
set(DOOBIUS_GNU_OPTIONS_RELEASE_DIAGNOSTICS -Wall)
set(DOOBIUS_GNU_OPTIONS_RELEASE_COMPILING)                   # RELEASE enables -O3 which enables -finline-functions

if(DOOBIUS_32BIT)
    list(APPEND DOOBIUS_MSVC_OPTIONS_RELEASE_COMPILING /Oy)
    list(APPEND DOOBIUS_GNU_OPTIONS_RELEASE_COMPILING -fomit-frame-pointer)
else()
    if(NOT DOOBIUS_64BIT)
        message(WARNING "[DOOBIUS] Can't determine CMAKE_SIZEOF_VOID_P")
    endif()
    list(APPEND DOOBIUS_MSVC_OPTIONS_RELEASE_COMPILING /Oy-)
    list(APPEND DOOBIUS_GNU_OPTIONS_RELEASE_COMPILING -fno-omit-frame-pointer)
endif()

if(DOOBIUS_USING_MINGW)
    list(REMOVE_ITEM DOOBIUS_GNU_OPTIONS_RELEASE_LINKING "-rdynamic") #TODO: Once DLLs are supported for DoobiusGameLibrary, what about mingw's -export-dynamic?
endif()

list(APPEND DOOBIUS_MSVC_OPTIONS_RELEASE 
    ${DOOBIUS_MSVC_OPTIONS_RELEASE_COMPILING}
    ${DOOBIUS_MSVC_OPTIONS_RELEASE_DIAGNOSTICS}
)
list(APPEND DOOBIUS_GNU_OPTIONS_RELEASE 
    ${DOOBIUS_GNU_OPTIONS_RELEASE_COMPILING}
    ${DOOBIUS_GNU_OPTIONS_RELEASE_DIAGNOSTICS}
)

if(DOOBIUS_VERBOSE)
    message(STATUS "[DOOBIUS] sizeof(void*): ${CMAKE_SIZEOF_VOID_P}")

    message(STATUS "[DOOBIUS] Explicitly added flags:")
    message(STATUS "[DOOBIUS][FLAGS] MSVC+Debug: ${DOOBIUS_MSVC_OPTIONS_DEBUG}")
    message(STATUS "[DOOBIUS][FLAGS] GNU+Debug: ${DOOBIUS_GNU_OPTIONS_DEBUG}")
    message(STATUS "[DOOBIUS][FLAGS] MSVC+Release: ${DOOBIUS_MSVC_OPTIONS_RELEASE}")
    message(STATUS "[DOOBIUS][FLAGS] GNU+Release: ${DOOBIUS_GNU_OPTIONS_RELEASE}")

    message(STATUS "[DOOBIUS][FLAGS] MSVC+Debug linker flags: ${DOOBIUS_MSVC_OPTIONS_DEBUG_LINKING}")
    message(STATUS "[DOOBIUS][FLAGS] GNU+Debug linker flags: ${DOOBIUS_GNU_OPTIONS_DEBUG_LINKING}")
    if (CMAKE_GENERATOR MATCHES "Visual Studio")
        message(STATUS "[DOOBIUS][Visual Studio] Platform Name: ${CMAKE_VS_PLATFORM_NAME}")
        message(STATUS "[DOOBIUS][Visual Studio] Platform Version: ${CMAKE_VS_WINDOWS_TARGET_PLATFORM_VERSION}")
        message(STATUS "[DOOBIUS][Visual Studio] Platform Toolset: ${CMAKE_VS_PLATFORM_TOOLSET}")
    endif()
endif()

# ==================== Target Properties ==================== #
target_compile_features(DoobiusCompileAndLink
    INTERFACE
    cxx_std_20
)
target_compile_definitions(DoobiusCompileAndLink
    INTERFACE
	DOOBIUS_PATH_TO_CONFIGS_DIR="${PROJECT_SOURCE_DIR}" #TODO: Allow a user to set this.
	DOOBIUS_PATH_TO_OUT_LOG_DIR="${PROJECT_SOURCE_DIR}/out/build/${DOOBIUS_PRESET_NAME}/log/$<LOWER_CASE:$<CONFIG>>" #TODO: Allow a user to set this.
	DOOBIUS_CONFIG_$<UPPER_CASE:$<CONFIG>>
)

target_compile_options(DoobiusCompileAndLink
    INTERFACE
    $<$<CONFIG:Debug>:$<IF:$<CXX_COMPILER_FRONTEND_VARIANT:MSVC>,${DOOBIUS_MSVC_OPTIONS_DEBUG},${DOOBIUS_GNU_OPTIONS_DEBUG}>>
    $<$<NOT:$<CONFIG:Debug>>:$<IF:$<CXX_COMPILER_FRONTEND_VARIANT:MSVC>,${DOOBIUS_MSVC_OPTIONS_RELEASE},${DOOBIUS_GNU_OPTIONS_RELEASE}>>
)

target_link_options(DoobiusCompileAndLink
    INTERFACE
    $<$<CONFIG:Debug>:$<IF:$<CXX_COMPILER_FRONTEND_VARIANT:MSVC>,${DOOBIUS_MSVC_OPTIONS_DEBUG_LINKING},${DOOBIUS_GNU_OPTIONS_DEBUG_LINKING}>>
)

target_include_directories(DoobiusCompileAndLink 
	INTERFACE
    ${PROJECT_SOURCE_DIR}/include
)

# target_sources(DoobiusCompileAndLink
#     INTERFACE
#     FILE_SET public_headers
#     TYPE HEADERS
#     BASE_DIRS ${CMAKE_CURRENT_LIST_DIR}/include
#     FILES
#         doobius/stdtypes.h
#         doobius/logging/logging.h
#         doobius/logging/custom_assert.h
# )

# #TODO: This not working as intended
# source_group(TREE ${CMAKE_CURRENT_SOURCE_DIR}/include
#              PREFIX "Header Files"
#              FILES
#                 include/doobius/stdtypes.h
#                 include/doobius/logging/logging.h
#                 include/doobius/logging/custom_assert.h
# )

# If you use VSCode's or Visual Studio CMake integration, this is obselete, just use that extension
if(DOOBIUS_USE_VCPKG AND DOOBIUS_COPY_VCPKG_INCLUDE)
    message(STATUS "[DOOBIUS][VCPKG] Now copying over vcpkg_installed/${VCPKG_TARGET_TRIPLET}/include/ to dev-include. This operation may take time...")
    file( # DO NOT ADD ANY / AT THE END, IT WILL NOT COME UNDER INCLUDE! See the copying files chapter from Professional CMake
        COPY ${VCPKG_INSTALLED_DIR}/${VCPKG_TARGET_TRIPLET}/include 
        DESTINATION ${PROJECT_SOURCE_DIR}/dev-include/${DOOBIUS_PRESET_NAME}
    )
endif()