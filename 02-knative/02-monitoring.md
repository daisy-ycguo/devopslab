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

## 第二步：服务监控

一， Prometheus

- 获取 Prometheus 服务实例地址：

使用 CloudShell 时，无法利用本地 localhost 监听 Prometheus 服务实例，我们可以通过 `NodeIP:NodePort` 的方式直接访问 Prometheus ，输入以下命令获取访问地址：
```
$ echo $(kubectl get nodes -o jsonpath='{.items[0].status.addresses[1].address}'):$(kubectl -n knative-monitoring get services prometheus-system-np -ojsonpath='{.spec.ports[0].nodePort}')
xxx.xxx.xxx.xxx:xxxxx
```

- 在浏览器中打开 Prometheus 工作窗口

   - 在 **Expression** 框中输入 `istio_requests_total` , 点击 **Execute** 按钮，您将在 **Graph** 或 **Console** 标签页中观测到最近一段时间内 Kuberneters 系统中所有请求的数量。

   - 在 **Expression** 框中输入 `istio_requests_total{destination_service_name=~"hello.*"}`, 点击 **Execute** 按钮，您将在 **Graph** 或 **Console** 标签页中观测到最近一段时间内所有路由到 `hello` 服务的请求数量。

![Prometheus](https://user-images.githubusercontent.com/42362436/70297723-1cda8f00-182a-11ea-93a4-168991f6355d.png)

另外，您还可以通过[Grafana](https://grafana.com/)进行监控，或者通过安装 Kiali 以及 Jaeger 工具帮助监测服务。参考 [Jaeger](https://www.jaegertracing.io/docs/1.15/) 以及 [Kiali](https://kiali.io/) 了解更多内容。

恭喜您，您已经完成了全部实验的内容。
