add_library(||.LName||_core)
add_library(||.SName||::core ALIAS ||.LName||_core)

set(||.UName||_CORE_VERSION 0.1.0.0)
set(||.UName||_CORE_SOVERSION 1)

set_target_properties(
  ||.LName||_core 
  PROPERTIES
    VERSION "${||.UName||_CORE_VERSION}"
    SOVERSION "${||.UName||_CORE_SOVERSION}"
)

add_dll_support(||.LName||_core)
add_standards(||.LName||_core)

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