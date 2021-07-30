#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

go build -o "${SCRIPT_DIR}/update-gradle-smp-plugin" "${SCRIPT_DIR}/update-gradle-smp-plugin.go"
