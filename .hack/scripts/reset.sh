#!/bin/bash

echo "===> Removing all old DB files..."
rm $HOME/Library/Application\ Support/Dock/*.db
rm -rf $TMPDIR../0/com.apple.dock.launchpad/db
rm -rf $TMPDIR../0/com.apple.dock.launchpad/db.corrupt

echo "===> Resetting LaunchPad..."
defaults write com.apple.dock ResetLaunchPad -bool true

echo "===> Killing Dock process..."
killall Dock
