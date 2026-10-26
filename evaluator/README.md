# Archived evaluator

This directory contains the old tree-walking evaluator. It is **not built,
tested, shipped, or supported** by default.

The package is behind the `legacy_eval` build tag and is retained for one
release as a reference. New code must target the VM.

To inspect the archive locally:

```sh
go test -tags legacy_eval ./evaluator
```

This package will be deleted in the next release.
