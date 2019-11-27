# 搭建Tekton pipeline及其依赖

## 1.检查Tekton pipeline，保证所有的pod都是Running状态
```
$ kubectl get pods --namespace tekton-pipelines
```
## 2.Fork tekton-tutorial 项目到你自己的repo并且clone到local workstation
```
$ git clone https://github.com/<Your repo>/tekton-tutorial
```
## 3.创建一个任务来构建image并且push到自己的container registry
```
$ kubectl apply -f tekton/tasks/source-to-image.yaml 
```

## 4.创建一个任务来 把image部署到你的Kubernetes cluster
```
$ kubectl apply -f tekton/tasks/deploy-using-kubectl.yaml
```

## 5.创建一个pipeline
```
$ kubectl apply -f tekton/pipeline/build-and-deploy-pipeline.yaml
```

## 6.创建一个pipeline run和pipeline resource
编辑文件 tekton/run/picalc-pipeline-run.yaml，把<REGISTRY>和<NAMESPACE>替换成自己的container registry。
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
注意：
- <REGISTRY>来自于```ibmcloud cr region```的输出。
- <NAMESPACE>来自于```ibmcloud cr namespace-list```的输出。
 
修改完成后执行：
```
$ kubectl apply -f tekton/resources/picalc-git.yaml
```

## 7.定义一个service account
你在安装实验中获取了集群的token或者API key后，用下列步骤创建secret。

```
$ kubectl create secret generic ibm-cr-push-secret --type="kubernetes.io/basic-auth" --from-literal=username=<USER> --from-literal=password=<TOKEN/APIKEY>
$ kubectl annotate secret ibm-cr-push-secret tekton.dev/docker-0=<REGISTRY>
```
其中，
- <USER> 或者是你在登陆时使用的token，或者是如果你使用API key的方法登陆ibmcloud，<USER> 就是字符串“iamapikey” （纯字符串，不是真正的API key的值）
- <REGISTRY> 是你的container registry的URL，如：us.icr.io 或者 registry.ng.bluemix.net。通过如下命令获得<REGISTRY> 的值：
```
$ ibmcloud cr region
You are targeting region 'us-south', the registry is 'us.icr.io'.
```

创建secret的例子：
```
$ kubectl create secret generic ibm-cr-push-secret --type="kubernetes.io/basic-auth" --from-literal=username=iamapikey --from-literal=password=*** 
$ kubectl annotate secret ibm-cr-push-secret tekton.dev/docker-0=us.icr.io 
```
然后apply这个account：
```
$ kubectl apply -f tekton/pipeline-account.yaml
```

## 8.运行 pipeline:
```
$ kubectl create -f tekton/run/picalc-pipeline-run.yaml
pipelinerun.tekton.dev/picalc-pr-db6p6 created
```
检查taskruns的状态，知道成功：
```
$ kubectl get taskruns
NAME                                          SUCCEEDED   REASON      STARTTIME   COMPLETIONTIME
picalc-pr-kpxbr-deploy-to-cluster-j5bpj       True        Succeeded   53s         41s
picalc-pr-kpxbr-source-to-image-hbbw4         True        Succeeded   108s        53s
```
注意：如果以上步骤遇到问题，尝试下面的方法进行诊断：
1. 如果一个taskrun失败了，检查该taskrun的describe之后的‘Message’：
e.g.
```
tkn taskrun describe picalc-pr-kpxbr-source-to-image-hbbw4
...
Message
"step-build-and-push" exited with code 1 (image: "gcr.io/kaniko-project/executor@sha256:9c40a04cf1bc9d886f7f000e0b7fa5300c31c89e2ad001e97eeeecdce9f07a29"); for logs run: kubectl -n default logs picalc-pr-7gr8f-source-to-image-sn7r8-pod-02e5da -c step-build-and-push
```
按照提示检查pod的log获取失败的详细信息。

例如：
```
$ kubectl -n default logs picalc-pr-fwdz7-source-to-image-6tvck-pod-9a97c6 -c step-build-and-push
error checking push permissions -- make sure you entered the correct tag name, and that you are authenticated correctly, and try again: checking push permission for "us.icr.io/sophy/picalc:1.0": DENIED: You have exceeded your storage quota. Delete one or more images, or review your storage quota and pricing plan. For more information, see https://ibm.biz/BdjFwL; [map[Action:pull Class: Name:sophy/picalc Type:repository] map[Action:push Class: Name:sophy/picalc Type:repository]]
```

## 9.检查pipeline的状态
```
$ kubectl describe pipelinerun picalc-pr-dqktl
```
最后的Events应该是‘Succeeded’状态，例如：
```
Normal   Succeeded          5m43s                  pipeline-controller  All Tasks have completed executing
```

## 检查Knative service的状态是否是Ready
```
$ kubectl get ksvc picalc
NAME      DOMAIN                                                              LATESTCREATED   LATESTREADY    READY     REASON
picalc    picalc-default.testcluster.us-south.containers.appdomain.cloud   picalc-00001    picalc-00001   True

$ curl picalc-default.testcluster.us-south.containers.appdomain.cloud?iterations=20000000
   3.1415926036
```

注意：如果以上步骤遇到问题，尝试下面的方法进行诊断：
1. 如果knative service的状态不是Ready，运行命令`kubectl describe ksvc picalc`检查出错的详细信息。
例如：
```
$ kubectl describe ksvc picalc
...
    Message:                     Revision "picalc-zf4fg" failed with message: Unable to fetch image "us.icr.io/liqiujie/picalc:1.0": UNAUTHORIZED: authentication required; [map[Action:pull Class: Name:liqiujie/picalc Type:repository]].
```
