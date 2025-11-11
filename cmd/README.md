# `cmexl` CLI

Bootstrap projects, inspect presets/triplets, and schedule parallel runs.

## `cmexl init`

Bootstrap a new **cmexl** project from a template and create a local config.

**Usage**

```
cmexl init --name <name> --short-name <id> --template <tpl> [options]
```

**Flags**

| Flag                    | Description                                                                   |
| ----------------------- | ----------------------------------------------------------------------------- |
| `--name <string>`       | **(required)** Formal project name (used in folders, `project()` name, etc.). |
| `--short-name <string>` | **(required)** Identifier (`lower-case-no-spaces`) for targets/binaries.      |
| `--template <string>`   | **(required)** Template folder inside `templates/` (auto-completed).          |
| `--no-vcpkg`            | Skip vcpkg integration.                                                       |
| `--add-support <list>`  | Extra toolchains: `mingw64`, `clang` (comma-separated).                       |
| `--add-platform <list>` | Target OS list: `linux`, `windows` (comma-separated).                         |
| `--version <semver>`    | Initial version (default `0.1.0.0`). Produces `version.h` / `VersionInfo.rc`. |
| `--add-configs <list>`  | Append configs to defaults (`Debug`, `Release`, …).                           |
| `--configs <list>`      | **Replace** config list (mutually exclusive with `--add-configs`).            |

**Examples**

```bash
# Plain C++ with defaults
cmexl init --name Hello --short-name hello --template cpp-console

# Windows-only, no vcpkg, extra configs
cmexl init --name Viewer --short-name viewer --template cpp-glfw \
           --no-vcpkg --add-platform windows --add-configs Dev,CI,Ship

# Cross-platform with clang & mingw64
cmexl init --name Utils --short-name utils --template cpp-lib \
           --add-platform linux,windows --add-support clang,mingw64
```

---

## `cmexl list`

List project scaffolding info from the working directory.

**Usage**

```
cmexl list presets [configure|build|test|package|workflow] [flags]
cmexl list triplets [--targets=windows,linux,macos]
```

> With no subcommand, `cmexl list` behaves like `cmexl list presets`.

### Subcommand: `presets`

Show every CMake preset, grouped by category.

**Flags**

| Flag               | Description                                                  |
| ------------------ | ------------------------------------------------------------ |
| `-n, --names-only` | Print only preset names and type—no quotes, no descriptions. |
| `-h, --help`       | Show built-in help.                                          |

**Examples**

```bash
cmexl list presets            # or: cmexl list
cmexl list presets build -n
```

### Subcommand: `triplets`

List vcpkg triplets visible for specific platforms.

**Usage**

```
cmexl list triplets [--targets=windows,linux,macos]
```

**Flags**

| Flag               | Description                                                                    |
| ------------------ | ------------------------------------------------------------------------------ |
| `--targets=<list>` | Comma-separated platforms to filter (`windows`,`linux`,`macos`). Default: all. |
| `-h, --help`       | Show built-in help.                                                            |

**Examples**

```bash
cmexl list triplets
cmexl list triplets --targets=windows,linux
```

**Tip**
For vcpkg basics on triplets, see `vcpkg help triplet`.

---

## `cmexl schedule`

Schedule a group of CMake presets for **parallel** execution and stream a live progress feed (status, logs, exit codes).

**Usage**

```
cmexl schedule [--group <group-name>] [--type=<preset-type>] [<preset>...]
```

> If neither `--group` nor explicit `<preset>` names are provided, the default preset group is used.

**Flags**

| Flag                   | Description                                                  |
| ---------------------- | ------------------------------------------------------------ |
| `--group <name>`       | Predefined preset group to schedule.                         |
| `--type <preset-type>` | One of: `configure`, `build`, `test`, `package`, `workflow`. |
| `--no-stream`          | Don’t stream live logs; print a summary at the end.          |
| `--fail-fast`          | Stop remaining runs when any preset fails.                   |
| `-h, --help`           | Show built-in help.                                          |

**Examples**

```bash
cmexl schedule --group msvc-all-configs
cmexl schedule --type workflow win-dev win-rel
cmexl schedule --group ci --no-stream
```
