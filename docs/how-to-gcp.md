This document explains how to optimize gcp compute instances

## 1. Have the following Software Installed

1.  Kaytu CLI - refer quick start [here](../README.md#quick-start---optimize-aws-ec2-rbs-and-rds-workload)
2.  Gcloud CLI - refer [official gcloud cli documentaion](https://cloud.google.com/sdk/docs/install)

## 2. Login to kaytu CLI

```
kaytu login
```

## 3. Install CLI plugin:

```
kaytu plugin install gcp
```

## 4. Create a Service Account

Create a service account with the following roles:
1. Compute Engine 
2. Cloud Monitoring

<details>
<summary>Create service account using Gcloud CLI</summary>

### 1. Create the service account 

```
gcloud iam service-accounts create kaytu-sa \
  --description="Service account for use with Kaytu CLI" \
  --display-name="Kaytu Service Account"
```

### 2. Add required roles to the service account

```
gcloud projects add-iam-policy-binding kaytu-428817 \
  --member="serviceAccount:kaytu-sa@kaytu-428817.iam.gserviceaccount.com" \
  --role="roles/monitoring.viewer"

gcloud projects add-iam-policy-binding kaytu-428817 \
  --member="serviceAccount:kaytu-sa@kaytu-428817.iam.gserviceaccount.com" \
  --role="roles/compute.viewer"
```

</details>


## 5. Create Key and store credentials file

> **_NOTE:_** Make sure you have permissions to create service account keys

```
gcloud iam service-accounts keys create credentials.json --iam-account=kaytu-sa@kaytu-428817.iam.gserviceaccount.com
```


## 6. Export credentials file path

Export the above created credentials file path in the environment variable

```
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
```

## 7. Run optimization Check

```
kaytu optimize compute-instance
```
