add_executable(||.LName||_client)

target_link_libraries(
    ||.LName||_client
    PRIVATE
        ||.SName||::core
        ||.SName||::build
)
add_standards(||.LName||_client)

if(WIN32)
    target_link_libraries(||.LName||_client PRIVATE ws2_32)
endif()


target_sources(
    ||.LName||_client
    PRIVATE 
        client.cpp
)

install(
  TARGETS 
    ||.LName||_client
  EXPORT 
    ||.LName||_client
  RUNTIME 
    DESTINATION ${CMAKE_INSTALL_BINDIR}/$<CONFIG>
)
