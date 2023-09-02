#!/bin/bash  

echo "===> Removing old DB..."
rm launchpad.db*

echo "===> Showing DB files..."
tree $TMPDIR../0/com.apple.dock.launchpad

echo "===> Copying new DB..."
cp $TMPDIR../0/com.apple.dock.launchpad/db/db ./launchpad.db