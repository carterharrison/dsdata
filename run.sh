#!/bin/bash

NUM=0

for FILENAME in ./temp/ftp.ngs.noaa.gov/pub/DS_ARCHIVE/DataSheets/*.txt; do
  ./dsdata "$FILENAME"
#	GO=$(./dsdata "$FILENAME" | wc -l)
#	NUM="$((NUM+GO))"
#	echo $NUM
done
