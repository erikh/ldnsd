#!bash

printf 'package version //nolint\nconst Version = "%s"' "${VERSION}" > generated.go
