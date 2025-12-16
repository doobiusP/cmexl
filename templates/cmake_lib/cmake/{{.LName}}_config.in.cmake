@PACKAGE_INIT@
include(CMakeFindDependencyMacro)

macro(_mmsg STR)
  message(VERBOSE "[||.UName||] ${STR}")
endmacro()

# Only 1 Comp
macro(find_external_comp LIBRARY COMP CONFIG_REQUIRED)
    unset(extraArgs)
    if(${CMAKE_FIND_PACKAGE_NAME}_FIND_QUIETLY)
        list(APPEND extraArgs QUIET)
    endif()
    if(${CMAKE_FIND_PACKAGE_NAME}_FIND_REQUIRED)
        list(APPEND extraArgs REQUIRED)
    endif()
    if(${CONFIG_REQUIRED})
        list(APPEND extraArgs CONFIG)
    endif()
    find_package(${LIBRARY} COMPONENTS ${COMP} ${extraArgs})
endmacro()

include("${CMAKE_CURRENT_LIST_DIR}/||.LName||_topo_sort.cmake")

file(GLOB COMP_DEP_FILES "${CMAKE_CURRENT_LIST_DIR}/comps/*.cmake")

set(||.UName||_AVAILABLE_COMPONENTS)
foreach(COMP_DEP IN LISTS COMP_DEP_FILES)
    _mmsg("Found ${COMP_DEP}")
    include(${COMP_DEP})
endforeach()

set(||.UName||_TOPO_SORT)
_get_topo_sort_with_deps(||.UName||_TOPO_SORT ${CMAKE_FIND_PACKAGE_NAME}_FIND_COMPONENTS)

foreach(COMP IN LISTS ||.UName||_TOPO_SORT)
    set(EXTERNAL_DEP_CALL "_||.LName||_get_${COMP}_external_deps")
    cmake_language(CALL ${EXTERNAL_DEP_CALL})
endforeach()

set(${CMAKE_FIND_PACKAGE_NAME}_comps ${||.UName||_TOPO_SORT})

foreach(comp IN LISTS ${CMAKE_FIND_PACKAGE_NAME}_comps)
    if(${CMAKE_FIND_PACKAGE_NAME}_FIND_REQUIRED_${comp} AND
    NOT EXISTS ${CMAKE_CURRENT_LIST_DIR}/||.LName||_${comp}.cmake)
        set(${CMAKE_FIND_PACKAGE_NAME}_NOT_FOUND_MESSAGE "||.Name|| missing required dependency: ${comp}")
        set(${CMAKE_FIND_PACKAGE_NAME}_FOUND FALSE)
        return()
    endif()
endforeach()

include("${CMAKE_CURRENT_LIST_DIR}/||.LName||_build_settings.cmake")
foreach(comp IN LISTS ${CMAKE_FIND_PACKAGE_NAME}_comps)
    include(${CMAKE_CURRENT_LIST_DIR}/||.LName||_${comp}.cmake OPTIONAL)
endforeach()