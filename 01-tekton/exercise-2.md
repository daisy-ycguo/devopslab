# Tekton Trigger

## 实验目标

- 了解Trigger的基本概念
- 创建TriggerTemplate, TriggerBinding和EventListener
- 创建一个pipeline
- 为git repo创建一个webhook,当有push操作发生的时候，发送请求到EventListener来创建pipeline resource和pipeline run,执行指定的pipeline。

## 基本概念

- TriggerTemplate - 资源模板 例如：PipelineResources和PipelineRun
- TriggerBinding - 检验events并且抽取payload中的fields
- EventListener - 将TriggerBindings和TriggerTemplates连接起来，提供一个可访问的endpoint (事件接收器). 它使用TriggerBinding从events中抽取出来的内容作为参数，来创建TriggerTemplate中指定的资源。

## 实验准备
1.Fork 开源Trigger 项目到自己的repo并clone到local workstation
https://github.com/tektoncd/triggers

## 实验步骤
下面的实验中，我们将使用Trigger来创建一个PipelineRun和一个PipelineResource。这个PipelineRun克隆了一个GitHub repository并打印一些信息。

### 2.1 创建实验的资源

按如下步骤配置Trigger：
```
$ cd examples
$ kubectl apply -f role-resources
rolebinding.rbac.authorization.k8s.io/tekton-triggers-example-binding created
role.rbac.authorization.k8s.io/tekton-triggers-example-minimal created
secret/githubsecret created
serviceaccount/tekton-triggers-example-sa created

$ kubectl apply -f triggertemplates/triggertemplate.yaml
triggertemplate.tekton.dev/my-pipeline-template created

$ kubectl apply -f triggerbindings/triggerbinding.yaml
triggerbinding.tekton.dev/my-pipeline-binding created

$ kubectl apply -f eventlisteners/eventlistener.yaml
eventlistener.tekton.dev/my-listener created

```

### 2.2 检查我们所需要的pods和services已经建好并且状态健康
```
$ kubectl get svc
NAME                 TYPE           CLUSTER-IP      EXTERNAL-IP                                           PORT(S)             AGE
el-listener          ClusterIP      172.21.86.19    <none>                                                8080/TCP            3s

$ kubectl get pods
NAME                                                 READY   STATUS      RESTARTS   AGE
el-listener-5f98f8cdcd-nrd4l                         1/1     Running     0          18s
```

### 2.3 Apply pipeline和tasks
pipeline和tasks被TriggerTemplate中的piepeline引用。
```
$ kubectl apply -f example-pipeline.yaml
pipeline.tekton.dev/build-and-deploy-pipeline configured
task.tekton.dev/source-to-image configured
task.tekton.dev/deploy-using-kubectl configured
```

### 2.4 配置Ingress
使得listner endpoint可以被从cluster外部访问。后面我们会通过git repository的webhook来访问这个listener endpoint。
1. 获取你的集群的Ingress Subdomain
```
$ ibmcloud ks cluster-get testcluster | grep 'Ingress Subdomain'

Ingress Subdomain: testcluster-973348.us-south.containers.appdomain.cloud
```

2. 更新myexample目录下的ingress.yaml文件
- 将host的值替换为el-listener.<Ingress Subdomain 的值>， 例如 el-listener.testcluster-973348.us-south.containers.appdomain.cloud    
- serviceName替换为2.2步骤里的service name, 即el-listener
- metadata name也替换为el-listener
```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: el-listener
  namespace: default
spec:
  rules:
  - host: el-listener.testcluster-973348.us-south.containers.appdomain.cloud
    http:
      paths:
      - backend:
          serviceName: el-listener
          servicePort: 8080
        path: /
```
3. Apply ingress文件:
```
$ kubectl apply -f myexample/ingress.yaml
ingress.extensions/el-listener created
```
4. 查看ingress
```
$ k get ing
NAME             HOSTS                                                              ADDRESS         PORTS   AGE
el-listener   el-listener.capacity-demo.us-south.containers.appdomain.cloud   169.47.66.178   80      60s
```


