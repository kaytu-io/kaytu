<p align="center">
<a href="https://www.kaytu.io"><img src=".github/assets/Kaytu-New-Logo.svg" alt="Kaytu Logo" width="300" /></a>
<p align="center">Kaytu recommends optimal workload configurations based on <b>actual-usage,</b>, ensuring savings without compromising reliability.
</p>

![Kaytu Gif](.github/assets/kaytu.gif)
## Overview
- **Ease of use**: One-line command. Use without modifying workloads or making configuration changes.
- **Optimize**: Optimize AWS EC2 Instances & AWS RDS Instances/Clusters.
- **Base on actual Usage**: Analyzes the past seven days of usage from Cloud native monitoring (CloudWatch).
- **Customize**: Optimize for region, CPU, memory, network performance, storage, licenses, and more to match your specific requirements.
- **Secure** - no credentials to share; extracts required metrics from the client side
- **Open philosophy** Use without fear of lock-in. The CLI is open-sourced, and the Server side will be open-sourced soon.
- **Coming Soon**: Non-Interactive mode, GCP, Azure, GPU Optimization and Observability data from Prometheus
## Getting Started
### 1. Install Kaytu CLI
**MacOS**
```shell
brew tap kaytu-io/cli-tap && brew install kaytu
```
**Linux**
```shell
curl -fsSL https://raw.githubusercontent.com/kaytu-io/kaytu/main/scripts/install.sh | sh
```
**Windows (and all Binaries)**
Download Windows (Linux, and MacOS) binary from [releases](https://github.com/kaytu-io/kaytu/releases) 
### 2. Login to AWS CLI
Kaytu works with your existing AWS CLI profile (read-only access required) to gather metrics.  
To confirm your AWS CLI login is working correctly:
```
aws sts get-caller-identity
```
[Click here to see how to log in to AWS CLI.](https://docs.aws.amazon.com/signin/latest/userguide/command-line-sign-in.html)
We respect your privacy. Our open-source code guarantees that we never collect sensitive information like AWS credentials, IPs, tags, etc.
### 3. Run Kaytu CLI
```shell
kaytu
```
This will run and install any plugins.
To see how you can optimize EC2 Instances, run this command:
```shell
kaytu optimize ec2-instance
```
Some optimizations such as RDS require login: 
```shell
kaytu login
```
For RDS:
```shell
kaytu optimize rds-instance
```
