#!/bin/bash

apt update
apt install -y wget curl

DEB_FILE_URL=$(curl -s https://api.github.com/repos/kaytu-io/kaytu/releases/latest \
| grep "browser_download_url.*amd64.deb" \
| cut -d : -f 2,3 \
| tr -d \" )

echo $DEB_FILE_URL

wget -q $DEB_FILE_URL

FILENAME=$(basename "$DEB_FILE_URL")

apt install ./$FILENAME

