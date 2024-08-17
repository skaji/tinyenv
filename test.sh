#!/bin/bash
export TINYENV_ROOT=$PWD/_root
exec go run *.go "$@"
