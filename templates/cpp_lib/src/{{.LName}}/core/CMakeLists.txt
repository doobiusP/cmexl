add_library(||.LName||_core)
add_library(${LIBRARY_ALIAS}::core ALIAS ||.LName||_core)
set(||.UName||_CORE_VERSION 0.1.0.0)
set(||.UName||_CORE_SOVERSION 1)

find_package(ZLIB REQUIRED)

target_link_libraries(
  ||.LName||_core
	PRIVATE
    ||.LName||_build_settings
  PUBLIC
    ZLIB::ZLIB
)

set_target_properties(
  ||.LName||_core 
  PROPERTIES
    VERSION "${||.UName||_CORE_VERSION}"
    SOVERSION "${||.UName||_CORE_SOVERSION}"
    EXPORT_NAME core
)

add_dll_support(||.LName||_core)

target_sources(
  ||.LName||_core
  PRIVATE print.cpp
  PUBLIC
    FILE_SET HEADERS
      FILES print.h
    FILE_SET generated_headers
      TYPE HEADERS
      BASE_DIRS $<TARGET_PROPERTY:||.LName||_core,BINARY_DIR>
      FILES ${CMAKE_CURRENT_BINARY_DIR}/||.LName||_core_export.h
)

install(
  TARGETS 
    ||.LName||_core
  EXPORT 
    ||.LName||_core
  FILE_SET 
    HEADERS 
    DESTINATION "${||.UName||_GLOBAL_INCLUDE_DIR}/core"
    COMPONENT "core"
  FILE_SET 
    generated_headers
    DESTINATION "${||.UName||_GLOBAL_INCLUDE_DIR}/core"
    COMPONENT "core"
  RUNTIME 
    DESTINATION ${CMAKE_INSTALL_BINDIR}/$<CONFIG>
    COMPONENT "core"
  LIBRARY
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/$<CONFIG>
    COMPONENT "core"
  ARCHIVE
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/$<CONFIG>
    COMPONENT "core"
  INCLUDES
    DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}
)

install(
  EXPORT
    ||.LName||_core
  DESTINATION
    "${||.UName||_GLOBAL_EXPORT_DIR}"
  NAMESPACE 
    ${LIBRARY_ALIAS}::
  COMPONENT
    "core"
)

if(BUILD_SHARED_LIBS)
  install(
      FILES
        $<TARGET_PDB_FILE:||.LName||_core>
      DESTINATION 
        $<IF:$<STREQUAL:$<TARGET_PROPERTY:||.LName||_core,TYPE>,SHARED_LIBRARY>,bin,lib>/$<CONFIG>
      COMPONENT
        "core"
      OPTIONAL
  )
endif()