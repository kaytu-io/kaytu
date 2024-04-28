
![Kaytu Gif](.github/assets/kaytu.gif)

## Overview

Kaytu CLI enables engineering, DevOps, and SRE teams to reduce cloud costs by recommending optimal workload configurations based on **actual-usage**, ensuring savings without compromise.

- **Ease of use**: One-line command. Use without modifying workloads or making configuration changes.
- **Base on actual Usage**: Analyzes the past seven days of usage from Cloud native monitoring (CloudWatch), including advanced AWS CloudWatch metrics (where available).
- **Customize**: Optimize for region, CPU, memory, network performance, storage, licenses, and more to match your specific requirements.
- **Secure** - no credentials to share; extracts required metrics from the client side
- **Open-core philosophy** Use without fear of lock-in. The CLI is open-sourced, and the Server side will be open-sourced soon.
- **Coming Soon**: Non-Interactive mode, Azure support, GPU Optimization, Credit utilization for Burst instances, and Observability data from Prometheus

## Getting Started

### 1. Install Kaytu CLI

**MacOS**
```shell
brew tap kaytu-io/cli-tap && brew install kaytu
```

**Windows w/Chocolatey**
```shell
choco install kaytu
```

**Linux**
```shell
curl -fsSL https://raw.githubusercontent.com/kaytu-io/kaytu/main/scripts/install.sh | sh
```

**Binary Download**

Download and install Windows, MacOS, and Linux binaries manually from [releases](https://github.com/kaytu-io/kaytu/releases) 

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
