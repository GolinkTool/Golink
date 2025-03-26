# Use Go-Mod to migrate Example

## Description

While migrating the example project from GoPath to Go Modules, we found that the build process fails **due to mismatched or unresolved dependencies.**

The following error log was produced during the build process:

```text
go: finding module for package github.com/ethereum/go-ethereum/crypto/sha3
go: downloading github.com/ethereum/go-ethereum v1.15.5
go: Example/Go_Example2 imports
        github.com/ethereum/go-ethereum/crypto/sha3: module github.com/ethereum/go-ethereum@latest found (v1.15.5), but does not contain package github.com/ethereum/go-ethereum/crypto/sha3
```
## Result

The build fails with errors related to missing or mismatched dependencies.

The error dependency is `github.com/ethereum/go-ethereum`.

## Reason

This issue appears to be caused by the absence of precise version tracking in GOPATH, which leads to inconsistency in dependency resolution.
