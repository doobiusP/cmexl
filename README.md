# cmexl

### A Lightweight CMake/C++ Project Tool

cmexl is a CMake/C++ project bootstrapper and parallel build orchestrator

It provides the following capabilities:
* CMake/C++
    * Bootstrap a new project based on parametrised templates
    * Comes with a set of toolchains out-of-the-box for cross-compiling between Windows and Linux w/ different compilers for common use-cases
    * Sets up sane-default presets that target 95% use-cases across different toolchains 
    * Provides single-command parallel build execution for multiple presets
* vcpkg
    * Automates deployment of the vcpkg port for CMake-based libraries
    * Minimizes vcpkg debugging when installs fail by fetching the necessary logs
    * Simplifies vcpkg output

## 1. Installation

### Prerequisites

You will need
* Go (1.16+)
* CMake (3.31+)
* vcpkg (optional)

1. Install **Go**: 
    * Visit the official [Go Downloads page](https://go.dev/dl/)

2. Install **CMake**: 
    * Follow the instructions at [https://cmake.org/download](https://cmake.org/download)

3. Install **vcpkg** (optional):  
   * Follow the instructions at [Microsoft's vcpkg tutorial](https://learn.microsoft.com/en-us/vcpkg/get_started/get-started?pivots=shell-powershell)
   * Set the environment variable `VCPKG_ROOT` to your vcpkg installation directory
   

### Build from Source

Once Go is installed and configured, you can build the cmexl executable:

1.  Clone the repo

    ```bash
    git clone https://github.com/doobiusP/cmexl.git
    cd cmexl
    ```

2.  Build the executable
    ```bash
    go build
    ```

3.  (Optional) Install the executable
    ```bash
    go install
    ```

## 2. Quick Start

Get a new C++/CMake project running and test its cross-compilation support in a few steps.

1.  **Initialize a new project**
    Use `cmexl init` to bootstrap a new project from a template. This command automatically sets up CMake and vcpkg integration.

    ```bash
    # Creates a new project: Formal name "MyGreatLibrary", identifier "my-great-lib", using the 'default-tpl' template.
    cmexl init --name "My Great Library" --short-name my-great-lib --template default-tpl
    ```

2.  **Run a parallel build and test schedule**
    Use `cmexl schedule` to execute a group of CMake presets in parallel, which will automatically fetch dependencies and run cross-compilation tests.

    ```bash
    # Build all presets defined in the 'msvc-all-configs' group
    cmexl schedule --group msvc-all-configs --type build

    # Run the tests for the default preset group
    cmexl schedule --type test
    ```

## 3. Work in progress

cmexl is a work in progress with the following features yet to be implemented:

* Templates
    * Fully self-contained example project template
    * Plugin/Sub-process support for project templates that need stronger project setup
* vcpkg
    * Automatic vcpkg port management
    * vcpkg error log fetching
    * Minimizing vcpkg contention during parallel configure steps by scheduling vcpkg only when needed
* CMake
    * Saner defaults for toolchains
    * Cleaner and more modular CMake for the default library template
    * Concept of groups for parallel scheduling
    * Local configuration for cmexl to read from
* Cmexl
    * Cleaner GO
    * Complete documentation
    * General edge-case debugging