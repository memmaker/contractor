#!/bin/sh
# Regenerate one content pack and check it: lint, dialogue validation, map/chapter tests, the pack's playtests.
#   ./content_check.sh f            regenerate content/f.py and run everything
#   ./content_check.sh f ecto       ...but only playtests named f_ecto*
cd "$(dirname "$0")" || exit 1
pack=${1:?usage: ./content_check.sh <pack> [playtest filter]}
bin=${TMPDIR:-/tmp}/contractor_check
python3 "content/$pack.py" || exit 1
python3 tools/content.py lint "$pack" || exit 1
go build -o "$bin" . || exit 1
"$bin" val_dialogue 2>&1 | grep "ERR" && exit 1
go test ./game -run 'ChapterMaps|Chapters|StarterQuest' || exit 1
fail=0
for f in data_atom/playtests/${pack}_${2}*.rec; do
  r=$(timeout 180 "$bin" autoplay -headless "$(basename "$f" .rec)" 2>&1 | tail -1)
  echo "$r"
  case "$r" in PASS*) ;; *) fail=1 ;; esac
done
exit $fail
