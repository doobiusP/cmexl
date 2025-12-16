cpack_add_component(
    "math"
    DISPLAY_NAME 
        "Math"
    DESCRIPTION
        "Math component with build artifacts, headers and pdbs"
    DEPENDS
        "export" "core"  
    GROUP
        "SDK"
    INSTALL_TYPES
        Full Default
)