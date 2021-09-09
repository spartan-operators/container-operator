## Container Operator Commands

> Domain specific information
```
Domain: vnv.wrstudio.cloud
Repository: github.com/spartan-operators/container-operator
Kind: ContainerJob
Type: Cluster
```

> Creating a new Operator SDK project
```
operator-sdk init --domain vnv.wrstudio.cloud --repo github.com/example/memcached-operator
```

> Creating the API and Controller
```
operator-sdk create api --group container --version v1alpha1 --kind ContainerJob --resource --controller
```
