#!/bin/bash
go work init
for file in `find . -name go.mod`; do
    dirpath=$(dirname $file)
    echo $dirpath
    go work use $dirpath
done