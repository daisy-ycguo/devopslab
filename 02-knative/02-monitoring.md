# 使用 Knative 监控服务


## 前提

* Istio 和 Knative 在 IBM Kubernetes Cluster 上安装完毕。
* Service `hello` 已通过 Knative pipeline 正确安装。

## 第一步：配置 Knative 监控环境

一， 验证 Knative 监控工具（Prometheus 和 Grafana）已正确安装：
```
$ kubectl get pods --namespace knative-monitoring
NAME                              READY   STATUS    RESTARTS   AGE
elasticsearch-logging-0           1/1     Running   0          13h
elasticsearch-logging-1           1/1     Running   0          13h
grafana-69bdcb4686-2x9wd          1/1     Running   0          13h
kibana-logging-8587dcbb89-pk428   1/1     Running   0          13h
prometheus-system-0               1/1     Running   0          13h
prometheus-system-1               1/1     Running   0          13h

$ kubectl get services --namespace knative-monitoring
NAME                          TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)               AGE
elasticsearch-logging         ClusterIP   172.21.121.125   <none>        9200/TCP              13h
fluentd-ds                    ClusterIP   172.21.99.72     <none>        24224/TCP,24224/UDP   13h
grafana                       NodePort    172.21.108.212   <none>        30802:30834/TCP       13h
kibana-logging                NodePort    172.21.191.39    <none>        5601:31253/TCP        13h
kube-controller-manager       ClusterIP   None             <none>        10252/TCP             13h
kube-state-metrics            ClusterIP   None             <none>        8443/TCP,9443/TCP     13h
node-exporter                 ClusterIP   None             <none>        9100/TCP              13h
prometheus-system-discovery   ClusterIP   None             <none>        9090/TCP              13h
prometheus-system-np          NodePort    172.21.98.144    <none>        8080:30132/TCP        13h
```

二， 获取 `hello` 服务地址，并验证服务：
```
$ kubectl get ksvc hello
NAME    URL                                                                      LATESTCREATED   LATESTREADY   READY   REASON
hello   http://hello-default.<CLUSTER-NAME>.us-south.containers.appdomain.cloud  hello-xxxx      hello-xxxx    True   

$ kubectl get service hello-xxxxx
NAME          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
hello-xxxx    ClusterIP   172.21.60.115   <none>        80/TCP    45m

$ curl http://hello-default.<CLUSTER-NAME>.us-south.containers.appdomain.cloud
[ 20191031 ] Hello world, this is BLUE-<your-name>!!!
```

三， 手动为服务添加工作负载
```
$ for i in {1..50}; do sleep 0.5; curl "http://hello-default.<CLUSTER-NAME>.us-south.containers.appdomain.cloud"; done
[ 20191031 ] Hello world, this is BLUE-<your-name>!!!
[ 20191031 ] Hello world, this is BLUE-<your-name>!!!
......
```

## 第二步：服务监控

一， Prometheus

- 获取 Prometheus 服务实例地址：

(1) 使用 CloudShell 时，无法利用本地 localhost 监听 Prometheus 服务实例，我们可以通过 `NodeIP:NodePort` 的方式直接访问 Prometheus ，输入以下命令获取访问地址：
```
$ echo $(kubectl get nodes -o jsonpath='{.items[0].status.addresses[1].address}'):$(kubectl -n knative-monitoring get services prometheus-system-np -ojsonpath='{.spec.ports[0].nodePort}')
xxx.xxx.xxx.xxx:xxxxx
```

(2) 当使用本地 CommandLine 工具时，我们还可以利用本地端口 **9090** 监听 Prometheus 服务实例，这种情况下 Prometheus 地址是 http://localhost:9090 ：

```
$ kubectl -n knative-monitoring port-forward \
  $(kubectl -n knative-monitoring get pod -l app=prometheus -o jsonpath='{.items[0].metadata.name}') \
  9090:9090 &
```

- 在浏览器中打开 Prometheus 工作窗口

   - 在 **Expression** 框中输入 `istio_requests_total` , 点击 **Execute** 按钮，您将在 **Graph** 或 **Console** 标签页中观测到最近一段时间内 Kuberneters 系统中所有请求的数量。

   - 在 **Expression** 框中输入 `istio_requests_total{destination_service_name=~"hello.*"}` （您需要使用实际的服务实例名称替换 `hello-xxxxx` ） , 点击 **Execute** 按钮，您将在 **Graph** 或 **Console** 标签页中观测到最近一段时间内所有路由到 `hello` 服务的请求数量。

![Prometheus](https://user-images.githubusercontent.com/42362436/70297723-1cda8f00-182a-11ea-93a4-168991f6355d.png)


二， Grafana

- 获取 Grafana 工作窗口地址：

(1) 使用 CloudShell 时，无法利用本地 localhost 监听 Grafana 服务实例，我们可以通过 `NodeIP:NodePort` 的方式直接访问 Grafana Dashboard，输入以下命令获取访问地址：

```
$ echo $(kubectl get nodes -o jsonpath='{.items[0].status.addresses[1].address}'):$(kubectl -n knative-monitoring get services grafana -ojsonpath='{.spec.ports[0].nodePort}')
xxx.xxx.xxx.xxx:xxxxx
```

(2) 当使用本地 CommandLine 工具时，我们还可以利用本地端口 **3000** 监听 Grafana 服务实例，这种情况下 Grafana 地址是 http://localhost:3000 ：

```
$ kubectl -n knative-monitoring port-forward \
  $(kubectl -n knative-monitoring get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') \
  3000:3000 &
```

- 在浏览器中打开 Grafana 工作窗口地址，点击 **Home** 菜单项，展开 **General** 列表，Grafana 已为您配置了多个的监控项供选择：
  - **Deployment**：在 **Namespace** 下拉框中选择 `default` , **Deployment** 下拉框中选择 `hello-xxxxx-deployment`，页面刷新完成后，您将观测到当前服务部署的统计指标图表。
  - **Pods**：在 **Namespace** 下拉框中选择 `default` , **Pod** 下拉框中选择 `hello-xxxxx-deployment-xxxxxxxxxx-xxxxx`，页面刷新完成后，您将观测到当前pod的统计指标图表。

- 您可以继续使用前面提到的命令为服务增加工作负载，持续监控服务数据。

```
$ for i in {1..50}; do sleep 0.5; curl "http://hello-default.<CLUSTER-NAME>.us-south.containers.appdomain.cloud?iterations=20000000"; done
[ 20191031 ] Hello world, this is BLUE-<your-name>!!!
[ 20191031 ] Hello world, this is BLUE-<your-name>!!!
......
```

这样，只须通过简单配置，Grafana 就可以为您的服务提供基础的监控能力。

另外，您还可以安装通过安装 Kiali 以及 Jaeger 工具帮助监测服务。参考 [Jaeger](https://www.jaegertracing.io/docs/1.15/) 以及 [Kiali](https://kiali.io/) 了解更多内容。

恭喜您，您已经完成了全部实验的内容。
