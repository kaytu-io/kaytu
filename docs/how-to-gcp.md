This document explains how to optimize gcp compute instances

## 1. Have the following Software Installed

1.  Kaytu CLI


## 2. Login to kaytu CLI
`kaytu login`

Install CLI plugin:

    kaytu plugin install gcp


## 4. Create a Service Account

Create a service account with the following roles:
1. Cloud Billing
2. Compute Engine 

## 5. Create Key and store credentials file

## 6. Export credentials file path

Export credentials file path in the environment variable

    export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"

## 7. Run optimization Check

> kaytu optimize compute-instance
