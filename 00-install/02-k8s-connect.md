# 准备Kubernetes集群环境

Knative Lab使用了IBM公有云上的Kubernetes集群，以及一个云上的命令行窗口CloudShell。您只需要拥有IBM Cloud的注册账号，就可以进行下面的操作。

## 前提

* 拥有一个IBM Cloud账号，也被称为IBM ID。如果没有注册，请到[http://cloud.ibm.com](http://cloud.ibm.com)上注册。
* 准备一个可以联网的浏览器，推荐Chrome，Firefox，和Safari。

## 第一步：分配Kuberntes集群

我们预先为这次实验创建了若干个多节点的Kubernetes集群，请您到IBM工作人员那里分配一个Kubernetes集群。 分配到集群后，请记住您的集群的名字。

通过浏览器，打开您github代码库下面的文件：[src/setenv.sh](../src/setenv.sh)，直接点击编辑按钮，把您的集群名称填写在`export MYCLUSTER=`后面，然后commit。

```
export GITACCOUNT=<your_account>

export MYCLUSTER=tektonknativebeijing66
export KUBECONFIG=

export NAMESPACE=tektondevops-<your_name>
export EMAIL=<your_email>
```

## 第二步：准备CloudShell &  连接到IBM Account下的Kubernetes集群

一，在CloudShell页面中，点击右上角切换账户，在弹出的下拉框里面，选择*IBM*。 

![](https://github.com/daisy-ycguo/knativelab/raw/master/images/cloudshell-account.png)


二，账户切换后， 原先的cloud shell terminal 窗口会关闭， 请重新点击*命令行窗口图标*， 创建新的命令行窗口。 

三，您领取到的集群名称大约为`tektonknativebeijing**`，其中`**`部分为您的集群编号，如`tektonknativebeijing66`。把这个集群名称记录在环境变量中。

   在CloudShell页面，输入：
   ```text
   export MYCLUSTER=<your_cluster_name>
   ```

四，获取你的集群的更多信息：

运行命令：
```text
ibmcloud ks cluster-get $MYCLUSTER
```

期待输出：
```
Retrieving cluster knative-guoyc...
OK

Name:                           knative-guoyc
ID:                             c6e0aec577364c6faa3f1a68596bc986
State:                          normal
Created:                        2019-06-20T03:08:12+0000
Location:                       syd01
Master URL:                     https://c2.jp-tok.containers.cloud.ibm.com:30425
Public Service Endpoint URL:    https://c2.jp-tok.containers.cloud.ibm.com:30425
Private Service Endpoint URL:   -
Master Location:                Sydney
Master Status:                  Ready (1 hour ago)
Master State:                   deployed
Master Health:                  normal
Ingress Subdomain:              knative-guoyc.jp-tok.containers.appdomain.cloud
Ingress Secret:                 knative-guoyc
Workers:                        2
Worker Zones:                   syd01
Version:                        1.13.7_1526
Owner:                          guoyingc@cn.ibm.com
Monitoring Dashboard:           -
Resource Group ID:              2a926a9173174d94a6eb13284e089f88
Resource Group Name:            default
```

***注意*** 如果返回错误`The specified cluster could not be found.`，请检查
- CloudShell右上角，用户名那里是否换到为`IBM`
- 集群的名字是否正确

六，下载你的集群的配置文件到CloudShell终端：

   运行命令：
   ```text
   ibmcloud ks cluster config $MYCLUSTER
   ```
   期待输出：
   ```
   OK
   The configuration for kubeconsh-guoyc was downloaded successfully.
   
   Export environment variables to start using Kubernetes.
   
   export KUBECONFIG=/usr/shared-data/cloud-ibm-com-47b84451ab70b94737518f7640a9ee42-1/.bluemix/plugins/container-service/clusters/kubeconsh-guoyc/kube-config-syd01-kubeconsh-guoyc.yml
   ```

七，上面一条命令输出的最后一行是黄色高亮的export命令，在CloudShell中拷贝该命令，并黏贴执行：

   ```text
   export KUBECONFIG=/usr/shared-data/cloud-ibm-com-47b84451ab70b94737518f7640a9ee42-1/.bluemix/plugins/container-service/clusters/......
   ```

八，验证您已经可以用kubectl连接到云端的Kubernetes集群：

   运行命令：
   ```text
   kubectl get nodes
   ```
   期待输出：
   ```
   NAME             STATUS   ROLES    AGE     VERSION
   10.138.173.77   Ready    <none>   112m   v1.13.7+IKS
   10.138.173.88   Ready    <none>   112m   v1.13.7+IKS
   ```

这里，`kubectl get nodes`能够得到正确返回，看到您的集群中的节点，那么您就可以继续下面的实验了。

九，记录环境变量。

通过浏览器，打开您github代码库下面的文件：[src/setenv.sh](../src/setenv.sh)，直接点击编辑按钮。   
- 把您的KUBECONFIG所在的路径填写到`export KUBECONFIG=`后面。   
- 把第四步命令输出中的Ingress Subdomain内容填写在`INGRESS=`后面。   
然后commit。至此，您已经完成了所有环境变量的配置。   

```
export GITACCOUNT=<your_account>

export MYCLUSTER=tektonknativebeijing**
export KUBECONFIG=/usr/shared-data/cloud-ibm-com-47b84451ab70b94737518f7640a9ee42-1/.bluemix/plugins/container-service/clusters/kubeconsh-guoyc/kube-config-syd01-kubeconsh-guoyc.yml

export NAMESPACE=tektonknativebeijing**
export EMAIL=<your_email>
export INGRESS=knative-guoyc.jp-tok.containers.appdomain.cloud
```

继续 [安装Istio和Knative](./03-istio-knative-install.md).

