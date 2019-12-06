# 准备Kubernetes集群环境

Knative Lab使用了IBM公有云上的Kubernetes集群，以及一个云上的命令行窗口CloudShell。您只需要拥有IBM Cloud的注册账号，就可以进行下面的操作。

## 前提

* 拥有一个IBM Cloud账号，也被称为IBM ID。如果没有注册，请到[http://cloud.ibm.com](http://cloud.ibm.com)上注册。
* 准备一个可以联网的浏览器，推荐Chrome，Firefox，和Safari。

## 第一步：分配Kuberntes集群

我们预先为这次实验创建了若干个多节点的Kubernetes集群，请您到IBM工作人员那里分配一个Kubernetes集群。 分配到集群后，请记住您的集群的名字。

## 第二步：获取IBM Cloud API KEY。 

在IBM cloud中， 您的用户可以被分配到不同的`Account`下， 在同一`Account`内部， Kubernetes集群和对应的IBM Cloud Resgitry 会进行自动关联； 

但在这个实验中，为了完整演示Tekton/Knative使用中的相关步骤，我们建议您使用不同的`Account`.  您将需要使用`您个人账户`中的IBM Cloud Resgitry 去存放Image文件， 并使用`IBM Acount`提供的Kubernetes集群, 完成全部的实验步骤， 为此，您需要获取`您个人账户`的IBM Cloud API Key ，用于在后面的步骤中访问您的IBM Cloud Resgitry。

在这一步，我们需要带领您获取`您的个人账户`中IBM Cloud API KEY.

1. 登录UI https://cloud.ibm.com/login 切换到**您自己的**ibm account(很重要!不要使用IBM ccount)。
![alt text](https://github.com/daisy-ycguo/devopslab/blob/master/images/login-personal-account.png)

2. 打开 https://cloud.ibm.com/iam/apikeys 页面， 点击“Create an IBM Cloud API key”按钮。

3. 输入一个名字，点击"Create"按钮。
![alt text](https://github.com/daisy-ycguo/devopslab/blob/master/images/create-api-key.png)

4. Download API key，打开apikey.json文件获取apikey。
![alt text](https://github.com/daisy-ycguo/devopslab/blob/master/images/download-apikey.png)

5. 点击上图中的Copy按钮，记录一下APIKEY


## 第三步：准备CloudShell & 配置您账户下的IBM Cloud Container Registry

一，访问[CloudShell](https://workshop.shell.cloud.ibm.com)，点击左上角的Login按钮，用IBM Cloud 用户名登陆。

![](https://github.com/daisy-ycguo/knativelab/raw/master/images/cloudshell-overview.png)

二，登陆后，出现一个页面，要求输入CloudShell的访问密码。咨询IBM工作人员获取访问密码。输入密码后，就进入CloudShell页面。

![](https://github.com/daisy-ycguo/knativelab/raw/master/images/cloudshell-passw.png)

三，在CloudShell页面中，点击右上角您的用户名，会弹出一个下拉框，选择`您的个人账户`。 

四，点击上图中右上角账户信息左侧的*命令行窗口图标*，页面会开始刷新。首次使用需要等待1-5分钟（等待时间与网速有关），一个云上的命令行窗口就创建好了。

五，在Cloud Shell 页面中，设置您的IBM Cloud Registry Region 信息

```
>> ibmcloud cr region-set us-south
The region is set to 'us-south', the registry is 'us.icr.io'.
OK
```

六，列出您的`namespace`   
`ibmcloud cr namespaces`   

七, 如果您还没有一个namespace,创建一个。   
`ibmcloud cr namespace-add <your_namespace>`

请记住您的NAMESPACE 名称。 


## 第四步：准备CloudShell &  连接到IBM Account下的Kubernetes集群


一，在CloudShell页面中，点击右上角切换账户，在弹出的下拉框里面，选择IBM。 

![](https://github.com/daisy-ycguo/knativelab/raw/master/images/cloudshell-account.png)


二，账户切换后， 原先的cloud shell terminal 窗口会关闭， 请重新点击*命令行窗口图标*， 创建新的命令行窗口。 


三，您领取到的集群名称大约为`kubecon19-knative**`，其中`**`部分为您的集群编号，如`kubecon19-knative66`。把这个集群名称记录在环境变量中。

   在CloudShell页面，输入：
   ```text
   export MYCLUSTER=<your_cluster_name>
   ```
四， 此外，我们需要把上一步记录的IBM Cloud API KEY,  Cloud Container Registry Region, Namespace信息，也记录在环境变量中，以备后续使用。 
   ```text
   export APIKEY=<your_api_key> 
   export REGISTRY=us.icr.io
   export NAMESPACE=<your_namespace>
   export EMAIL=<your_email_address>
   ```

五，获取你的集群的更多信息：

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
Master URL:                     https://c2.au-syd.containers.cloud.ibm.com:30425
Public Service Endpoint URL:    https://c2.au-syd.containers.cloud.ibm.com:30425
Private Service Endpoint URL:   -
Master Location:                Sydney
Master Status:                  Ready (1 hour ago)
Master State:                   deployed
Master Health:                  normal
Ingress Subdomain:              knative-guoyc.au-syd.containers.appdomain.cloud
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

继续 [Exercise 2 安装Istio和Knative](./02-istio-knative-install.md).

