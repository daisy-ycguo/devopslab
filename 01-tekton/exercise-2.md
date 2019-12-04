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

### 1.Fork devopslab项目到自己的repo并clone到local workstation
https://github.com/daisy-ycguo/devopslab.git

### 2. 创建实验的资源

#### 2.1 按如下步骤配置Trigger：
```
$ cd devopslab/src/tekton/trigger
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

#### 2.2 检查我们所需要的pods和services已经建好并且状态健康
```
$ kubectl get svc
NAME                 TYPE           CLUSTER-IP      EXTERNAL-IP                                           PORT(S)             AGE
el-my-listener       ClusterIP      172.21.142.64   <none>                                                8080/TCP            4s

$ kubectl get pods
NAME                                                 READY   STATUS      RESTARTS   AGE
el-my-listener-99b595cc6-4vqq6                          1/1     Running     0          21s
```

#### 2.3 配置Ingress
使得listner endpoint可以被从cluster外部访问。后面我们会通过git repository的webhook来访问这个listener endpoint。
1. 获取你的集群的Ingress Subdomain
```
$ ibmcloud ks cluster-get testcluster | grep 'Ingress Subdomain'

Ingress Subdomain: testcluster-973348.us-south.containers.appdomain.cloud
```

2. 更新ingress.yaml文件
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
$ kubectl apply -f ingress.yaml
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
1. 在浏览器中您的tekton-tutorial repo https://github.com/<your name>/tekton-tutorial   
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

#### 2.5 修改source code并push
Push操作这个event发生时，webhook会发送一个POST请求到listener的endpoint,从而出发一个pipeline run。
```
$ vi src/picalc.go
$ git status
$ git add src/picalc.go
$ git commit -m "first change"
$ git push
```
[为github account添加ssh key](https://help.github.com/en/enterprise/2.19/user/authenticating-to-github/adding-a-new-ssh-key-to-your-github-account)

## 3.	查看webhook的变化
观察你的github repo webhook的变化，有新的delivery产生。
![image](https://github.com/daisy-ycguo/devopslab/blob/master/images/webhook-deliveries.png)

## 4.验证Pipeline运行成功
```
$ kubectl get pipelinerun
$ kubectl get taskrun
$ kubectl get pod
```

## 5. 了解发生了什么
一个PipelineResource被创建出来了，其url参数的值是webhook发出的POST request的body里面提供的。
```
$ kubectl get pipelineresource xxx -o yaml

```
一个PipelineRun被创建出来，使用了resource和指定的Tekton Pipeline。这个PipelineRun执行了simple-pipeline中定义的build-and-deploy-pipeline中的两个个task。
```
$ kubectl describe pr xxx

```
