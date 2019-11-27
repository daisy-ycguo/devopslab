# 使用 Istio 监控服务


## 前提

* Istio 和 Knative 在 IBM Kubernetes Cluster 上安装完毕。
* Service `picalc` 已通过 Knative pipeline 正确安装。

## 第一步：配置 Istio 监控工具

一， 安装 Istio 监控工具（Prometheus，Grafana，Kiali 和 Jaeger）：
```
$ curl -L https://git.io/getLatestIstio | ISTIO_VERSION=1.4.0 sh -
$ cd istio-1.4.0
$ kubectl apply -f install/kubernetes/istio-demo.yaml
```

二， 验证 Istio 监控工具已正确安装：
```
$ kubectl get pods -n istio-system
NAME                                      READY   STATUS      RESTARTS   AGE
cluster-local-gateway-6cc69b596c-b4rzx    1/1     Running     0          3h43m
grafana-584949b9c6-wm7l4                  1/1     Running     0          9m47s
istio-citadel-dc67bbdcc-fx5w8             1/1     Running     0          8m46s
istio-egressgateway-77498dc769-2hq6k      1/1     Running     0          8m46s
istio-egressgateway-77498dc769-jqfs4      1/1     Running     0          8m34s
istio-galley-6f98b5c7cf-7vdhm             1/1     Running     0          8m46s
istio-grafana-post-install-1.4.0-2j972    0/1     Completed   0          10m
istio-ingressgateway-774f7ff8b6-d7vjx     1/1     Running     0          8m45s
istio-ingressgateway-774f7ff8b6-j82j6     1/1     Running     0          8m34s
istio-pilot-848b8fdb4-4gd6z               2/2     Running     0          8m45s
istio-policy-cbfcb8f9b-nxbxz              2/2     Running     0          8m45s
istio-security-post-install-1.4.0-fpn2p   0/1     Completed   0          10m
istio-sidecar-injector-7bb7fdffbb-fx7n4   1/1     Running     0          8m45s
istio-telemetry-568789f56b-g6l6f          2/2     Running     0          8m45s
istio-tracing-795c9c64c4-ds7dm            1/1     Running     0          9m41s
kiali-7d4cf866cc-4rf9t                    1/1     Running     0          9m46s
prometheus-5d99c46489-h5ggl               1/1     Running     0          8m45s

$ kubectl get services -n istio-system
NAME                     TYPE           CLUSTER-IP       EXTERNAL-IP     PORT(S)                                                                                                                                      AGE
cluster-local-gateway    ClusterIP      172.21.228.5     <none>          80/TCP,443/TCP,31400/TCP,15011/TCP,8060/TCP,15029/TCP,15030/TCP,15031/TCP,15032/TCP                                                          3h43m
grafana                  ClusterIP      172.21.75.179    <none>          3000/TCP                                                                                                                                     10m
istio-citadel            ClusterIP      172.21.255.43    <none>          8060/TCP,15014/TCP                                                                                                                           3h43m
istio-egressgateway      ClusterIP      172.21.133.206   <none>          80/TCP,443/TCP,15443/TCP                                                                                                                     3h43m
istio-galley             ClusterIP      172.21.133.211   <none>          443/TCP,15014/TCP,9901/TCP                                                                                                                   3h43m
istio-ingressgateway     LoadBalancer   172.21.238.12    150.238.5.163   15020:31004/TCP,80:31380/TCP,443:31390/TCP,31400:31400/TCP,15029:30244/TCP,15030:30569/TCP,15031:32102/TCP,15032:32537/TCP,15443:31824/TCP   3h43m
istio-pilot              ClusterIP      172.21.137.197   <none>          15010/TCP,15011/TCP,8080/TCP,15014/TCP                                                                                                       3h43m
istio-policy             ClusterIP      172.21.44.93     <none>          9091/TCP,15004/TCP,15014/TCP                                                                                                                 3h43m
istio-sidecar-injector   ClusterIP      172.21.174.177   <none>          443/TCP,15014/TCP                                                                                                                            3h43m
istio-telemetry          ClusterIP      172.21.231.193   <none>          9091/TCP,15004/TCP,15014/TCP,42422/TCP                                                                                                       3h43m
jaeger-agent             ClusterIP      None             <none>          5775/UDP,6831/UDP,6832/UDP                                                                                                                   10m
jaeger-collector         ClusterIP      172.21.248.2     <none>          14267/TCP,14268/TCP,14250/TCP                                                                                                                10m
jaeger-query             ClusterIP      172.21.49.231    <none>          16686/TCP                                                                                                                                    10m
kiali                    ClusterIP      172.21.58.253    <none>          20001/TCP                                                                                                                                    10m
prometheus               ClusterIP      172.21.187.128   <none>          9090/TCP                                                                                                                                     3h43m
tracing                  ClusterIP      172.21.178.150   <none>          9411/TCP                                                                                                                                     10m
zipkin                   ClusterIP      172.21.229.24    <none>          9411/TCP
```

三， 获取服务地址，并验证服务地址：
```
$ kubectl get ksvc picalc
NAME     URL                                                                                                                       LATESTCREATED   LATESTREADY    READY   REASON
picalc   http://picalc-default.<CLUSTER-NAME>.us-south.containers.appdomain.cloud   picalc-76xmn    picalc-76xmn   True

$ curl "http://picalc-default.<CLUSTER-NAME>.us-south.containers.appdomain.cloud?iterations=20000000"
3.1415926036
```

四， 手动为服务添加工作负载
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
kubectl -n istio-system port-forward \
  $(kubectl -n istio-system get pod -l app=prometheus -o jsonpath='{.items[0].metadata.name}') \
  9090:9090 &
```

- 在浏览器中打开 Prometheus 工作窗口：http://localhost:9090/graph

- 在 **Expression** 框中输入 `istio_requests_total` , 点击 **Execute** 按钮，您将在 **Graph** 或 **Console** 标签页中观测到最近一段时间内 Kuberneters 系统中所有请求的数量。

- 在 **Expression** 框中输入 `istio_requests_total{destination_service="picalc-76xmn.default.svc.cluster.local"}` , 点击 **Execute** 按钮，您将在 **Graph** 或 **Console** 标签页中观测到最近一段时间内所有路由到 `picalc-76xmn` 服务的请求数量。

二， Grafana

- 使用本地端口 **3000** 监听 Grafana 服务实例：
```
kubectl -n istio-system port-forward \
  $(kubectl -n istio-system get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') \
  3000:3000 &
```

- 在浏览器中打开 Grafana 工作窗口：http://localhost:3000 ，点击 **Home** 菜单项，展开 **Istio** 列表，点击 **Istio Service Dashboard**。在 Service 下拉框中选择 `picalc-XXXXX.default.svc.cluster.local` 服务，刷新完成后，您将在页面中观测到当前服务 Request 相关统计指标的图表。

- 您可以继续使用前面提到的命令为服务增加工作负载，持续监控服务数据。

```
$ for i in {1..50}; do sleep 0.5; curl "http://picalc-default.<CLUSTER-NAME>.us-south.containers.appdomain.cloud?iterations=20000000"; done
3.1415926036
3.1415926036
......
```

这样，只须通过简单配置，Grafana 就可以为您的服务提供基础的监控能力。

另外，第一步中的步骤二已为您的 Cluster 安装了 Kiali 以及 Jaeger 工具，它们同样可以帮助监测服务。参考 [Jaeger](https://www.jaegertracing.io/docs/1.15/) 以及 [Kiali](https://kiali.io/) 了解更多内容。
