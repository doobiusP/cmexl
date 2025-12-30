# cmexl

## A Lightweight CMake/C++ Project Tool

cmexl is an opiniated CMake/C++ project bootstrapper and batch-builder

It provides the following capabilities:

* CMake/C++
  * Bootstrap a new project based on parametrised templates
  * Default templates come with a set of toolchains out-of-the-box for cross-compiling between Windows and Linux w/ different compilers for common use-cases
  * Sets up sane-default presets that target 95% use-cases across the major
    desktop platforms via different toolchains
  * Provides single-command serial batch-build execution for multiple workflow presets
* vcpkg
  * Automates deployment of the vcpkg port for libraries
  * Provides component dependency inclusion for your libraries to automatically bundle the minimal set of dependencies when an end-user requests for a feature

## 1. Installation

### Prerequisites

You will need

* Go (1.24+)
* CMake (3.31+)
* vcpkg

1. Install **Go**:
    * Visit the official [Go Downloads page](https://go.dev/dl/)

2. Install **CMake**:
    * Follow the instructions at [https://cmake.org/download](https://cmake.org/download)

3. Install **vcpkg**:  
   * Follow the instructions at [Microsoft's vcpkg tutorial](https://learn.microsoft.com/en-us/vcpkg/get_started/get-started?pivots=shell-powershell)
   * Set the environment variable `VCPKG_ROOT` to your vcpkg installation directory

### Build from Source

Once Go is installed and configured, you can build the cmexl executable:

1. Clone the repo

    ```bash
    git clone https://github.com/doobiusP/cmexl.git
    cd cmexl
    ```

2. Build the executable

    ```bash
    go build
    ```

3. (Optional) Install the executable

    ```bash
    go install
    ```

## 2. Quick Start

Get a new CMake/C++ project running and test its cross-compilation support in a few steps.

1. **Initialize a new project**
    Use `cmexl init` to bootstrap a new CMake project with automatic vcpkg integration

    ```bash
    cmexl init --name "MyLib1" --template "cpp_lib"
    cd MyLib1
    ./cmexl_bootstrap.bat
    ```

2. **Run a batch build**
    Use `cmexl build` to execute a group of CMake workflows, which will automatically fetch dependencies, configure, build and test the project

    ```bash
    cmexl build -p "msvc-dev,msvc-rel"
    ```

    or with the use of **tasks** (collection of workflow presets written into `cmexlconf.json`, see `cmexlconf_schema.json` for more details)

    ```bash
    cmexl build -t "win"
    ```  

## Templates

To create your own template, simply create a directory under `cmexl/templates`!

cmexl then uses the following variables that are to be replaced with using the Go templating engine:

```json
{
    "Name": "formal name of your project, i.e. FastGAME",
    "SName": "short name of your project, i.e FGame",
    "UName": "SName but uppercase, i.e. FGAME",
    "LName": "SName but lowercase, i.e. fgame"
}
```

To replace inside file contents, do `||.Var||`

To replace inside filenames and directory names, do `{{.Var}}`

## 3. Who should use cmexl?

If you want to start a **new** CMake/C++ project but don't want to go through the hassle
of hand-rolling your own CMake and dependency management, then cmexl can help.

But as it stands, cmexl is only useful if you plan on targeting multiple major desktop platforms (Linux, Windows)(macOS support is non-existent right now). The range of compiler support
is also somewhat lacking with more thorough testing required for Linux outside of just
GCC.

You will not benefit much from using CMake (and cmexl by extension) if:

* If you only plan on targeting a single platform, it is far better to use the standard
build system/IDE for that platform: Visual Studio for Windows, Makefile for Linux, XCode for
macOS, etc.
* If your library is focused enough that it can fit in a few files, then just ask users to copy
the files into their projects. This is in fact the most portable solution.
* If you want to create a header-only library. While CMake has support for header-only
libraries, currently cmexl does not offer a template for it (this may change in the future)

## 4. How does cmexl structure a project?

Note that the default templates are designed to be edited however you like once instantiated. There is no hard restriction imposed anywhere with these templates but sane defaults have been chosen.

* cpp_lib
  * Makes no reference to "cmexl" anywhere in the project; cmexl is out of the picture once templated
  * The library is made up of components and all source files, header or implementation,
    exist in some component
  * Each component is its own CMake target which has a set of immediate internal and external
    dependencies
  * There is a 1-1 mapping between vcpkg features and cmake components so that each component
    can be consumed independently without bringing in the whole project. Dependencies (internal) are automatically resolved by generating the dependency graph on your behalf
  * All targets privately link to a common build setting INTERFACE target that do not expose macros unless linked to publicly

## 5. Future Work

1. Provide automated library extension mechanisms for `cpp_lib` through python scripts
2. Provide the `cpp_hlib` header-only library template
3. Introduce MacOS support
