#include "||.LName||_version.h"

namespace ||.LName||
{
    std::string get_version()
    {
        return "@||.Name||_VERSION@";
    }
    unsigned get_major_version()
    {
        return @||.Name||_VERSION_MAJOR@;
    }
    unsigned get_minor_version()
    {
        return @||.Name||_VERSION_MINOR@+0;
    }
    unsigned get_tweak_version()
    {
        return @||.Name||_VERSION_PATCH@+0;
    }
    unsigned get_patch_version()
    {
        return @||.Name||_VERSION_TWEAK@+0;
    }
}