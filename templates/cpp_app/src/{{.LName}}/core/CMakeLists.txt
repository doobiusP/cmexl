add_library(||.LName||_core)
add_library(||.SName||::core ALIAS ||.LName||_core)

set(||.LName||_CORE_VERSION 0.1.0.0)
set(||.LName||_CORE_SOVERSION 1)

set_target_properties(
  ||.LName||_core
  PROPERTIES
  VERSION "${||.LName||_CORE_VERSION}"
  SOVERSION "${||.LName||_CORE_SOVERSION}"
)

target_link_libraries(
  ||.LName||_core
  PRIVATE
  ||.SName||::build
)
add_standards(||.LName||_core)
add_dll_support(||.LName||_core)

target_sources(
  ||.LName||_core
  PRIVATE network.cpp
  PUBLIC
  FILE_SET HEADERS
  FILES network.h
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
  RUNTIME 
    DESTINATION ${CMAKE_INSTALL_BINDIR}/$<CONFIG>
  LIBRARY
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/$<CONFIG>
  ARCHIVE
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/$<CONFIG>
  FILE_SET 
    HEADERS 
    DESTINATION "${FGAME_GLOBAL_INCLUDE_DIR}/core"
  FILE_SET 
    generated_headers
    DESTINATION "${FGAME_GLOBAL_INCLUDE_DIR}/core"
)