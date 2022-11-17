#!/usr/bin/env sh

for i in $@; do
	vnstati --style 1 -L -s -o ${i}-s.png $i
	vnstati --style 1 -L --fivegraph 576 218 -o ${i}-5g.png $i
	vnstati --style 1 -L -hg -o ${i}-hg.png $i
	vnstati --style 1 -L -h 24 -o ${i}-h.png $i
	vnstati --style 1 -L -d 30 -o ${i}-d.png $i
	vnstati --style 1 -L -t 10 -o ${i}-t.png $i
	vnstati --style 1 -L -m 12 -o ${i}-m.png $i
	vnstati --style 1 -L -y 5 -o ${i}-y.png $i
done
