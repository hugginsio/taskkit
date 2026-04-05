# TaskKit

A toolkit for task and project management. It supports a variety of task metadata, such as projects, tags, scheduling, dependencies, and dynamic urgency.

TaskKit is both a command line tool and a client library. In fact, the CLI was built using the same client that you can import into your Go programs: `go get github.com/hugginsio/taskkit/client`.

## Installation

While in v0, `taskkit` must be installed using the Go toolchain:

```sh
go install github.com/hugginsio/taskkit/cmd/taskkit@latest
```

## Prior Art

There are a lot of wonderful tools for tracking projects and tasks - I've used quite a number of them myself. These are the ones that have helped me the most:

- [@zk-org/zk](https://zk-org.github.io/zk/index.html)
- [Org Mode](https://orgmode.org)
- [Taskwarrior TUI](https://github.com/kdheepak/taskwarrior-tui)
- [Taskwarrior](https://taskwarrior.org)
- [Timewarrior](https://timewarrior.net)
