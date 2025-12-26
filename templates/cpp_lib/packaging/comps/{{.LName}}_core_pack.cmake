cpack_add_component(
    "core"
    DISPLAY_NAME 
        "Core"
    DESCRIPTION
        "Core component with build artifacts, headers and pdbs"
    DEPENDS
        "export"
    GROUP
        "SDK"
    INSTALL_TYPES
        Full Minimal Default
)