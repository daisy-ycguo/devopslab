# 练习二：Tekton Trigger

## 实验目标

- 了解Trigger的基本概念
- 创建TriggerTemplate, TriggerBinding和EventListener
- 创建一个pipeline
- 为git repo创建一个webhook,当有push操作发生的时候，发送请求到EventListener来创建pipeline resource和pipeline run,执行指定的pipeline。

## 基本概念

- TriggerTemplate - 资源模板 例如：PipelineResources和PipelineRun
- TriggerBinding - 检验events并且抽取payload中的fields
- EventListener - 将TriggerBindings和TriggerTemplates连接起来，提供一个可访问的endpoint (事件接收器). 它使用TriggerBinding从events中抽取出来的内容作为参数，来创建TriggerTemplate中指定的资源。

## 前提
成功完成[tekton exercise-1](./01-exercise-1.md)

## 实验步骤
下面的实验中，我们将使用Trigger来创建一个PipelineRun和一个PipelineResource。这个PipelineRun运行了我们[tekton exercise-1](./01-exercise-1.md)中创建的pipeline - "build-and-deploy-pipeline"。

### 1. 按如下步骤配置Trigger

1.1 创建service account
在CloudShell中请执行命令：
```
kubectl apply -f devopslab/src/tekton/trigger/role-resources
```

期待输出：
```
rolebinding.rbac.authorization.k8s.io/tekton-triggers-example-binding created
role.rbac.authorization.k8s.io/tekton-triggers-example-minimal created
secret/githubsecret created
serviceaccount/tekton-triggers-example-sa created
```

1.2 更新devopslab/src/tekton/trigger/triggertemplates/triggertemplate.yaml，将PipelineRun中的参数imageUrl的值`us.icr.io/<NAMESPACE>/hello`替换。请参考[tekton exercise-1](./01-exercise-1.md)步骤5.5   
  
```
...
  - apiVersion: tekton.dev/v1alpha1
    kind: PipelineRun
    metadata:
      generateName: hello-pr-
    spec:
      pipelineRef:
        name: build-and-deploy-pipeline
      resources:
        - name: git-source
          resourceRef:
            name: hello-git-${uid}
      params:
        - name: pathToYamlFile
          value: "src/tekton/basic/knative/hello.yaml"
        - name: imageUrl
          value: us.icr.io/<NAMESPACE>/hello
        - name: imageTag
          value: "2.0"
      serviceAccount: pipeline-account
```
1.3 创建TriggerTemplate   

在CloudShell中请执行命令：
```
kubectl apply -f devopslab/src/tekton/trigger/triggertemplates/triggertemplate.yaml
```

期待输出：
```
triggertemplate.tekton.dev/my-pipeline-template created
```
1.4 创建TriggerBinding   

在CloudShell中请执行命令：
```
kubectl apply -f devopslab/src/tekton/trigger/triggerbindings/triggerbinding.yaml
```

期待输出：
```
triggerbinding.tekton.dev/my-pipeline-binding created
```
1.5 创建EventListener   

在CloudShell中请执行命令：
```
kubectl apply -f devopslab/src/tekton/trigger/eventlisteners/eventlistener.yaml
```

期待输出：
```
eventlistener.tekton.dev/my-listener created
```

#### 2 检查我们所需要的pods和services已经建好并且状态健康
在CloudShell中请执行命令：
```
kubectl get svc
```

期待输出：
```
NAME                 TYPE           CLUSTER-IP      EXTERNAL-IP                                           PORT(S)             AGE
el-my-listener       ClusterIP      172.21.142.64   <none>                                                8080/TCP            4s
```

在CloudShell中请执行命令：
```
kubectl get pods
```

期待输出：
```
NAME                                                 READY   STATUS      RESTARTS   AGE
el-my-listener-99b595cc6-4vqq6                          1/1     Running     0          21s
```

#### 3 配置Ingress
使得listner endpoint可以被从cluster外部访问。后面我们会通过git repository的webhook来访问这个listener endpoint。
1. 获取你的集群的Ingress Subdomain

在CloudShell中请执行命令：
```
ibmcloud ks cluster-get $MYCLUSTER | grep 'Ingress Subdomain'
```

期待输出：
```
Ingress Subdomain: <CLUSTER-NAME>.us-south.containers.appdomain.cloud
```

