# `cmexl` CLI

Bootstrap projects, inspect presets/triplets, and schedule parallel runs.

## `cmexl init`

Bootstrap a new **cmexl** project from a template

**Usage**

```
cmexl init --name <name> --short-name <id> --template <tpl>
```

**Flags**

| Flag                    | Description                                                                   |
| ----------------------- | ----------------------------------------------------------------------------- |
| `--name <string>`       | **(required)** Formal project name (used in folders, `project()` name, etc.). |
| `--short-name <string>` | **(required)** Identifier (`lower-case-no-spaces`) for targets/binaries.      |
| `--template <string>`   | **(required)** Template folder inside `templates/`.          |

**Examples**

```bash
cmexl init --name MyLib1 --template cmake_lib

cmexl init --name CrazyEngine --short-name CzEng --template cmake_app
```

## `cmexl list`

List cmke presets info from the working directory into json format

**Usage**

```
cmexl list [configure|build|test|package|workflow] [flags]
```

> With no subcommand, `cmexl list` behaves like `cmexl list <all-preset-types>`.

**Flags**

| Flag               | Description                                                  |
| ------------------ | ------------------------------------------------------------ |
| `-n, --names-only` | Print only preset names and type—no quotes, no descriptions. |

**Examples**

```bash
cmexl list
cmexl list build -n
```

## `cmexl schedule`

Schedule CMake presets for parallel or serial execution with live progress tracking.

**Usage**

```bash
# Execute presets by type (mandatory if not using task)
cmexl schedule -t <type> <preset-names...> [flags]

# Execute a predefined task from cmexlconf.json
cmexl schedule task <task-name> [flags]
```

**Flags**

| Flag | Description |
| --- | --- |
| `-t, --type <string>` | **Required (non-task).** One of: `configure`, `build`, `test`, `package`, `workflow`. |
| `-s, --serial` | Force serial execution. Recommended if the underlying build system lacks parallel support. |
| `--save-events` | Persist logs to `.cmexl/events/{presetName}.log`. |
| `-h, --help` | Show built-in help. |

**Examples**

```bash
# Run multiple build presets in parallel
cmexl schedule -t build win-dev linux-dev

# Run a specific task defined in cmexlconf.json
cmexl schedule task ci-pipeline

# Force serial execution for stability
cmexl schedule -t configure debug-base release-base --serial
```