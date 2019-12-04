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

## 实验步骤
下面的实验中，我们将使用Trigger来创建一个PipelineRun和一个PipelineResource。这个PipelineRun运行了我们[tekton实验](https://github.com/daisy-ycguo/devopslab/blob/master/01-tekton/exercise-1.md)中创建的pipeline - "build-and-deploy-pipeline"。

### 1. 按如下步骤配置Trigger：
1.1 创建service account
```
$ kubectl apply -f devopslab/src/tekton/trigger/role-resources
rolebinding.rbac.authorization.k8s.io/tekton-triggers-example-binding created
role.rbac.authorization.k8s.io/tekton-triggers-example-minimal created
secret/githubsecret created
serviceaccount/tekton-triggers-example-sa created
```
1.2 更新devopslab/src/tekton/trigger/triggertemplates/triggertemplate.yaml，将PipelineRun中的参数imageUrl的值<REGISTRY>/<NAMESPACE>/hello替换。请参考[tekton实验](https://github.com/daisy-ycguo/devopslab/blob/master/01-tekton/exercise-1.md)步骤5.1   
  
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
          value: <REGISTRY>/<NAMESPACE>/hello
        - name: imageTag
          value: "1.0"
      serviceAccount: pipeline-account
```
1.3 创建TriggerTemplate   
```
$ kubectl apply -f devopslab/src/tekton/trigger/triggertemplates/triggertemplate.yaml
triggertemplate.tekton.dev/my-pipeline-template created
```
1.4 创建TriggerBinding   
```
$ kubectl apply -f devopslab/src/tekton/trigger/triggerbindings/triggerbinding.yaml
triggerbinding.tekton.dev/my-pipeline-binding created
```
1.5 创建EventListener   
```
$ kubectl apply -f devopslab/src/tekton/trigger/eventlisteners/eventlistener.yaml
eventlistener.tekton.dev/my-listener created
```

#### 2 检查我们所需要的pods和services已经建好并且状态健康
```
$ kubectl get svc
NAME                 TYPE           CLUSTER-IP      EXTERNAL-IP                                           PORT(S)             AGE
el-my-listener       ClusterIP      172.21.142.64   <none>                                                8080/TCP            4s

$ kubectl get pods
NAME                                                 READY   STATUS      RESTARTS   AGE
el-my-listener-99b595cc6-4vqq6                          1/1     Running     0          21s
```

#### 3 配置Ingress
使得listner endpoint可以被从cluster外部访问。后面我们会通过git repository的webhook来访问这个listener endpoint。
1. 获取你的集群的Ingress Subdomain
```
$ ibmcloud ks cluster-get testcluster | grep 'Ingress Subdomain'

Ingress Subdomain: testcluster-973348.us-south.containers.appdomain.cloud
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
```
$ kubectl apply -f devopslab/src/tekton/trigger/ingress.yaml
ingress.extensions/el-my-listener created
```
4. 查看ingress
```
$ kubectl get ingress
NAME             HOSTS                                                              ADDRESS         PORTS   AGE
el-my-listener   el-my-listener.testcluster-973348.us-south.us-south.containers.appdomain.cloud   169.47.66.178   80      14s
```


#### 2.4 配置webhook
当指定的event发生时，Webhook会发送一个POST请求到其配置的URL。这个URL就是我们上面建好的listener的endpoint。
进入你在[Tekton实验1](https://github.com/daisy-ycguo/devopslab/blob/master/01-tekton/exercise-1.md)实验步骤1中fork到您自己的git账户下的repo，配置这个repo的webhook。   
1. 在浏览器中您的devopslab repo https://github.com/<your-git-account>/devopslab  
2. 点击最右侧的Settings tab，从左侧导航栏选择Webhooks   
3. 点击'Add webhook'按钮   
4. Add webhook
 其中：
 - Payload URL: http://加上步骤2.4中配置的ingress的HOSTS 例如：http://el-my-listener.testcluster-973348.us-south.us-south.containers.appdomain.cloud
 - Content type: application/json
 - Secret: 任意输入一个密码
 - 选择 'Just the push event'
 - 勾选 'Active'
 ![image](https://github.com/daisy-ycguo/devopslab/blob/master/images/create-webhook.png)

#### 2.5 修改hello.go source code并push
Push操作发生时，webhook会发送一个POST请求到listener的endpoint,从而出发一个pipeline run。
```
$ vi devopslab/src/app/hello.go
e.g. 修改为 fmt.Fprintf(w, "%s\n", say("BLUE-XiaoMing!!!"))
$ cd devopslab
$ git status
$ git add src/app/hello.go
```
准备
```
$ git config --global user.email "you@example.com"
$ git config --global user.name "Your Name"
$ ssh-keygen -t rsa -b 4096 -C "your_email@example.com"
```
拷贝cat ~/.ssh/id_rsa.pub的内容，在您的git account下创建sshkey。参考[为github account添加ssh key](https://help.github.com/en/enterprise/2.19/user/authenticating-to-github/adding-a-new-ssh-key-to-your-github-account)

```
$ git commit -m "first change"
$ git push
```

## 3.	查看webhook的变化
观察你的github repo webhook的变化，有新的delivery产生，response应该为200。
![image](https://github.com/daisy-ycguo/devopslab/blob/master/images/webhook-deliveries.png)

## 4.验证Pipeline运行成功
```
$ kubectl get pipelinerun
NAME             SUCCEEDED   REASON      STARTTIME   COMPLETIONTIME
hello-pr-bczph   True        Succeeded   5m20s       4m18s
$ kubectl get taskrun
NAME                                     SUCCEEDED   REASON      STARTTIME   COMPLETIONTIME
hello-pr-bczph-deploy-to-cluster-57zsv   True        Succeeded   4m56s       4m47s
hello-pr-bczph-source-to-image-fv5rl     True        Succeeded   5m49s       4m56s
$ kubectl get pod
hello-pr-bczph-deploy-to-cluster-57zsv-pod-f53367   0/3     Completed   0          5m10s
hello-pr-bczph-source-to-image-fv5rl-pod-7936e2     0/2     Completed   0          6m3s
```

## 5. 查看service被更新了
```
$ kubectl get ksvc
NAME    URL                                                                      LATESTCREATED   LATESTREADY   READY   REASON
hello   http://hello-default.capacity-demo.us-south.containers.appdomain.cloud   hello-jqznl     hello-jqznl   True   
curl http://hello-default.capacity-demo.us-south.containers.appdomain.cloud
[ 20191031 ] Hello world, this is BLUE-xiaoming!!!
```

## 6. 了解发生了什么
一个PipelineResource被创建出来了，其url参数的值是webhook发出的POST request的body里面提供的。
```
$ kubectl get pipelineresource
hello-git-zmfgt   

$ kubectl describe pipelineresource hello-git-zmfgt
Name:         hello-git-zmfgt
Namespace:    default
Labels:       tekton.dev/eventlistener=my-listener
              tekton.dev/triggers-eventid=vvlrg
Annotations:  <none>
API Version:  tekton.dev/v1alpha1
Kind:         PipelineResource
Metadata:
  Creation Timestamp:  2019-12-04T09:27:16Z
  Generation:          1
  Resource Version:    6495867
  Self Link:           /apis/tekton.dev/v1alpha1/namespaces/default/pipelineresources/hello-git-zmfgt
  UID:                 43396bbb-1678-11ea-8ad1-ae88ce260b9c
Spec:
  Params:
    Name:   revision
    Value:  4a544c9ca1bad5c14d2aa46d973205b183d833c7
    Name:   url
    Value:  https://github.com/<your-git-account>/devopslab
  Type:     git
Events:     <none>
```
一个PipelineRun被创建出来，使用了resource和指定的Tekton Pipeline。这个PipelineRun执行了simple-pipeline中定义的build-and-deploy-pipeline中的两个个task。
```
$ kubectl describe pr hello-pr-bczph
Name:         hello-pr-bczph
Namespace:    default
Labels:       tekton.dev/eventlistener=my-listener
              tekton.dev/pipeline=build-and-deploy-pipeline
              tekton.dev/triggers-eventid=9x726
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"tekton.dev/v1alpha1","kind":"Pipeline","metadata":{"annotations":{},"name":"build-and-deploy-pipeline","namespace":"default...
API Version:  tekton.dev/v1alpha1
Kind:         PipelineRun
Metadata:
  Creation Timestamp:  2019-12-04T09:32:51Z
  Generate Name:       hello-pr-
  Generation:          1
  Resource Version:    6497002
  Self Link:           /apis/tekton.dev/v1alpha1/namespaces/default/pipelineruns/hello-pr-bczph
  UID:                 0a988731-1679-11ea-8ad1-ae88ce260b9c
Spec:
  Params:
    Name:   pathToYamlFile
    Value:  src/tekton/basic/knative/hello.yaml
    Name:   imageUrl
    Value:  us.icr.io/liqiujie/hello
    Name:   imageTag
    Value:  1.0
  Pipeline Ref:
    Name:  build-and-deploy-pipeline
  Resources:
    Name:  git-source
    Resource Ref:
      Name:         hello-git-b88f5
  Service Account:  pipeline-account
  Timeout:          1h0m0s
Status:
  Completion Time:  2019-12-04T09:33:53Z
  Conditions:
    Last Transition Time:  2019-12-04T09:33:53Z
    Message:               All Tasks have completed executing
    Reason:                Succeeded
    Status:                True
    Type:                  Succeeded
  Start Time:              2019-12-04T09:32:51Z
  Task Runs:
    hello-pr-bczph-deploy-to-cluster-57zsv:
      Pipeline Task Name:  deploy-to-cluster
      Status:
        Completion Time:  2019-12-04T09:33:53Z
        Conditions:
          Last Transition Time:  2019-12-04T09:33:53Z
          Message:               All Steps have completed executing
          Reason:                Succeeded
          Status:                True
          Type:                  Succeeded
        Pod Name:                hello-pr-bczph-deploy-to-cluster-57zsv-pod-f53367
        Start Time:              2019-12-04T09:33:44Z
        Steps:
          Name:  update-yaml
          Terminated:
            Container ID:  containerd://13006e87b8de41fbf968d137af661c2a585f6f24d0a5413aec49f9ea1b8603ef
            Exit Code:     0
            Finished At:   2019-12-04T09:33:51Z
            Reason:        Completed
            Started At:    2019-12-04T09:33:48Z
          Name:            run-kubectl
          Terminated:
            Container ID:  containerd://a26049d6e66979c033715ee6c4c8bad0b0f9c1a9c69b7563cf8e46ca5ca27624
            Exit Code:     0
            Finished At:   2019-12-04T09:33:52Z
            Reason:        Completed
            Started At:    2019-12-04T09:33:48Z
          Name:            git-source-hello-git-b88f5-lhkt4
          Terminated:
            Container ID:  containerd://77502bdc5e8a2679d749bb913ae7ac46cf414c0f54c247de53d21be1ffbeeae0
            Exit Code:     0
            Finished At:   2019-12-04T09:33:51Z
            Reason:        Completed
            Started At:    2019-12-04T09:33:47Z
    hello-pr-bczph-source-to-image-fv5rl:
      Pipeline Task Name:  source-to-image
      Status:
        Completion Time:  2019-12-04T09:33:44Z
        Conditions:
          Last Transition Time:  2019-12-04T09:33:44Z
          Message:               All Steps have completed executing
          Reason:                Succeeded
          Status:                True
          Type:                  Succeeded
        Pod Name:                hello-pr-bczph-source-to-image-fv5rl-pod-7936e2
        Start Time:              2019-12-04T09:32:51Z
        Steps:
          Name:  build-and-push
          Terminated:
            Container ID:  containerd://3db9c038028012c10c1e8d0a6e81470f07a538103680928fb39ec5e1d791f838
            Exit Code:     0
            Finished At:   2019-12-04T09:33:43Z
            Reason:        Completed
            Started At:    2019-12-04T09:32:55Z
          Name:            git-source-hello-git-b88f5-59g6m
          Terminated:
            Container ID:  containerd://de9f0fcb3c9e2bfb58a7c18945cdbed4d55799d07ec5b06903885e75b1e5f123
            Exit Code:     0
            Finished At:   2019-12-04T09:32:58Z
            Reason:        Completed
            Started At:    2019-12-04T09:32:55Z
Events:
  Type     Reason             Age   From                 Message
  ----     ------             ----  ----                 -------
  Warning  PipelineRunFailed  10m   pipeline-controller  PipelineRun failed to update labels/annotations
  Normal   Succeeded          9m5s  pipeline-controller  All Tasks have completed executing

```

## 问题诊断
### Webhook Recent Deliveries response 501
查看ingress HOSTS是否可以被访问
### Webhook Recent Deliveries response 200但是没有pipeline run生成
查看listener的pod
```
kubectl get pod | grep el-my-listener
kubectl logs el-my-listener-99b595cc6-4vqq6
```
### 有pipeline run生成但是pipeline run失败，参考[tekton实验](https://github.com/daisy-ycguo/devopslab/blob/master/01-tekton/exercise-1.md)的问题诊断
