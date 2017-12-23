#!/bin/bash  

echo "===> Removing old DB..."
rm launchpad.db*

echo "===> Copying new DB..."
cp $TMPDIR../0/com.apple.dock.launchpad/db/db ./launchpad.db