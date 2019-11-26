# Tekton Hands-on Lab - from source code to production 

## 实验目标
图

## 实验准备   
1. 登录Online command Line tool   
用浏览器打开https://workshop.shell.cloud.ibm.com/    
passcode: ikslab

2. 登录Kubecluster    
>~~apikey怎么拿到~~   
`ibmcloud login --apikey xxxxxx`   

3. 确保您已完成[准备 Kubecluster 和安装 knative](https://github.com/QiuJieLi/devopslab/tree/master/00-install)的步骤
验证knative和isto的pods都是Running状态   
*示例:*
```
$ kubectl get pods --namespace istio-system
NAME                                      READY   STATUS    RESTARTS   AGE
cluster-local-gateway-5d8ccd46db-kt4vk    1/1     Running   0          6h14m
istio-citadel-654897999b-pbxjs            1/1     Running   0          6h15m
istio-egressgateway-77cfcd4f8d-8cbw2      1/1     Running   0          6h15m
istio-egressgateway-77cfcd4f8d-xgdfr      1/1     Running   0          6h14m
istio-galley-67987bf6cd-xwhlj             1/1     Running   0          6h15m
istio-ingressgateway-55b8654b85-8k9lh     1/1     Running   0          6h14m
istio-ingressgateway-55b8654b85-gph7k     1/1     Running   0          6h15m
istio-pilot-796cfc6987-b7qmg              2/2     Running   0          6h15m
istio-policy-68f46ddd67-sn7x2             2/2     Running   4          6h15m
istio-sidecar-injector-75479d8b85-vnwk8   1/1     Running   0          6h15m
istio-telemetry-b8dbc5985-wmj48           2/2     Running   4          6h15m
prometheus-7b87f6d744-rpgnq               1/1     Running   0          6h15m

$ kubectl get pods --namespace knative-serving
NAME                                READY   STATUS    RESTARTS   AGE
activator-7654759547-vc45g          2/2     Running   2          6h13m
autoscaler-74878dccf9-7b8gc         2/2     Running   2          6h13m
autoscaler-hpa-6fc598cdb-njhnf      1/1     Running   0          6h13m
controller-dc64bc644-s5f24          1/1     Running   0          6h13m
networking-istio-65f5b87479-nxftw   1/1     Running   0          6h13m
webhook-76c4d8d998-5sg4p            1/1     Running   0          5h12m
```

4. 安装 Tekton   
`kubectl apply --filename https://storage.googleapis.com/tekton-releases/latest/release.yaml`
>需要么？跟5 重复？   
>`unable to recognize "https://storage.googleapis.com/tekton-releases/latest/release.yaml": Get http://localhost:8080/api?timeout=32s: dial tcp [::1]:8080: connect: connection refused`   

5. 安装 Tekton pipeline 及其依赖   
`kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml`    
使用以下命令查看Tekton Pipelines components 直到 STATUS 都显示 `Running`:   
*示例:*
```
$ kubectl get pods --namespace tekton-pipelines
NAME                                           READY   STATUS    RESTARTS   AGE
tekton-pipelines-controller-7769bc5b76-tzsq9   1/1     Running   0          5h49m
tekton-pipelines-webhook-7849d4f75f-vd49j      1/1     Running   0          4h47m
```

6. 安装 Kn Client   
安装步骤参考https://github.com/knative/client/blob/master/docs/README.md
例如，在linux上   
```
$ wget https://storage.googleapis.com/knative-nightly/client/latest/kn-linux-amd64
$ chmod +x kn-linux-amd64
$ export PATH=$PATH:$PWD
```   

验证kn安装成功   
```
$ kn
Manage your Knative building blocks:
...
```

7. Fork tekton-tutorial项目到您自己的git repo中 - 因为后续步骤需要修改代码。Clone您fork的项目到本地目录.   
在浏览器中打开`https://github.com/IBM/tekton-tutorial`，点击Fork   
Clone您fork的项目到本地目录   
```
$ git clone https://github.com/<您的git名>/tekton-tutorial
$ cd tekton-tutorial
$ ls
ACKNOWLEDGEMENTS.md	LICENSE			README.md		knative			tekton
CONTRIBUTING.md		MAINTAINERS.md		doc			src
```   

## 实验步骤
1. 创建一个Task来build一个image并push到container registry。这里使用了[kaniko](https://github.com/GoogleContainerTools/kaniko)。   
```
$ kubectl apply -f tekton/tasks/source-to-image.yaml
task.tekton.dev/source-to-image created
```

2. 创建另一个Task来将image部署到Kubernetes cluster   
```
$ kubectl apply -f tekton/tasks/deploy-using-kubectl.yaml
task.tekton.dev/deploy-using-kubectl created
```

3. 创建一个Pipeline来组合以上两个Task
```
$ kubectl apply -f tekton/pipeline/build-and-deploy-pipeline.yaml
pipeline.tekton.dev/build-and-deploy-pipeline created
```

4. 修改Pipelinerun，指向正确的<REGISTRY>/<NAMESPACE>   
- 修改tekton/run/picalc-pipeline-run.yaml   
将文件中的<REGISTRY>和<NAMESPACE>用您个人account下的private container registry信息替代
```
apiVersion: tekton.dev/v1alpha1
kind: PipelineRun
metadata:
  generateName: picalc-pr-
spec:
  pipelineRef:
    name: build-and-deploy-pipeline
  resources:
    - name: git-source
      resourceRef:
        name: picalc-git
  params:
    - name: pathToYamlFile
      value: "knative/picalc.yaml"
    - name: imageUrl
      value: <REGISTRY>/<NAMESPACE>/picalc
    - name: imageTag
      value: "1.0"
  trigger:
    type: manual
  serviceAccount: pipeline-account
```   
执行以下命令获得<REGISTRY>，在以下例子中<REGISTRY>为us.icr.io
```
$ ibmcloud cr region
You are targeting region 'us-south', the registry is 'us.icr.io'.
```

在您个人账户下的IBM Container Registry中创建一个<NAMESPACE>。<NAMESPACE>为tektonlab

```
$ ibmcloud cr login
$ ibmcloud cr namespace-add tektonlab
Adding namespace 'tektonlab'...

Successfully added namespace 'tektonlab'

OK
```

5. 创建secret  

替换<APIKEY>为实验准备->2 中使用的apikey，执行以下命令创建一个secret
```
kubectl create secret generic ibm-cr-push-secret --type="kubernetes.io/basic-auth" --from-literal=username=iamapikey --from-literal=password=<APIKEY>
```
*输出示例:*
`kubectl annotate secret ibm-cr-push-secret tekton.dev/docker-0=secret/ibm-cr-push-secret created`

替换<REGISTRY>为实验步骤->4 中获得的<REGISTRY>
```
kubectl annotate secret ibm-cr-push-secret tekton.dev/docker-0=<REGISTRY>

```
*输出示例:*
`secret/ibm-cr-push-secret annotated`

6. 创建service account。其中使用了上一步创建的secret   
```
kubectl apply -f tekton/pipeline-account.yaml
```

7. 创建pipelineresource,指向应用代码所在的repo
```
$ kubectl apply -f tekton/resources/picalc-git.yaml
pipelineresource.tekton.dev/picalc-git created
```

8. 执行Pipeline   
```
$ kubectl create -f tekton/run/picalc-pipeline-run.yaml
pipelinerun.tekton.dev/picalc-pr-ps5tv created
```

9. 检查
