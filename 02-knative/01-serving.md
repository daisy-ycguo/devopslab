# 流量管控

## 1.检查service和revision
完成了[01-tekton/exercise-1](../01-tekton/01-exercise-1.md),[01-tekton/02-exercise-2.md](../01-tekton/exercise-2.md)和[02-knative/00-eventing](./00-eventing.md)后， 我们已经有三个service revision版本。在本实验中我们只关心前两个版本, GENERATION 1 和 2。
```
$ kn service list
NAME    URL                                                                      LATEST        AGE     CONDITIONS   READY   REASON
hello   http://hello-default.capacity-demo.us-south.containers.appdomain.cloud   hello-jhvvz   2m33s   3 OK / 3     True    

$ kn revision list
NAME                    SERVICE         GENERATION   AGE     CONDITIONS   READY   REASON
hello-6dzrh             hello           3            42s     4 OK / 4     True
hello-b527t             hello           2            9m10s   3 OK / 4     True
hello-95kvq             hello           1            26m     3 OK / 4     True
```

## 2. 给两个revision版本添加tag
把下面命令中的hello-jhvvz替换为您的GENERATION为2的revision，给它打上tag`version2`
```
$ kn service update hello --tag hello-b527t=version2
Updating Servi ce 'hello' in namespace 'default':

  0.164s The Route is still working to reflect the latest desired specification.
  0.407s Ingress has not yet been reconciled.
  0.666s Waiting for VirtualService to be ready
  1.918s Ready to serve.

Service 'hello' updated with latest revision 'hello-b527t' (unchanged) and URL:
http://hello-default.capacity-demo.us-south.containers.appdomain.cloud
```
把下面命令中的hello-xfljl替换为您的GENERATION为1的revision，给它打上tag`version1`
```
$ kn service update hello --tag hello-95kvq=version1
Updating Service 'hello' in namespace 'default':

  0.090s The Route is still working to reflect the latest desired specification.
  0.351s Ingress has not yet been reconciled.
  0.709s Waiting for VirtualService to be ready
  1.804s Ready to serve.

Service 'hello' updated with latest revision 'hello-95kvq' (unchanged) and URL:
http://hello-default.capacity-demo.us-south.containers.appdomain.cloud
```
## 2.让两个版本各分50%的流量

```
$ kn service update hello --traffic version1=50 --traffic version2=50
Updating Service 'hello' in namespace 'default':

  0.088s The Route is still working to reflect the latest desired specification.
  0.160s Ingress has not yet been reconciled.
  0.289s Waiting for VirtualService to be ready
  1.623s Ready to serve.

Service 'hello' updated with latest revision 'hello-jhvvz' (unchanged) and URL:
http://hello-default.capacity-demo.us-south.containers.appdomain.cloud
```

## 4.验证流量管控
访问应用两个版本各处理50%的请求。
```
$ for i in {1..50}; do curl http://hello-default.$INGRESS; doneHello world, this is BLUE-IBM!!!
Hello world, this is BLUE-IBM!!!
Hello world, this is BLUE-IBM!!!
Hello world, this is BLUE-IBM!!!
Hello world, this is BLUE-IBM!!!
Hello world, this is GREEN-IBM!!!
Hello world, this is BLUE-IBM!!!
Hello world, this is GREEN-IBM!!!
Hello world, this is GREEN-IBM!!!
Hello world, this is GREEN-IBM!!!
Hello world, this is GREEN-IBM!!!
Hello world, this is GREEN-IBM!!!
Hello world, this is GREEN-IBM!!!

...
```
Route中显示了traffic的分布。
```
$ kn route list
NAME    URL                                                                      READY
hello   http://hello-default.capacity-demo.us-south.containers.appdomain.cloud   True

$ kn route describe hello
...
  traffic:
  - latestRevision: false
    percent: 50
    revisionName: hello-jhvvz
    tag: version2
    url: http://version2-hello-default.capacity-demo.us-south.containers.appdomain.cloud
  - latestRevision: false
    percent: 50
    revisionName: hello-xfljl
    tag: version1
    url: http://version1-hello-default.capacity-demo.us-south.containers.appdomain.cloud
  url: http://hello-default.capacity-demo.us-south.containers.appdomain.cloud
```

## 5.把100%的请求路由到版本2.0
```
$ kn service update hello --traffic version2=100
Updating Service 'hello' in namespace 'default':

  0.046s The Route is still working to reflect the latest desired specification.
  0.178s Ingress has not yet been reconciled.
  0.474s Waiting for VirtualService to be ready
  1.582s Ready to serve.

Service 'hello' updated with latest revision 'hello-jhvvz' (unchanged) and URL:
http://hello-default.capacity-demo.us-south.containers.appdomain.cloud

$ kn service describe hello
Name:       hello
Namespace:  default
Age:        17m
URL:        http://hello-default.capacity-demo.us-south.containers.appdomain.cloud
Address:    http://hello.default.svc.cluster.local

Revisions:  
  100%  hello-jhvvz (current @latest) #version2 [2] (16m)
        Image:  us.icr.io/liqiujie/hello:2.0 (at 342b4c)
     +  hello-xfljl #version1 [1] (17m)
        Image:  us.icr.io/liqiujie/hello:1.0 (at 740f26)

Conditions:  
  OK TYPE                   AGE REASON
  ++ Ready                  15s 
  ++ ConfigurationsReady    15m 
  ++ RoutesReady            15s 
```

## 6.验证所有流量都被路由到版本2.0
```
$ for i in {1..50}; do curl http://hello-default.$INGRESS; done
Hello world, this is BLUE-IBM!!!
Hello world, this is BLUE-IBM!!!
Hello world, this is BLUE-IBM!!!
Hello world, this is BLUE-IBM!!!
Hello world, this is BLUE-IBM!!!
Hello world, this is BLUE-IBM!!!
Hello world, this is BLUE-IBM!!!
Hello world, this is BLUE-IBM!!!
```

恭喜您，您已经完成了Knative Serving的实验。下面进行[监控服务](./02-monitoring.md)
