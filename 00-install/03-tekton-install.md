## 前提：确认Kubernetes集群以准备好
如果您已经按照文档00-install/01-k8s-connect.md准备好自己的Kubernetes集群，本步骤可以忽略。


如果想用IBM Cloud API Key 登陆IBM Cloud，需要记得自己的API Key，或者马上申请，用一下两种方法之一申请：
(1) 在浏览器上打开Console： https://cloud.ibm.com/iam/apikeys，点击 “Create an IBM Cloud API key”， 输入key名字和描述信息。 
 
 Key生成之后，下载API Key并保存为apikey.json：
 
 然后可以用Key登陆：
 ```
 $ ibmcloud login --apikey @apikey.json
 ```

(2) 通过命令行申请API key：
```
     $ ibmcloud iam api-key-create tekton-lab-apikey --file apikey.json
     Creating API key tekton-lab-apikey as someuser@email.address...
     OK
     API key tekton-lab-apikey was created
     Successfully save API key information to apikey.json
     $ ibmcloud login --apikey @apikey.json
```

## 第一步：在CloudShell里安装必要的工具
* 安装Tekton
```
$ kubectl apply --filename https://storage.googleapis.com/tekton-releases/latest/release.yaml
$ kubectl apply --filename https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
```
* 在准备好的集群上安装Knative
```
$ ibmcloud ks cluster addon enable knative --cluster testcluster -y
```

## 第二步：检查集群已经配置好
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

检查istio-system和knative-*** namespaces已经安装完毕。

```
$ kubectl get pods --namespace istio-system
NAME                                      READY   STATUS    RESTARTS   AGE
cluster-local-gateway-5d8ccd46db-jvfk6    1/1     Running   0          5d13h
istio-citadel-654897999b-blq9v            1/1     Running   0          5d13h
istio-egressgateway-77cfcd4f8d-44ztp      1/1     Running   0          5d13h
istio-egressgateway-77cfcd4f8d-64rsg      1/1     Running   0          5d13h
istio-galley-67987bf6cd-h5ngb             1/1     Running   0          5d13h
istio-ingressgateway-55b8654b85-69rhh     1/1     Running   0          5d13h
istio-ingressgateway-55b8654b85-tknnh     1/1     Running   0          5d13h
istio-pilot-796cfc6987-nnc8h              2/2     Running   0          5d13h
istio-policy-68f46ddd67-667ss             2/2     Running   3          5d13h
istio-sidecar-injector-75479d8b85-gg9qt   1/1     Running   0          5d13h
istio-telemetry-b8dbc5985-r2vzw           2/2     Running   4          5d13h
prometheus-7b87f6d744-n8gql               1/1     Running   0          5d13h

$ kubectl get pods --namespace knative-serving
NAME                                READY   STATUS    RESTARTS   AGE
activator-7654759547-csp8h          2/2     Running   4          5d13h
autoscaler-74878dccf9-8g6tz         2/2     Running   4          5d13h
autoscaler-hpa-6fc598cdb-g5rv7      1/1     Running   0          5d13h
controller-dc64bc644-9vbfc          1/1     Running   0          5d13h
networking-istio-65f5b87479-dbmhw   1/1     Running   0          5d13h
webhook-76c4d8d998-jmf85            1/1     Running   0          11h
```



