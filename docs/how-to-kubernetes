This document explains how to optimize Kubernetes clusters

## 1. Have the following Software Installed

    

1.  Kaytu CLI
2.  Kubectl configured
    

## 2. Login to kaytu CLI
`kaytu login`

Install CLI plugin:

    kaytu plugin install kubernetes

## 3. (Optional) Login to Cloud Provider

AWS, Azure, Google will generate Kubeconfig from their 

## 4. Authenticate to Kubernetes Cluster

Ensure Kubectl configured and connected to Kubernetes Cluster
    
## 5. Add the helm Repo

   
    helm repo add kaytu-io [https://kaytu-io.github.io/kaytu-charts](https://kaytu-io.github.io/kaytu-charts)


## 6. Generate API Key
    
Generate API Key to enroll the Cluster by running the following command:

    kaytu apikey generate <key-name>  
    
Copy the generated key and use it as the kaytu.kaytu.authToken value.

## 7. Install the Chart

1.  Get the values file for the chart: helm show values kaytu-io/kaytu-agent > kaytu-agent-values.yaml
    
2.  Edit the values file to suit your needs. be sure to set the kaytu.kaytu.authToken and verify the kaytu.kaytu.prometheus.address value.
    
3.  Install the kaytu-agent chart: helm install --create-namespace -n kaytu-system my-kaytu-agent kaytu-io/kaytu-agent -f kaytu-agent-values.yaml


5. Run optimization (for both agent and non-agent)

## 7. Run optimization Checks

> kaytu optimize kubernetes-pods  
> kaytu optimize kubernetes-deployments
> kaytu optimize kubernetes-statefulsets
> kaytu optimize kubernetes-daemonsets
> kaytu optimize kubernetes-jobs
