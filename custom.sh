#!/usr/bin/env bash

set -xeuo pipefail

# DUCKTAPE_VERSION is set in the build container, unset otherwise
if [[ ! -v DUCKTAPE_VERSION ]] ; then
    exit 0
fi

pacman -Syu pcsclite
