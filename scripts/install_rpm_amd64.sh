#!/bin/bash

yum install -y wget curl

RPM_FILE_URL=$(curl -s https://api.github.com/repos/adorigi/kaytu/releases/latest \
| grep "browser_download_url.*amd64.rpm" \
| cut -d : -f 2,3 \
| tr -d \" )

echo $RPM_FILE_URL

wget -q $RPM_FILE_URL

FILENAME=$(basename "$RPM_FILE_URL")

yum install -y $FILENAME
