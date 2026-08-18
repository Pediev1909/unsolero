#!/bin/sh
set -eu

if [ "$#" -eq 0 ]; then
  echo "usage: $0 DOCKERFILE..." >&2
  exit 2
fi

awk '
  toupper($1) == "FROM" {
    image_index = 2
    if ($2 ~ /^--platform=/) {
      image_index = 3
    }
    image = $image_index
    if (!(image in stages) && image !~ /@sha256:[0-9a-f]{64}$/) {
      printf "%s:%d: external base image is not digest pinned: %s\n", FILENAME, FNR, image > "/dev/stderr"
      failed = 1
    }
    if (toupper($(image_index + 1)) == "AS") {
      stages[$(image_index + 2)] = 1
    }
  }
	$1 == "image:" {
	  image = $2
	  gsub(/["\047]/, "", image)
	  if (image !~ /@sha256:[0-9a-f]{64}$/) {
	    printf "%s:%d: compose image is not digest pinned: %s\n", FILENAME, FNR, image > "/dev/stderr"
	    failed = 1
	  }
	}
  END { exit failed }
' "$@"
