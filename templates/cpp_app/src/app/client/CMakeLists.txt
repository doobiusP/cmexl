add_executable(||.LName||_client)

target_link_libraries(
    ||.LName||_client
    PRIVATE
        ||.SName||::core
)

if(WIN32)
    target_link_libraries(||.LName||_client PRIVATE ws2_32)
endif()

add_standards(||.LName||_client)

target_sources(
    ||.LName||_client
    PRIVATE 
        client.cpp
)