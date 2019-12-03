# 使用 Istio 监控服务


## 前提

* Istio 和 Knative 在 IBM Kubernetes Cluster 上安装完毕。
* Service `picalc` 已通过 Knative pipeline 正确安装。

## 第一步：配置 Knative 监控环境

一， 验证 Knative 监控工具（Prometheus 和 Grafana）已正确安装：
```
$ kubectl get pods --namespace knative-monitoring
NAME                                  READY   STATUS    RESTARTS   AGE
elasticsearch-logging-0               1/1     Running   0          9h
elasticsearch-logging-1               1/1     Running   0          9h
grafana-84fdfd44db-t5wcj              1/1     Running   0          9h
kibana-logging-75bc875c85-25sjp       1/1     Running   0          9h
kube-state-metrics-7ccb9449df-2htsh   4/4     Running   0          9h
node-exporter-kzljf                   2/2     Running   0          9h
node-exporter-pvzlg                   2/2     Running   0          9h
node-exporter-zgwwf                   2/2     Running   0          9h
prometheus-system-0                   1/1     Running   0          9h
prometheus-system-1                   1/1     Running   0          9h

$ kubectl get services --namespace knative-monitoring
NAME                          TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)               AGE
elasticsearch-logging         ClusterIP   172.21.167.162   <none>        9200/TCP              9h
fluentd-ds                    ClusterIP   172.21.208.103   <none>        24224/TCP,24224/UDP   9h
grafana                       NodePort    172.21.77.228    <none>        30802:30979/TCP       9h
kibana-logging                NodePort    172.21.90.152    <none>        5601:31462/TCP        9h
kube-controller-manager       ClusterIP   None             <none>        10252/TCP             9h
kube-state-metrics            ClusterIP   None             <none>        8443/TCP,9443/TCP     9h
node-exporter                 ClusterIP   None             <none>        9100/TCP              9h
prometheus-system-discovery   ClusterIP   None             <none>        9090/TCP              9h
prometheus-system-np          NodePort    172.21.202.105   <none>        8080:32344/TCP        9h
```

二， 获取 `picalc` 服务地址，并验证服务：
```
$ kubectl get ksvc picalc
NAME     URL                                                                        LATESTCREATED   LATESTREADY    READY   REASON
picalc   http://picalc-default.<CLUSTER-NAME>.us-south.containers.appdomain.cloud   picalc-xxxxx    picalc-xxxxx   True

$ kubectl get service picalc-xxxxx
NAME           TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)   AGE
picalc-xxxxx   ClusterIP   172.21.255.191   <none>        80/TCP    8h

$ kubectl get deployment
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
picalc-xxxxx-deployment   1/1     1            1           8h

$ kubectl get pods
NAME                                       READY   STATUS    RESTARTS   AGE
picalc-xxxxx-deployment-xxxxxxxxxx-xxxxx   2/2     Running   0          14m

$ curl "http://picalc-default.<CLUSTER-NAME>.us-south.containers.appdomain.cloud?iterations=20000000"
3.1415926036
```

三， 手动为服务添加工作负载
```
$ for i in {1..50}; do sleep 0.5; curl "http://picalc-default.<CLUSTER-NAME>.us-south.containers.appdomain.cloud?iterations=20000000"; done
3.1415926036
3.1415926036
......
```

## 第二步：服务监控

一， Prometheus

- 使用本地端口 **9090** 监听 Prometheus 服务实例：
```
kubectl -n knative-monitoring port-forward \
  $(kubectl -n knative-monitoring get pod -l app=prometheus -o jsonpath='{.items[0].metadata.name}') \
  9090:9090 &
```

- 在浏览器中打开 Prometheus 工作窗口：http://localhost:9090/graph

- 在 **Expression** 框中输入 `istio_requests_total` , 点击 **Execute** 按钮，您将在 **Graph** 或 **Console** 标签页中观测到最近一段时间内 Kuberneters 系统中所有请求的数量。

- 在 **Expression** 框中输入 `istio_requests_total{destination_service_name='picalc-xxxxx'}` （您需要使用实际的服务实例名称替换 `picalc-xxxxx` ） , 点击 **Execute** 按钮，您将在 **Graph** 或 **Console** 标签页中观测到最近一段时间内所有路由到 `picalc` 服务的请求数量。

二， Grafana

- 使用本地端口 **3000** 监听 Grafana 服务实例：
```
kubectl -n knative-monitoring port-forward \
  $(kubectl -n knative-monitoring get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') \
  3000:3000 &
```

- 在浏览器中打开 Grafana 工作窗口：http://localhost:3000 ，点击 **Home** 菜单项，展开 **General** 列表，Grafana 已为您配置了多个的监控项供选择：
  - **Deployment**：在 **Namespace** 下拉框中选择 `default` , **Deployment** 下拉框中选择 `picalc-xxxxx-deployment`，页面刷新完成后，您将观测到当前服务部署的统计指标图表。
  - **Pods**：在 **Namespace** 下拉框中选择 `default` , **Pod** 下拉框中选择 `picalc-xxxxx-deployment-xxxxxxxxxx-xxxxx`，页面刷新完成后，您将观测到当前pod的统计指标图表。

- 您可以继续使用前面提到的命令为服务增加工作负载，持续监控服务数据。

```
$ for i in {1..50}; do sleep 0.5; curl "http://picalc-default.<CLUSTER-NAME>.us-south.containers.appdomain.cloud?iterations=20000000"; done
3.1415926036
3.1415926036
......
```

这样，只须通过简单配置，Grafana 就可以为您的服务提供基础的监控能力。

另外，您还可以安装通过安装 Kiali 以及 Jaeger 工具帮助监测服务。参考 [Jaeger](https://www.jaegertracing.io/docs/1.15/) 以及 [Kiali](https://kiali.io/) 了解更多内容。
