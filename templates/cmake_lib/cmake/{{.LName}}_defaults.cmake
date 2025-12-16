file(RELATIVE_PATH relDir
  ${CMAKE_CURRENT_BINARY_DIR}/${CMAKE_INSTALL_BINDIR}
  ${CMAKE_CURRENT_BINARY_DIR}/${CMAKE_INSTALL_LIBDIR}
)
set(CMAKE_INSTALL_RPATH $ORIGIN $ORIGIN/${relDir})

set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/bin)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/lib)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/bin)

set(CMAKE_POSITION_INDEPENDENT_CODE ON)
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_C_STANDARD 99)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

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
        message(FATAL_ERROR "[||.UName||][MINGW64] Unable to find ${NAME_OF_DEP} for mingw64")
    endif()
    message(VERBOSE "[||.UName||][MINGW64] Found ${NAME_OF_DEP} at ${_PATH_TO_DEP}")
    set(${PATH_TO_DEP} ${_PATH_TO_DEP} PARENT_SCOPE)
    configure_file(${_PATH_TO_DEP} ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/${NAME_OF_DEP} COPYONLY)
    message(VERBOSE "[||.UName||][MINGW64] Finished copying over ${NAME_OF_DEP} to ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}...")
endfunction()

# if(||.UName||_USING_MINGW AND CMAKE_SOURCE_DIR STREQUAL CMAKE_CURRENT_SOURCE_DIR)
#     copy_mingw64_dependency(LIBSTDCXX-6_DLL libstdc++-6.dll)
#     copy_mingw64_dependency(LIBWINPTHREAD-1_DLL libwinpthread-1.dll)
#     if(CMAKE_SIZEOF_VOID_P EQUAL 4)
#         copy_mingw64_dependency(LIBGCC_S_DW2-1_DLL libgcc_s_dw2-1.dll)
#     else()
#         copy_mingw64_dependency(LIBGCC_S_SEH-1_DLL libgcc_s_seh-1.dll)
#     endif()
# endif()

macro(add_dll_support TARGET)
    message(VERBOSE "[||.UName||] adding dll support for ${TARGET}")

    set_target_properties(
        ${TARGET} 
        PROPERTIES
            CXX_VISIBILITY_PRESET hidden
            VISIBILITY_INLINES_HIDDEN YES
    )

    generate_export_header(${TARGET}) 

    target_include_directories(
        ${TARGET}
        PUBLIC 
        $<BUILD_INTERFACE:${CMAKE_CURRENT_BINARY_DIR}>
    )

    set_target_properties(
        ${TARGET} 
        PROPERTIES
            IMPORTED_PDB_FILE "${TARGET}.pdb"
    )
endmacro()