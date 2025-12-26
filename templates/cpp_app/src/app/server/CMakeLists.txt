add_executable(||.LName||_server)

target_link_libraries(
    ||.LName||_server
    PRIVATE
        ||.SName||::core
)

if(WIN32)
    target_link_libraries(||.LName||_server PRIVATE ws2_32)
endif()

add_standards(||.LName||_server)

target_sources(
    ||.LName||_server
    PRIVATE 
        server.cpp
)