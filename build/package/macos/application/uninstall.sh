#!/bin/bash

#Check running user
if (( $EUID != 0 )); then
    echo "Please run as root."
    exit
fi

echo "Welcome to the Weep Uninstaller"
echo "The following packages will be REMOVED:"
echo "  weep"
while true; do
    read -p "Do you wish to continue [Y/n]?" answer
    [[ $answer == "y" || $answer == "Y" || $answer == "" ]] && break
    [[ $answer == "n" || $answer == "N" ]] && exit 0
    echo "Please answer with 'y' or 'n'"
done


echo "Uninstalling Weep"
# remove binary symlink
if rm -rf "/usr/local/bin/weep"
then
  echo "[1/3] [DONE] Successfully deleted shortcut links"
else
  echo "[1/3] [ERROR] Could not delete shortcut links" >&2
fi

#forget from pkgutil
if pkgutil --forget "com.netflix.weep" > /dev/null 2>&1
then
  echo "[2/3] [DONE] Successfully deleted application information"
else
  echo "[2/3] [ERROR] Could not delete application information" >&2
fi

#remove application source distribution
if rm -rf "/Library/weep"
then
  echo "[3/3] [DONE] Successfully deleted application"
else
  echo "[3/3] [ERROR] Could not delete application" >&2
fi

echo "Application uninstall process finished"
exit 0
