add_executable(||.LName||_server)

target_link_libraries(
    ||.LName||_server
    PRIVATE
        ||.SName||::core
        ||.SName||::build
)
add_standards(||.LName||_server)

if(WIN32)
    target_link_libraries(||.LName||_server PRIVATE ws2_32)
endif()


target_sources(
    ||.LName||_server
    PRIVATE 
        server.cpp
)

install(
  TARGETS 
    ||.LName||_server
  EXPORT
    ||.LName||_server
  RUNTIME 
    DESTINATION ${CMAKE_INSTALL_BINDIR}/$<CONFIG>
)