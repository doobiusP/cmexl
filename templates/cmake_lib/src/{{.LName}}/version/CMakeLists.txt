add_library(||.LName||_version STATIC)

configure_file(||.LName||_version.in.cpp ||.LName||_version.cpp @ONLY)

target_link_libraries(
	||.LName||_version
	PRIVATE
    ||.LName||_build_settings
)

target_sources(
  ||.LName||_version
  PRIVATE ${CMAKE_CURRENT_BINARY_DIR}/||.LName||_version.cpp
  PUBLIC
    FILE_SET HEADERS
    FILES ||.LName||_version.h
)