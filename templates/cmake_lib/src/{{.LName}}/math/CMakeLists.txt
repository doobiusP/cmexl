add_library(||.LName||_math)
add_library(${LIBRARY_ALIAS}::math ALIAS ||.LName||_math)

set(||.UName||_MATH_VERSION 0.1.0.0)
set(||.UName||_MATH_SOVERSION 1)

target_link_libraries(
  ||.LName||_math
	PRIVATE
	  ||.LName||_build_settings
  PUBLIC
    ${LIBRARY_ALIAS}::core
)

set_target_properties(
  ||.LName||_math
  PROPERTIES
    VERSION ${||.UName||_MATH_VERSION}
    SOVERSION ${||.UName||_MATH_SOVERSION}
    EXPORT_NAME math
)

add_dll_support(||.LName||_math)

target_sources(
  ||.LName||_math
  PRIVATE add.cpp
  PUBLIC
    FILE_SET HEADERS
      FILES add.h
    FILE_SET generated_headers
      TYPE HEADERS
      BASE_DIRS $<TARGET_PROPERTY:||.LName||_math,BINARY_DIR>
      FILES ${CMAKE_CURRENT_BINARY_DIR}/||.LName||_math_export.h
)

install(
  TARGETS 
    ||.LName||_math
  EXPORT 
    ||.LName||_math
  FILE_SET 
    HEADERS 
    DESTINATION ${||.UName||_GLOBAL_INCLUDE_DIR}/math
    COMPONENT "math"
  FILE_SET 
    generated_headers
    DESTINATION ${||.UName||_GLOBAL_INCLUDE_DIR}/math
    COMPONENT "math"
  RUNTIME 
    DESTINATION ${CMAKE_INSTALL_BINDIR}/$<CONFIG>
    COMPONENT "math"
  LIBRARY
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/$<CONFIG>
    COMPONENT "math"
  ARCHIVE
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/$<CONFIG>
    COMPONENT "math"
  INCLUDES
    DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}
)

install(
  EXPORT
    ||.LName||_math
  DESTINATION
    ${||.UName||_GLOBAL_EXPORT_DIR}
  NAMESPACE 
    ${LIBRARY_ALIAS}::
  COMPONENT
    "math"
)

if(BUILD_SHARED_LIBS)
  install(
      FILES
        $<TARGET_PDB_FILE:||.LName||_math>
      DESTINATION 
        $<IF:$<STREQUAL:$<TARGET_PROPERTY:||.LName||_math,TYPE>,SHARED_LIBRARY>,bin,lib>/$<CONFIG>
      COMPONENT
        "math"
      OPTIONAL
  )
endif()
if(||.UName||_BUILD_TESTS)
  include("||.LName||_math_tests.cmake")
endif()