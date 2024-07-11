#!/bin/bash

apt update
apt install -y wget curl jq

DEB_FILE_URL=$(curl -s https://api.github.com/repos/kaytu-io/kaytu/releases/latest \
| jq -r '.assets[] | select(.name | endswith("arm64.deb")).browser_download_url')

echo $DEB_FILE_URL

wget -q $DEB_FILE_URL

FILENAME=$(basename "$DEB_FILE_URL")

apt install ./$FILENAME

