# 安装Tekton

Knative基于Kubernetes和Istio。IBM公有云上提供的Kubernetes集群可以一键安装Istio和Knative，省去安装的烦恼。

## 前提

* 分配到一个IBM Kubernetes Cluster；
* 启动CloudShell云端命令行窗口，本次试验的所有命令行输入都在CloudShell窗口中完成；
* 通过kubectl连接到了云端的Kubernetes集群。

## 第一步：使用IBM Cloud命令行工具安装

安装Tekton
```
$ kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.8.0/release.yaml
$ kubectl apply --filename https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
```


## 第二步：检查Tekton的Pipleline和Trigger已经安装好

查看所有安装的名称空间：
```
$ kubectl get namespace
NAME                 STATUS   AGE
default              Active   60d
ibm-cert-store       Active   60d
ibm-system           Active   60d
istio-system         Active   5d13h
knative-eventing     Active   5d13h
knative-monitoring   Active   5d13h
knative-serving      Active   5d13h
knative-sources      Active   5d13h
kube-node-lease      Active   60d
kube-public          Active   60d
kube-system          Active   60d
tekton-pipelines     Active   5d13h
```

查看tekton-pipelines名称空间下面的pod，确认tekton-pipelines*和tekton-triggers都处于运行状态。
```
$ kubectl get pods -n tekton-pipelines
NAME                                           READY     STATUS    RESTARTS   AGE
tekton-pipelines-controller-756f4f448f-x76fh   1/1       Running   0          5m5s
tekton-pipelines-controller-7769bc5b76-b7jx2   1/1       Running   0          20s
tekton-pipelines-webhook-7849d4f75f-6z8ds      1/1       Running   0          20s
tekton-pipelines-webhook-79fcc6c768-sq52q      1/1       Running   0          5m5s
tekton-triggers-controller-6d8fd9596b-wv5nw    1/1       Running   0          4m3s
tekton-triggers-webhook-5f8d569d5f-8v58q       1/1       Running   0          4m2s
```

恭喜您，您的环境已经准备好了。下面可以开始实验了[Tekton实验](./01-exercise-1.md)

