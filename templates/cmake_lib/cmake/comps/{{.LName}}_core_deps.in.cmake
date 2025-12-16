set(||.UName||_COMPONENT_DEPS_core)
set(||.UName||_COMPONENT_OUTGOING_core)
list(LENGTH ||.UName||_COMPONENT_DEPS_core ||.UName||_COMPONENTS_DEPS_IND_core)
list(APPEND ||.UName||_AVAILABLE_COMPONENTS "core")

macro(_||.LName||_get_core_external_deps)
    find_dependency(ZLIB)
endmacro()