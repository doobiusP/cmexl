set(CPACK_PACKAGE_NAME ||.Name||)
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "||.Name|| short description")
set(CPACK_PACKAGE_INSTALL_DIRECTORY ${CPACK_PACKAGE_NAME})
set(CPACK_VERBATIM_VARIABLES TRUE)
set(CPACK_PACKAGE_DESCRIPTION "||.Name|| long description") # Alternatively use CPACK_PACKAGE_DESCRIPTION_FILE with a file provided
set(CPACK_COMPONENTS_GROUPING ONE_PER_GROUP)

# Set the below with necessary files:
# set(CPACK_PACKAGE_ICON)
set(CPACK_RESOURCE_FILE_WELCOME ${CMAKE_CURRENT_LIST_DIR}/WELCOME.txt)
set(CPACK_RESOURCE_FILE_LICENSE ${CMAKE_SOURCE_DIR}/LICENSE.md)
set(CPACK_RESOURCE_FILE_README ${CMAKE_SOURCE_DIR}/README.md)

# Set the below with necessary files:
# set(CPACK_NSIS_MUI_ICON)   
# set(CPACK_NSIS_MUI_UNIICON)

set(CPACK_NSIS_DISPLAY_NAME "${LIBRARY_ALIAS} Suite")    
set(CPACK_NSIS_PACKAGE_NAME "${LIBRARY_ALIAS}")    
set(CPACK_NSIS_ENABLE_UNINSTALL_BEFORE_INSTALL TRUE)

# RPM-specific configuration
set(CPACK_RPM_COMPONENT_INSTALL TRUE)
set(CPACK_RPM_DEBUGINFO_PACKAGE TRUE)
if(CMAKE_SIZEOF_VOID_P EQUAL 4)
    set(CPACK_RPM_PACKAGE_ARCHITECTURE i686)
endif()

set(CPACK_DEB_COMPONENT_INSTALL TRUE)

include(CPack)

if(CPACK_GENERATOR MATCHES "^(7Z|TBZ2|TGZ|TXZ|TZ|ZIP)$")
    set(CPACK_ARCHIVE_COMPONENT_INSTALL TRUE)
    set(CPACK_COMPONENT_INCLUDE_TOPLEVEL_DIRECTORY TRUE)
endif()

cpack_add_install_type(Full)
cpack_add_install_type(Minimal)
cpack_add_install_type(Default)

cpack_add_component_group(
    "SDK"
    DESCRIPTION
        "All necessary files for library development"
    EXPANDED
)

# ====================================================================== #

cpack_add_component(
    "export"
    DISPLAY_NAME 
        "Export"
    DESCRIPTION
        "CMake export files"
    GROUP
        "SDK"
    INSTALL_TYPES
        Full Minimal Default
)

file(GLOB COMP_DEP_FILES "${CMAKE_CURRENT_LIST_DIR}/comps/*.cmake")

foreach(COMP_DEP IN LISTS COMP_DEP_FILES)
    message(VERBOSE "Found ${COMP_DEP}")
    include(${COMP_DEP})
endforeach()