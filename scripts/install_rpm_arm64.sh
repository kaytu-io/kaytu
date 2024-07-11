#!/bin/bash

yum install -y wget curl jq

RPM_FILE_URL=$(curl -s https://api.github.com/repos/kaytu-io/kaytu/releases/latest \
| jq -r '.assets[] | select(.name | endswith("arm64.rpm")).browser_download_url')

echo $RPM_FILE_URL

wget -q $RPM_FILE_URL

FILENAME=$(basename "$RPM_FILE_URL")

yum install -y $FILENAME