### 2.5 配置webhook
当指定的event发生时，Webhook会发送一个POST请求到下面配置的URL。这个URL就是我们上面建好的listener的endpoint。
进入你在[Tekton实验1](https://github.com/daisy-ycguo/devopslab/blob/master/01-tekton/exercise-1.md)实验步骤1中fork到您自己的git账户下的repo，配置这个repo的webhook。   
1. 在浏览器中您的tekton-tutorial repo https://github.com/<your name>/tekton-tutorial   
2. 点击最右侧的Settings tab，从左侧导航栏选择Webhooks   
3. 点击'Add webhook'按钮   
4. Add webhook
 其中：
 - Payload URL: http://加上步骤2.4中配置的ingress的HOSTS
 - Content type: application/json
 - Secret: 任意输入一个密码
 - 选择 'Just the push event'
 - 勾选 'Active'
 ![image](https://github.com/daisy-ycguo/devopslab/blob/master/images/create-webhook.png)

### 2.6 修改source code并push
Push操作这个event发生时，webhook会发送一个POST请求到listener的endpoint,从而出发一个pipeline run。
```
$ vi src/picalc.go
$ git status
$ git add src/picalc.go
$ git commit -m "first change"
$ git push
```
[为github account添加ssh key](https://help.github.com/en/enterprise/2.19/user/authenticating-to-github/adding-a-new-ssh-key-to-your-github-account)

## 3.验证Pipeline运行成功
```
$ kubectl get pipelinerun
simple-pipeline-runrzbt8   True        Succeeded            3m14s       2m46s
$ kubectl get taskrun
simple-pipeline-runrzbt8-say-bye-4dmtx       True        Succeeded   3m11s       3m3s
simple-pipeline-runrzbt8-say-hello-tmtlc     True        Succeeded   3m30s       3m22s
simple-pipeline-runrzbt8-say-message-bgzvb   True        Succeeded   3m21s       3m12s
$ kubectl get pod
simple-pipeline-runrzbt8-say-bye-4dmtx-pod-e2cd63       0/2     Completed   0          3m33s
simple-pipeline-runrzbt8-say-hello-tmtlc-pod-e8baa2     0/2     Completed   0          3m53s
simple-pipeline-runrzbt8-say-message-bgzvb-pod-d6cac1   0/2     Completed   0          3m43s
```
检查每个task的pod的log
```
$ kubectl logs logs simple-pipeline-runrzbt8-say-bye-4dmtx-pod-e2cd63 --all-containers
...
Goodbye Triggers!
$ kubectl logs simple-pipeline-runrzbt8-say-hello-tmtlc-pod-e8baa2 --all-containers
...
Hello Triggers!
$ kubectl logs simple-pipeline-runn4qps-say-bye-7xbk2-pod-116608  --all-containers
...
$(inputs.params.message)
```
 
## 4.	查看webhook的变化
观察你的github repo webhook的变化，有新的delivery产生。
![image](https://github.com/daisy-ycguo/devopslab/blob/master/images/webhook-deliveries.png)

## 5. 发生了什么
一个PipelineResource被创建出来了，其url参数的值是webhook发出的POST request的body里面提供的。
```
$ kubectl get pipelineresource git-source-v5tml -o yaml
apiVersion: tekton.dev/v1alpha1
kind: PipelineResource
metadata:
  creationTimestamp: 2019-12-03T08:07:05Z
  generation: 1
  labels:
    tekton.dev/eventlistener: listener
    tekton.dev/triggers-eventid: 4xk2x
  name: git-source-v5tml
  namespace: default
  resourceVersion: "6316396"
  selfLink: /apis/tekton.dev/v1alpha1/namespaces/default/pipelineresources/git-source-v5tml
  uid: e548d761-15a3-11ea-8937-167b3451333f
spec:
  params:
  - name: revision
    value: aa7bac85bd56bf5f8a40e5a887b43de39bee6f67
  - name: url
    value: https://github.com/QiuJieLi/tekton-tutorial
  type: git
```
一个PipelineRun被创建出来，使用了resource和指定的Tekton Pipeline:simple-pipeline。这个PipelineRun执行了simple-pipeline中定义的三个task。
```
$ kubectl describe pr simple-pipeline-runrzbt8
...
Spec:
  Params:
    Name:   message
    Value:  Hello from the Triggers EventListener!
    Name:   contenttype
    Value:  application/json
  Pipeline Ref:
    Name:  simple-pipeline
  Resources:
    Name:  git-source
    Resource Ref:
      Name:         git-source-v5tml
  Service Account:
  Timeout:          1h0m0s
Status:
  Completion Time:  2019-12-03T08:07:33Z
  Conditions:
    Last Transition Time:  2019-12-03T08:07:33Z
    Message:               All Tasks have completed executing
    Reason:                Succeeded
    Status:                True
    Type:                  Succeeded
  Start Time:              2019-12-03T08:07:05Z
  Task Runs:
  ...
  ```
