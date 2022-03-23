#!/bin/sh
HASHFILE=hashes.yml
BASE=../internal/engine/engineConfigs
CURDIR=$( pwd )

cd $BASE
CONFIGS=$( ls *.yml | paste -s - )
cd $CURDIR
rm -f "$HASHFILE"
for CFG in $CONFIGS
do
	if [ $CFG != '.' ];
	then
		MD5_DIGEST=$( md5sum "$BASE/$CFG" | awk '{print $1}' )
		SHA1_DIGEST=$( sha1sum "$BASE/$CFG" | awk '{print $1}' )
		echo "$CFG:" >> "$HASHFILE"
		echo "  md5digest:  $MD5_DIGEST" >> "$HASHFILE"
		echo "  sha1digest: $SHA1_DIGEST" >> "$HASHFILE"
	fi
done