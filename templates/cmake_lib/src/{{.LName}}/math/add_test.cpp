#define BOOST_TEST_MODULE MathAddTest
#include "||.LName||/math/add.h"
#include <boost/test/included/unit_test.hpp>

BOOST_AUTO_TEST_CASE(add_positive)
{
    BOOST_CHECK_EQUAL(add(2, 3), 5);
    BOOST_CHECK_EQUAL(add(0, 7), 7);
}

BOOST_AUTO_TEST_CASE(add_negative)
{
    BOOST_CHECK_EQUAL(add(-4, -5), -9);
    BOOST_CHECK_EQUAL(add(-4, 4),  0);
}

BOOST_AUTO_TEST_CASE(add_zero)
{
    BOOST_CHECK_EQUAL(add(0, 0), 0);
}