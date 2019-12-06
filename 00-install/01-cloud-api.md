# 准备您的镜像库和获取IBM Cloud API Key

本次实验的容器镜像将存储在IBM Cloud上的镜像库中，这里将准备镜像库并且获取IBM Cloud API Key。

## 前提

* 拥有一个IBM Cloud账号，也被称为IBM ID。如果没有注册，请到[http://cloud.ibm.com](http://cloud.ibm.com)上注册。
* 准备一个可以联网的浏览器，推荐Chrome，Firefox，和Safari。

## 第一步：准备CloudShell & 配置您账户下的IBM Cloud Container Registry

一，访问[CloudShell](https://workshop.shell.cloud.ibm.com)，点击左上角的Login按钮，用IBM Cloud 用户名登陆。

![](https://github.com/daisy-ycguo/knativelab/raw/master/images/cloudshell-overview.png)

二，登陆后，出现一个页面，要求输入CloudShell的访问密码。咨询IBM工作人员获取访问密码。输入密码后，就进入CloudShell页面。

![](https://github.com/daisy-ycguo/knativelab/raw/master/images/cloudshell-passw.png)

三，在CloudShell页面中，点击右上角您的用户名，会弹出一个下拉框，选择`您的个人账户`(非IBM账户)。 

四，点击上图中右上角账户信息左侧的*命令行窗口图标*，页面会开始刷新。首次使用需要等待1-5分钟（等待时间与网速有关），一个云上的命令行窗口就创建好了。

五，在Cloud Shell 页面中，set您的IBM Cloud Registry Region

```
ibmcloud cr region-set us-south
The region is set to 'us-south', the registry is 'us.icr.io'.
OK
```

六，列出您的`namespace`   
`ibmcloud cr namespaces`   
（如果列出很多namesapce,可能您使用了错误的账户，请务必使用`您的个人账户`(非IBM账户)，参考第三步）

七, 创建一个namespace，推荐使用您的CLUSTER名字，以免与其他人冲突。 （如果收到提示 *The requested namespace is already in use*，请修改 namespace 名称再重复操作）

`ibmcloud cr namespace-add <您的CLUSTER名字>`

八，继续修改环境变量，您应该已经在浏览器中打开了 github 代码库下面的文件：[src/setenv.sh](../src/setenv.sh)，将在上面步骤中创建的namespace 信息填入到`NAMESPACE=`后面，接下来将获取其余的环境变量，请继续保留这个页面。

```
export GITACCOUNT=<my_account>

export MYCLUSTER=
export KUBECONFIG=

export NAMESPACE=
export EMAIL=<my_email>
export INGRESS=
```

## 第二步：获取IBM Cloud API KEY。 

接下来需要获取`您个人账户`的IBM Cloud API Key ，用于在后面的步骤中访问您的IBM镜像库。

1. 打开页面 https://cloud.ibm.com/iam/apikeys ，如果没有登陆IBM Cloud，需要先登陆。切换到**您自己的**ibm account(很重要!不要使用IBM ccount)。
![alt text](https://github.com/daisy-ycguo/devopslab/blob/master/images/login-personal-account.png)

2. 点击“Create an IBM Cloud API key”按钮。

3. 输入一个名字，点击"Create"按钮。
![alt text](https://github.com/daisy-ycguo/devopslab/blob/master/images/create-api-key.png)

4. 点击Download API key，打开apikey.json文件获取apikey。
![alt text](https://github.com/daisy-ycguo/devopslab/blob/master/images/download-apikey.png)

apikey.json如下面所示。这个文件将保留在您的电脑上。您如果忘记apikey可以随时查看。
```
{
	"name": "test",
	"description": "test",
	"createdAt": "2019-12-06T03:14+0000",
	"apikey": "******1fBXOT****"
}
```

继续 [准备Kubernetes集群环境](./02-k8s-connect.md).



