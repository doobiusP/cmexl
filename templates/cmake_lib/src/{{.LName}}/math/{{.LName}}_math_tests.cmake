add_executable(||.LName||_math_test add_test.cpp)

find_package(boost_unit_test_framework CONFIG REQUIRED)

target_link_libraries(
    ||.LName||_math_test
    PRIVATE
        ||.LName||_build_settings
        ${LIBRARY_ALIAS}::math
        Boost::unit_test_framework
)

add_test(NAME add_tests COMMAND ||.LName||_math_test)
set_tests_properties(add_tests PROPERTIES LABELS "math")