2. 更新devopslab/src/tekton/trigger/ingress.yaml文件
- 将host的值替换为el-listener.<INGRESS-SUBDOMAIN>， 例如 el-my-listener.testcluster-973348.us-south.containers.appdomain.cloud    
```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: el-my-listener
  namespace: default
spec:
  rules:
  - host: el-my-listener.<INGRESS-SUBDOMAIN>
    http:
      paths:
      - backend:
          serviceName: el-my-listener
          servicePort: 8080
        path: /
```
3. Apply ingress文件:
在CloudShell中请执行命令：
```
kubectl apply -f devopslab/src/tekton/trigger/ingress.yaml
```

期待输出：
```
ingress.extensions/el-my-listener created
```
4. 查看ingress
在CloudShell中请执行命令：
```
kubectl get ingress
```

期待输出：
```
NAME             HOSTS                                                              ADDRESS         PORTS   AGE
el-my-listener   el-my-listener.testcluster-973348.us-south.us-south.containers.appdomain.cloud   169.47.66.178   80      14s
```

#### 4 配置webhook
当指定的event发生时，Webhook会发送一个POST请求到其配置的URL。这个URL就是我们上面建好的listener的endpoint。
进入你的git账户下的devopslab代码库，配置这个代码库的webhook。 

1. 在浏览器中您的devopslab repo `https://github.com/<your-git-account>/devopslab`
2. 点击最右侧的Settings tab，从左侧导航栏选择Webhooks   
3. 点击'Add webhook'按钮   
4. 添加 webhook
 其中：
 - Payload URL: http://加上步骤2.4中配置的ingress的HOSTS 例如：http://el-my-listener.testcluster-973348.us-south.us-south.containers.appdomain.cloud
 - Content type: application/json
 - Secret: 不用填写，留白
 - 选择 'Just the push event'
 - 勾选 'Active'
 ![image](https://github.com/daisy-ycguo/devopslab/blob/master/images/create-webhook.png)

#### 5 修改hello.go source code并push
Push操作发生时，webhook会发送一个POST请求到listener的endpoint,从而出发一个pipeline run。   
更新您自己的devopslab repo中的 [src/app/hello.go](../src/app/hello.go)文件，更新应用的输出。   
例如，第17行修改为 `fmt.Fprintf(w, "%s\n", say("GREEN-IBM!!!"))`   
您可以在github UI上直接更新并提交。
 ![image](https://github.com/daisy-ycguo/devopslab/blob/master/images/git-update.png)
  ![image](https://github.com/daisy-ycguo/devopslab/blob/master/images/git-commit.png)

### 6	查看webhook的变化
观察你的github repo webhook的变化，有新的delivery产生，response应该为200。
![image](https://github.com/daisy-ycguo/devopslab/blob/master/images/webhook-deliveries.png)

### 7 验证Pipeline运行成功
在CloudShell中请执行命令：
```
kubectl get pipelinerun
```

期待输出：
```
NAME             SUCCEEDED   REASON      STARTTIME   COMPLETIONTIME
hello-pr-njq8h   True        Succeeded   9m23s       8m19s           <-上个实验(exercise-1)产生的pipelinerun
hello-pr-xk29f   True        Succeeded   62s         1s              <-这次产生的pipelinerun
```
在CloudShell中请执行命令：
```
kubectl get taskrun
```

期待输出：
```
NAME                                     SUCCEEDED   REASON      STARTTIME   COMPLETIONTIME
hello-pr-njq8h-deploy-to-cluster-dtmlz   True        Succeeded   9m54s       9m46s <-上个实验(exercise-1)产生的taskrun
hello-pr-njq8h-source-to-image-vrx4x     True        Succeeded   10m         9m54s <-上个实验(exercise-1)产生的taskrun
hello-pr-xk29f-deploy-to-cluster-mtk92   True        Succeeded   98s         87s <-这次产生的taskrun
hello-pr-xk29f-source-to-image-tzvg2     True        Succeeded   2m28s       98s <-这次产生的taskrun
```
在CloudShell中请执行命令：
```
kubectl get pods | grep hello
```

期待输出：
```
hello-pr-njq8h-deploy-to-cluster-dtmlz-pod-e1eb68   0/3     Completed   0          10m <-上个实验(exercise-1)产生的pod
hello-pr-njq8h-source-to-image-vrx4x-pod-a6716a     0/2     Completed   0          11m <-上个实验(exercise-1)产生的pod
hello-pr-xk29f-deploy-to-cluster-mtk92-pod-6ab995   0/3     Completed   0          2m26s <-这次产生的pod
hello-pr-xk29f-source-to-image-tzvg2-pod-8c1a39     0/2     Completed   0          3m17s <-这次产生的pod
```

## 8 查看service被更新了
在CloudShell中请执行命令：
```
kubectl get ksvc
```

期待输出：
```
NAME    URL                                                                      LATESTCREATED   LATESTREADY   READY   REASON
hello   http://hello-default.capacity-demo.us-south.containers.appdomain.cloud   hello-ldtj8     hello-ldtj8   True  

 curl http://hello-default.capacity-demo.us-south.containers.appdomain.cloud
Hello world, this is GREEN-IBM!!!
```

## 9 清理已完成的pipelinerun
查看之前实验中的pipelinerun都是Succeeded状态。
```
kubectl get pr
```
清理pipelinerun记录。
```
kubectl delete pr --all
```

恭喜您！您已经完成了Tekton全部实验。下面继续阅读，或者进行下一步[使用Knative Eventing监控新服务](../02-knative/00-eventing.md)

## 了解发生了什么
一个PipelineRun被创建出来，这个PipelineRun执行了上个实验中我们创建好的pipeline 'build-and-deploy-pipeline'。
```
$ kubectl describe pr hello-pr-xk29f
Name:         hello-pr-xk29f
Namespace:    default
Labels:       tekton.dev/eventlistener=my-listener
              tekton.dev/pipeline=build-and-deploy-pipeline
              tekton.dev/triggers-eventid=rr7kk
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"tekton.dev/v1alpha1","kind":"Pipeline","metadata":{"annotations":{},"name":"build-and-deploy-pipeline","namespace":"default...
API Version:  tekton.dev/v1alpha1
Kind:         PipelineRun
Metadata:
  Creation Timestamp:  2019-12-05T05:52:51Z
  Generate Name:       hello-pr-
  Generation:          1
  Resource Version:    6643231
  Self Link:           /apis/tekton.dev/v1alpha1/namespaces/default/pipelineruns/hello-pr-xk29f
  UID:                 795127cd-1723-11ea-bcdf-4ae22433ba96
Spec:
  Params:
    Name:   pathToYamlFile
    Value:  src/tekton/basic/knative/hello.yaml
    Name:   imageUrl
    Value:  us.icr.io/liqiujie/hello
    Name:   imageTag
    Value:  2.0
  Pipeline Ref:
    Name:  build-and-deploy-pipeline
  Resources:
    Name:  git-source
    Resource Ref:
      Name:         hello-git-wjzj2
  Service Account:  pipeline-account
  Timeout:          1h0m0s
Status:
  Completion Time:  2019-12-05T05:53:52Z
  Conditions:
    Last Transition Time:  2019-12-05T05:53:52Z
    Message:               All Tasks have completed executing
    Reason:                Succeeded
    Status:                True
    Type:                  Succeeded
  Start Time:              2019-12-05T05:52:51Z
  ....
```
一个PipelineResource被创建出来了，其url参数的值是webhook发出的POST request的body里面提供的。
```
$ kubectl describe pipelineresource hello-git-wjzj2
Name:         hello-git-wjzj2
Namespace:    default
Labels:       tekton.dev/eventlistener=my-listener
              tekton.dev/triggers-eventid=rr7kk
Annotations:  <none>
API Version:  tekton.dev/v1alpha1
Kind:         PipelineResource
Metadata:
  Creation Timestamp:  2019-12-05T05:52:51Z
  Generation:          1
  Resource Version:    6642998
  Self Link:           /apis/tekton.dev/v1alpha1/namespaces/default/pipelineresources/hello-git-wjzj2
  UID:                 794a8faa-1723-11ea-bcdf-4ae22433ba96
Spec:
  Params:
    Name:   revision
    Value:  47a4a0d497095c93ca9b076b080712e40a072828
    Name:   url
    Value:  https://github.com/QiuJieLi/devopslab
  Type:     git
Events:     <none>
```
## 问题诊断

### Webhook Recent Deliveries response 501
查看ingress HOSTS是否可以被访问
### Webhook Recent Deliveries response 200但是没有pipeline run生成
查看listener的pod
```
$ kubectl get pod | grep el-my-listener
$ kubectl logs el-my-listener-99b595cc6-4vqq6
2019/12/04 10:19:59 Error getting TriggerBinding my-pipeline-binding: triggerbindings.tekton.dev "my-pipeline-binding" not found
```
### 有pipeline run生成但是pipeline run失败
参考[tekton exercise-1](./01-exercise-1.md)的问题诊断
