# Github的准备

本地实验将使用github作为代码库。

## 申请Github账号

如果你已经拥有github.com的账号，可忽略此文档。如果你还没有申请github.com的账号，可按如下方法申请。

1. 访问 https://github.com
2. 点击 右上角的Sign up，进入以下界面。
![image](https://github.com/daisy-ycguo/devopslab/blob/master/images/github1.png)


点击 Next: select a plan
![image](https://github.com/daisy-ycguo/devopslab/blob/master/images/github2.png)

可以点击 Choose Free。
你的github.com账号就创建完成了。

## Clone代码库

1. 在浏览器中打开https://github.com/daisy-ycguo/devopslab 点击Fork。
2. Fork成功后会跳转到您自己的github账户下的devopslab repo。

## 设置环境变量

通过浏览器，打开您github代码库下面的文件：[src/setenv.sh](../src/setenv.sh)，直接点击编辑按钮，把您的github的账号、邮箱填写，然后commit。

```
export GITACCOUNT=<my_account>

export MYCLUSTER=
export KUBECONFIG=

export REGISTRY=
export NAMESPACE=
export EMAIL=<my_email>
```


接下来几步将获取其余的环境变量，请保留这个文件的编辑窗口。

继续 [准备您的镜像库和获取IBM Cloud API Key](./01-cloud-api.md).
