## Container Operator Commands

> Domain specific information
```text
Domain: vnv.wrstudio.cloud
Repository: github.com/spartan-operators/container-operator
Kind: ContainerJob
Type: Cluster
```

> Creating a new Operator SDK project
```shell
operator-sdk init --domain vnv.wrstudio.cloud --repo github.com/example/memcached-operator
```

> Creating the API and Controller
```shell
operator-sdk create api --group container --version v1alpha1 --kind ContainerJob --resource --controller
```

> Push Images to ECR
```shell
# Replace <docker_images> and <tag> with real values
aws ecr get-login-password --profile wrdev --region us-east-1 | docker login --username AWS --password-stdin 945377166858.dkr.ecr.us-east-1.amazonaws.com
docker tag <docker_image> 945377166858.dkr.ecr.us-east-1.amazonaws.com/automation-sla-monitoring-develop:<tag>
docker push 945377166858.dkr.ecr.us-east-1.amazonaws.com/automation-sla-monitoring-develop:<tag>
```

> Retrieve kubeconfig from EKS cluster
```shell
aws eks --profile wrdev --region us-east-1 update-kubeconfig --name k8s-wr-vv
```