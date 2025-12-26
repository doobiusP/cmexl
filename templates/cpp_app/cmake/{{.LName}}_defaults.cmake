include(GenerateExportHeader)

function(copy_mingw64_dependency PATH_TO_DEP NAME_OF_DEP)
    execute_process(
        COMMAND
        ${CMAKE_CXX_COMPILER} --print-file-name=${NAME_OF_DEP}
        OUTPUT_VARIABLE
        _PATH_TO_DEP
        OUTPUT_STRIP_TRAILING_WHITESPACE
    )
    if(${_PATH_TO_DEP} STREQUAL ${NAME_OF_DEP})
        message(FATAL_ERROR "[||.UName||][MINGW64] Unable to find ${NAME_OF_DEP} for mingw64")
    endif()
    message(VERBOSE "[||.UName||][MINGW64] Found ${NAME_OF_DEP} at ${_PATH_TO_DEP}")
    set(${PATH_TO_DEP} ${_PATH_TO_DEP} PARENT_SCOPE)

    # Uncomment the below lines if you have write access to the binary directory, otherwise manually copy
    # configure_file(${_PATH_TO_DEP} ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/${NAME_OF_DEP} COPYONLY)
    # message(VERBOSE "[NTEST2][MINGW64] Finished copying over ${NAME_OF_DEP} to ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}...")
endfunction()

set(NTEST2_INCLUDE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/src)
set(NTEST2_TOP_BINARY_DIR ${CMAKE_CURRENT_BINARY_DIR})
macro(add_standards TARGET)
    _mmsg("adding standards for ${TARGET}")

    set_target_properties(
        ${TARGET} 
        PROPERTIES
            C_STANDARD 99
            C_STANDARD_REQUIRED ON
            C_EXTENSIONS OFF
            CXX_STANDARD 17
            CXX_STANDARD_REQUIRED ON
            CXX_EXTENSIONS OFF
            POSITION_INDEPENDENT_CODE ON
            RUNTIME_OUTPUT_DIRECTORY ${NTEST2_TOP_BINARY_DIR}/bin
            ARCHIVE_OUTPUT_DIRECTORY ${NTEST2_TOP_BINARY_DIR}/lib
            LIBRARY_OUTPUT_DIRECTORY ${NTEST2_TOP_BINARY_DIR}/bin
    )

    target_include_directories(
        ${TARGET}
        PRIVATE
            $<BUILD_INTERFACE:${NTEST2_INCLUDE_DIR}/src>
    )
endmacro()

macro(add_dll_support TARGET)
    _mmsg("adding dll support for ${TARGET}")

    set_target_properties(
        ${TARGET} 
        PROPERTIES
            CXX_VISIBILITY_PRESET hidden
            VISIBILITY_INLINES_HIDDEN YES
            IMPORTED_PDB_FILE "${TARGET}.pdb"
    )

    generate_export_header(${TARGET}) 

    target_include_directories(
        ${TARGET}
        PUBLIC 
            $<BUILD_INTERFACE:${CMAKE_CURRENT_BINARY_DIR}>
    )
endmacro()

string(FIND ${CMAKE_CXX_COMPILER} "mingw" MINGW_INDEX)
set(||.UName||_USING_MINGW FALSE)
if(NOT MINGW_INDEX EQUAL -1)
    set(||.UName||_USING_MINGW TRUE)
endif()

if(||.UName||_USING_MINGW AND CMAKE_SOURCE_DIR STREQUAL CMAKE_CURRENT_SOURCE_DIR)
    copy_mingw64_dependency(LIBSTDCXX-6_DLL libstdc++-6.dll)
    copy_mingw64_dependency(LIBWINPTHREAD-1_DLL libwinpthread-1.dll)
    if(CMAKE_SIZEOF_VOID_P EQUAL 4)
        copy_mingw64_dependency(LIBGCC_S_DW2-1_DLL libgcc_s_dw2-1.dll)
    else()
        copy_mingw64_dependency(LIBGCC_S_SEH-1_DLL libgcc_s_seh-1.dll)
    endif()
endif()