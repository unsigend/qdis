# Count Go implementation lines in a cloc-like table.
# Usage: awk -f scripts/line.awk file.go ...

BEGIN {
	nfiles = 0
	blank = 0
	comment = 0
	code = 0
}

FNR == 1 {
	nfiles++
	inblock = 0
}

{
	line = $0
	sub(/\r$/, "", line)
	if (inblock) {
		comment++
		if (index(line, "*/") > 0)
			inblock = 0
		next
	}
	if (line ~ /^[ \t]*$/) {
		blank++
		next
	}
	tmp = line
	sub(/^[ \t]+/, "", tmp)
	if (tmp ~ /^\/\*/) {
		comment++
		if (tmp !~ /\*\//)
			inblock = 1
		next
	}
	if (tmp ~ /^\/\//) {
		comment++
		next
	}
	code++
}

END {
	bar = "-------------------------------------------------------------------------------"
	printf "%s\n", bar
	printf "%-27s %11s %13s %13s %11s\n", "Language", "files", "blank", "comment", "code"
	printf "%s\n", bar
	printf "%-27s %11d %13d %13d %11d\n", "Go", nfiles, blank, comment, code
	printf "%s\n", bar
}
