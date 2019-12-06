# 使用 Knative Eventing 监控新服务，并发送通知邮件

这里我们将使用Broker和Trigger管理事件和订阅。我们将使用`ApiServerSource`作为事件源。

![](https://github.com/daisy-ycguo/knativelab/raw/master/images/Knative-triggermode.png)

发送到Broker的事件，将被转发给任何对该消息感兴趣的订阅者。

***注意*** 下面的操作要求在目录`devopslab/src/eventing`中进行，可以在CloudShell窗口执行下面命令进入该目录：
```
cd ~/devopslab/src/eventing/
```

## 前提

* Istio 和 Knative 在 IBM Kubernetes Cluster 上安装完毕。

## 步骤一：创建通知服务

通知服务是一个Knative上的服务，它接收`新服务创建`这个事件消息，并提取消息中的内容，将通知信息发送到邮箱。

您可以使用下面的命令创建通知服务，请注意将`<your email address>`的部分替换为您的有效邮箱：
```
kn service create --image docker.io/daisyycguo/servicenotifier mail-notifier --env EMAIL=$EMAIL
```
期待输出：
```
Service 'mail-notifier' successfully created in namespace 'default'.
Waiting for service 'mail-notifier' to become ready ... OK

Service URL:
http://mail-notifier-default.knative-guoyc-5290c8c8e5797924dc1ad5d1b85b37c0-0001.au-syd.containers.appdomain.cloud
```

## 步骤二：创建默认Broker

Broker可以通过两种方式创建：通过标记名称空间，可以创建默认的Broker。

运行命令：
```text
kubectl label namespace default knative-eventing-injection=enabled
```

期待输出：
```
namespace/default labeled
```

查看默认Broker已经被创建，运行命令:
```text
kubectl get broker
```

期待输出：
```
NAME      READY     REASON    HOSTNAME                                   AGE
default   True                default-broker.default.svc.cluster.local   14s
```

查看承载默认Borker的Pod已经启动，运行命令：
```
kubectl get pods
```

期待输出：
```
NAME                                              READY   STATUS    RESTARTS   AGE
default-broker-filter-798df8bc75-77m2r            1/1     Running   0          43s
default-broker-ingress-5fbb869648-q4xzb           1/1     Running   0          43s
```
他们一个是默认Broker的Ingress，负责接收消息，一个是默认Broker的过滤器，负责转发消息。

## 步骤二：创建ApiServerSource事件源

Kubernetes上的`ApiServerSource`能够捕获Kubernetes中的对象的创建、修改、删除等事件。这一步，我们将监控Knative Serving上运行的服务（Service）。

我们先来看一下`eventsource.yaml`的内容，这里描述了定时事件源的配置信息。

运行命令：
```text
cat eventsource.yaml
```

期待输出：
```
apiVersion: sources.eventing.knative.dev/v1alpha1
kind: ApiServerSource
metadata:
  name: service-update
  namespace: default
spec:
  serviceAccountName: service-sa
  mode: Resource
  resources:
  - apiVersion: serving.knative.dev/v1alpha1
    kind: Service
  sink:
    apiVersion: eventing.knative.dev/v1alpha1
    kind: Broker
    name: default
```

可以看到，它的`spec`主要包含四部分内容：
- serviceAccountName：监控程序运行时使用的ServiceAccount；
- resources：监控程序要监控的资源；
- sink：事件消息的发送目的地，这里的配置表示，事件消息将发送到默认的Broker中；
- mode：Resource表示事件信息里面传送的是完整的资源信息。 

首先创建ApiServerSource所需要的ServiceAccount，运行命令：
```
kubectl apply -f serviceaccount.yaml
```
期待输出：
```
serviceaccount/service-sa created
clusterrole.rbac.authorization.k8s.io/service-sa-watcher created
clusterrolebinding.rbac.authorization.k8s.io/service-sa-event-watcher-binding created
```

接着使用下面命令创建ApiServerSource `service-update`，运行命令：
```text
kubectl apply -f eventsource.yaml
```

期待输出：
```
apiserversource.sources.eventing.knative.dev/service-update created
```

查看`service-update`已经创建好，运行命令：
```text
kubectl get ApiServerSource
```

期待输出：
```
NAME             AGE
service-update   18s
```

检查承载`service-update`的Pod已经启动，运行命令：
```
kubectl get pods $(kubectl get pods --selector=eventing.knative.dev/sourceName=service-update --output=jsonpath="{.items..metadata.name}")
```

期待输出：
```
NAME                                                              READY     STATUS    RESTARTS   AGE
apiserversource-service-up-659843e6-14d0-11ea-82b7-b28431dtqgns   1/1       Running   0          99s
```

## 步骤三：创建Trigger，给Broker增加订阅者

Trigger表明了想要订阅某些事件的愿望。我们使用Trigger将服务`mail-notifier`订阅到默认的Broker，它将会把发送到Broker的消息打印到日志中。

我们先来看一下`trigger.yaml`的内容，这里描述了Trigger的配置信息，运行命令：
```text
cat trigger.yaml
```

期待输出：
```
apiVersion: eventing.knative.dev/v1alpha1
kind: Trigger
metadata:
  name: mytrigger
spec:
  broker: default
  filter:
    sourceAndType:
      type: dev.knative.apiserver.resource.update
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1alpha1
      kind: Service
      name: mail-notifier
```

可以看到，它的`spec`中的`subscriber`描述了一个订阅者，具体到这里，是`mail-notifier`服务。

使用下面命令创建Trigger `mytrigger`，运行命令：
```text
kubectl apply -f trigger.yaml
```

期待输出：
```
trigger.eventing.knative.dev/mytrigger created
```

查看`mytrigger`已经创建好，运行命令：
```text
kubectl get trigger
```

期待输出：
```
NAME        READY     REASON    BROKER    SUBSCRIBER_URI                                   AGE
mytrigger   True                default   http://mail-notifier.default.svc.cluster.local   3s
```

## 步骤四：获取通知邮件

在前面的实验中，您已经可以完成通过Tekton自动构建和部署Knative服务。接下来，您可以再次构建和部署一个Knative服务，查看`mail-notifier`服务被唤醒。在等待几分钟后，可以在您的邮箱看到通知邮件。

例如：您可以更新自己的devopslab repo中的 [src/app/hello.go](../src/app/hello.go)文件，将第17行修改为 `fmt.Fprintf(w, "%s\n", say("RED-IBM!!!"))` ，并且commit。  

下面命令将列出所有运行的Pod，观察`mail-notifier`应用所在的Pod已经开始运行，运行命令：
```
kubectl get pods
```

期待输出：
```
$ kubectl get pods
NAME                                                              READY     STATUS    RESTARTS   AGE
apiserversource-service-up-659843e6-14d0-11ea-82b7-b28431dtqgns   1/1       Running   0          6m53s
default-broker-filter-84dc88c975-8z6gl                            1/1       Running   0          8m15s
default-broker-ingress-7dd69cd749-zz9j7                           1/1       Running   0          8m15s
fib-knative-lmhzx-1-deployment-85899fb58-mxdrg                    2/2       Running   0          46s
mail-notifier-jhncd-1-deployment-69cf8dd499-cfpww                 2/2       Running   0          49s
```

查看`mail-notifier`的日志，运行命令：
```
kubectl logs -f $(kubectl get pods --selector=serving.knative.dev/configuration=mail-notifier --output=jsonpath="{.items..metadata.name}") user-container
```

能看到日志显示的CloudEvent标准消息如下面所示：
```
......
2019/12/02 08:42:09 Received an event:
2019/12/02 08:42:09 [2019-12-02T08:42:09.950309412Z] https://172.21.0.1:443 dev.knative.apiserver.resource.update
2019/12/02 08:42:09 Event Subject: /apis/serving.knative.dev/v1alpha1/namespaces/default/services/fib-knative
2019/12/02 08:42:09 send email to:  guoyingc@cn.ibm.com
2019/12/02 08:42:09 content:  Your service fib-knative is ready at http://fib-knative-default.knative-guoyc-5290c8c8e5797924dc1ad5d1b85b37c0-0001.au-syd.containers.appdomain.cloud.
2019/12/02 08:42:11 send email successfully.
```

可以看到，新服务的创建已经被监控到，并且发送通知邮件到您的邮箱。
请注意，您可能会因为邮件延迟过长而无法短时收到邮件。只要看到日志正确输出，则表示程序运转正常。

观察完毕，使用`ctrl + c`结束进程。

恭喜您，您已经完成了Knative Eventing的实验。下面进行[Serving流量管控](./01-serving.md)

