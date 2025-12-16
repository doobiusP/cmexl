file(GLOB COMP_DEP_FILES "${CMAKE_CURRENT_LIST_DIR}/comps/*.cmake")

set(||.UName||_AVAILABLE_COMPONENTS)
foreach(COMP_DEP IN LISTS COMP_DEP_FILES)
    include(${COMP_DEP})
endforeach()

set(||.UName||_INITIAL_REQUEST)
foreach(COMP IN LISTS ||.UName||_AVAILABLE_COMPONENTS)
    if(||.UName||_BUILD_COMP_${COMP})
        list(APPEND ||.UName||_INITIAL_REQUEST ${COMP})
    endif()
endforeach()

set(||.UName||_TOPO_SORT)
_get_topo_sort_with_deps(||.UName||_TOPO_SORT ||.UName||_INITIAL_REQUEST)

foreach(COMP IN LISTS ||.UName||_TOPO_SORT)
    add_subdirectory("||.LName||/${COMP}")
endforeach()

if(||.UName||_BUILD_EXAMPLES)
    add_subdirectory(examples/calculator)
endif()