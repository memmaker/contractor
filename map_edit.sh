#!/bin/sh
# Build and start the map editor (../contractor_map_edit) on this game's data. Extra args go to the editor, e.g. ./map_edit.sh 120x40
cd "$(dirname "$0")" || exit 1
ED=../contractor_map_edit
(cd "$ED" && go build -o mapedit .) || exit 1
exec "$ED/mapedit" data="$PWD/data_atom" "$@"
