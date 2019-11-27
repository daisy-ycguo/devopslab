# 配置Tekton Trigger

## 1.Fork 开源Trigger 项目到自己的repo并clone到local workstation
https://github.com/tektoncd/triggers

## 2.配置Trigger
按如下步骤配置Trigger：
```
$ kubectl apply -f triggertemplates/triggertemplate.yaml
triggertemplate.tekton.dev/my-pipeline-template created

$ kubectl apply -f triggerbindings/triggerbinding.yaml
triggerbinding.tekton.dev/my-pipeline-binding created

$ kubectl apply -f eventlisteners/eventlistener.yaml
eventlistener.tekton.dev/my-listener created

$ kubectl apply -f example-pipeline.yaml
pipeline.tekton.dev/build-and-deploy-pipeline configured
task.tekton.dev/source-to-image configured
task.tekton.dev/deploy-using-kubectl configured
```

然后获取你的集群的ingress subdomain name：
```
$ ibmcloud ks cluster-get testcluster
Retrieving cluster testcluster...
OK

Name: testcluster
ID: bmvtemed0lrb6d8hv0n0
State: normal
……
Ingress Subdomain: testcluster-973348.us-south.containers.appdomain.cloud
```

然后更新myexample目录下的ingress.yaml文件：

再apply ingress文件:
```
$ kubectl apply -f myexample/ingress.yaml
ingress.extensions/el-my-listener created
```

## 3.验证Trigger配置成功
```
$ Kubectl get pr 
Verify listener pod
$ Kubectl get pod
$ Kubectl log <listener-pod-name>
$ Kubectl get tr
$ kubectl get ksvc picalc

## 4.进入在Tekton实验中fork的你的github上的tekton-tutorial repo，并配置这个repo的webhook
 
## 5. 在你的tekton-tutorial 对source code做些修改并push 

```
$ vi src/picalc.go
$ git status
$ git add src/picalc.go
$ git commit -m "first change"
$ git push
```
 
## 6.验证Tekton pipeline 运行成功
```
$ Kubectl get pr 
$ Kubectl get tr
$ kubectl get ksvc picalc
```
注意：如果以上步骤显示有失败的情况，尝试下面的方法进行诊断：
```
$ Kubectl get pod
$ Kubectl log <listener-pod-name>
```
## 7.	监控webhook的变化


 
