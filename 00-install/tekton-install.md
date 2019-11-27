# 在提前准备好的Kubernetes集群上安装必要的工具

Knative Lab使用了IBM公有云上的Kubernetes集群，以及一个云上的命令行窗口CloudShell。在开始实验前，Kubernetes集群已经准备好了。您只需要拥有IBM Cloud的注册账号，就可以进行下面的操作。

## 前提

* 拥有一个IBM Cloud账号，也被称为IBM ID。如果没有注册，请到[http://cloud.ibm.com](http://cloud.ibm.com)上注册。
* 准备一个可以联网的浏览器，推荐Chrome，Firefox，和Safari。
* 在IBM公有云上准备一个可以使用的Kubernetes集群。

## 第一步：访问在线CommondLine 工具：
https://workshop.shell.cloud.ibm.com/

## 第二步：访问准备好的Kubenetes 集群，假设集群名字为testcluster：
```
$ ibmcloud login -u *** -p ***  (or –apikey)
$ ibmcloud ks cluster-config testcluster
```
从第二个命令的输出，拷贝`export KUBECONFIG=/…./xxx-testcluster.yml’ 并在命令行执行。
然后检查集群信息
```
$ ibmcloud ks cluster-get testcluster
```

对于OpenShift集群
* 登陆IBM Cloud，之后获得访问OpenShift集群的token： 
https://c100-e.us-south.containers.cloud.ibm.com:31465/oauth/token/display

* 拷贝类似下面的屏幕输出并在命令行执行： 
```
$ oc login --token=*** --server=https://c100-e.us-south.containers.cloud.ibm.com:31465
```

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

## 第三步：在在线CommondLine 里安装必要的工具：
* 检查必要的工具已经安装上了（ibmcloud,kubectl, ibmcloud ks, ibmcloud cr, git, kn)
* 安装Tekton
```
$ kubectl apply --filename https://storage.googleapis.com/tekton-releases/latest/release.yaml
$ kubectl apply --filename https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
```
* 在准备好的集群上安装Knative
```
$ ibmcloud ks cluster addon enable knative --cluster testcluster -y
```

## 第四步：检查集群已经配置好：
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



