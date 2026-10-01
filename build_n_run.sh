#!/usr/bin/env bash
set -e
cd "$(dirname "$0")"

go build -tags ebitensinglethread,ebiten,terminal -o ./contractor .
./contractor val_dialogue >/dev/null 2>&1 && ./contractor val_ammo >/dev/null 2>&1 ||
    echo "WARNING: content validation failed, run './contractor val_dialogue' and './contractor val_ammo'" >&2
exec ./contractor "$@"
