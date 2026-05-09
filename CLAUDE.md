# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`svenska-nyheter-reader` is a Go project for reading Swedish news. The repository is in early/empty state — no source code exists yet.

## Language & Toolchain

This is a Go project (inferred from the `.gitignore`). Standard Go commands apply once source is added:

```bash
go build ./...       # build
go test ./...        # run all tests
go test ./pkg/... -run TestName  # run a single test
go vet ./...         # static analysis
```
