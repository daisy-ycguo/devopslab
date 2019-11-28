# Tekton Hands-on Lab - from source code to production 

## 实验目标
- 了解Tekton pipelline的基本概念
- 创建一个pipeline来build和部署一个Knatvie的应用
- 执行一个pipeline并查看状态，学习问题处理
Tekton is an open source project to configure and run CI/CD pipelines within a Kubernetes cluster.

图

## 基本概念

- PipelineResource defines an object that is an input (such as a git repository) or an output (such as a docker image) of the pipeline.
- PipelineRun defines an execution of a pipeline. It references the Pipeline to run and the PipelineResources to use as inputs and outputs.
- Pipeline defines the set of Tasks that compose a pipeline.
- Task defines a set of build steps such as compiling code, running tests, and building and deploying images.

## 实验准备   
1. 登录Online command Line tool   
用浏览器打开https://workshop.shell.cloud.ibm.com/    
passcode: ikslab

2. 登录Kubecluster    
~~apikey怎么拿到~~   
`ibmcloud login --apikey xxxxxx`   

3. 确保您已完成[准备 Kubecluster 和安装 knative](https://github.com/QiuJieLi/devopslab/tree/master/00-install)的步骤   
验证knative和isto的pods都是Running状态      
*示例:*         
```
$ kubectl get pods --namespace istio-system
NAME                                      READY   STATUS    RESTARTS   AGE
cluster-local-gateway-5d8ccd46db-kt4vk    1/1     Running   0          6h14m
istio-citadel-654897999b-pbxjs            1/1     Running   0          6h15m
istio-egressgateway-77cfcd4f8d-8cbw2      1/1     Running   0          6h15m
istio-egressgateway-77cfcd4f8d-xgdfr      1/1     Running   0          6h14m
istio-galley-67987bf6cd-xwhlj             1/1     Running   0          6h15m
istio-ingressgateway-55b8654b85-8k9lh     1/1     Running   0          6h14m
istio-ingressgateway-55b8654b85-gph7k     1/1     Running   0          6h15m
istio-pilot-796cfc6987-b7qmg              2/2     Running   0          6h15m
istio-policy-68f46ddd67-sn7x2             2/2     Running   4          6h15m
istio-sidecar-injector-75479d8b85-vnwk8   1/1     Running   0          6h15m
istio-telemetry-b8dbc5985-wmj48           2/2     Running   4          6h15m
prometheus-7b87f6d744-rpgnq               1/1     Running   0          6h15m

$ kubectl get pods --namespace knative-serving
NAME                                READY   STATUS    RESTARTS   AGE
activator-7654759547-vc45g          2/2     Running   2          6h13m
autoscaler-74878dccf9-7b8gc         2/2     Running   2          6h13m
autoscaler-hpa-6fc598cdb-njhnf      1/1     Running   0          6h13m
controller-dc64bc644-s5f24          1/1     Running   0          6h13m
networking-istio-65f5b87479-nxftw   1/1     Running   0          6h13m
webhook-76c4d8d998-5sg4p            1/1     Running   0          5h12m
```

4. 安装 Tekton pipeline   
`kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml`     
使用以下命令查看Tekton Pipelines components 直到 STATUS 都显示 `Running`:     
*示例:*    
```
$ kubectl get pods --namespace tekton-pipelines
NAME                                           READY   STATUS    RESTARTS   AGE
tekton-pipelines-controller-7769bc5b76-tzsq9   1/1     Running   0          5h49m
tekton-pipelines-webhook-7849d4f75f-vd49j      1/1     Running   0          4h47m
```

5. 安装 Kn Client   
安装步骤参考https://github.com/knative/client/blob/master/docs/README.md   
例如，在linux上    
```
$ wget https://storage.googleapis.com/knative-nightly/client/latest/kn-linux-amd64
$ chmod +x kn-linux-amd64
$ export PATH=$PATH:$PWD
```   

验证kn安装成功   
```
$ kn
Manage your Knative building blocks:
...
```

6. 安装container-registry CLI plug-in
~~安装ibmcloud cli~~
`ibmcloud plugin install container-registry`

7. 创建您自己的container registry和namespace   
使用**您自己的**ibm account登录   
`$ ibmcloud login`   
登录container registry   
`$ ibmcloud cr login`   
列出您的`namespace`   
`$ ibmcloud cr namespaces`   
如果您还没有一个`namespace`,创建一个   
```
$ ibmcloud cr namespace-add tektonlab
Adding namespace 'yournamespace'...

Successfully added namespace 'yournamespace'

OK
```
执行以下命令获得`registry`，在以下例子中`registry`为us.icr.io   
```
$ ibmcloud cr region
You are targeting region 'us-south', the registry is 'us.icr.io'.
```

## 实验步骤
1. Clone tekton-tutorial项目到本地目录。   
`git clone https://github.com/IBM/tekton-tutorial`

2. 创建一个Task来build一个image并push到container registry。    
这个Task的文件在[tekton/tasks/source-to-image.yaml](https://github.com/IBM/tekton-tutorial/blob/master/tekton/tasks/source-to-image.yaml)。这个Taskbuild一个docker image并把它push到一个registry。   
一个Task可以包含一个或多个`Steps`。每个step定义了一个image用来执行这个step. 这个Task的步骤中使用了[kaniko](https://github.com/GoogleContainerTools/kaniko)项目来build source为一个docker image并把它push到一个registry。      
这个Task需要一个git类型的input resource,来定义souce的位置。这个git souce将被clone到本地的/workspace/git-source目录下。在Task中这个resource只是一个引用。后面我们将创建一个PipelineResources来定义真正的resouce资源。Task还使用了input parameters。这样做的好处是可以重用Task。   
后面我们会看到task是如何获得认证来puhs image到repository的。   
下面创建这个Task。   
`kubectl apply -f tekton/tasks/source-to-image.yaml`

3. 创建另一个Task来将image部署到Kubernetes cluster。   
这个Task的文件在[tekton/tasks/deploy-using-kubectl.yaml](https://github.com/IBM/tekton-tutorial/blob/master/tekton/tasks/deploy-using-kubectl.yaml)    
这个Task有两个步骤。    
第一，在container里通过执行sed命令更新yaml文件来部署第1步时通过source-to-image Task创建出来image。   
第二，使用Lachlan Evenson的k8s-kubectl container image执行kubectl命令来apply上一步的yaml文件。   
后面我们会看到这个task是如何获得认证来apply这个yaml文件中的resouce的。   
下面创建这个Task。   
`kubectl apply -f tekton/tasks/deploy-using-kubectl.yaml`

4. 创建一个Pipeline来组合以上两个Task。   
这个Pipeline文件在[tekton/pipeline/build-and-deploy-pipeline.yaml](https://github.com/IBM/tekton-tutorial/blob/master/tekton/pipeline/build-and-deploy-pipeline.yaml)    
Pipeline列出了需要执行的task，以及input output resources。所有的resources都必须定义为inputs或outputs。Pipeline 无法绑定一个PipelineResource。      
Pipeline还定义了每个task需要的input parameters。Task的input可以以多种方式进行定义，通过pipeline里的input parameter定义，或者直接设置，也可以使用task中的default值。在这个pipeline里，source-to-image task中的pathToContext parameter被暴露成为一个parameter pathToContext，而source-to-image task中pathToDockerFile则使用task中的default值。      
Task之间的顺序用runAfter关键字来定义。在这个例子中，deploy-using-kubectl task需要在source-to-image task之后执行。    

下面创建这个Pipeline。    
`kubectl apply -f tekton/pipeline/build-and-deploy-pipeline.yaml`

5. 创建PipelineRun和PipelineResources   
以上我们定义了可以重用的Pipeline和Task，下面我们来看看如何为它指定input resource和parameters并执行这个pipeline。     
下面是一个PipelineRun用来执行我们上面创建的Pipeline[tekton/run/picalc-pipeline-run.yaml](https://github.com/IBM/tekton-tutorial/blob/master/tekton/run/picalc-pipeline-run.yaml)      
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

    The PipelineRun does not have a fixed name. It uses generateName to generate a name each time it is created. This is because a particular PipelineRun resource executes the pipeline only once. If you want to run the pipeline again, you cannot modify an existing PipelineRun resource to request it to re-run -- you must create a new PipelineRun resource. While you could use name to assign a unique name to your PipelineRun each time you create one, it is much easier to use generateName.

    The Pipeline resource is identified under the pipelineRef key.

    The git resource required by the pipeline is bound to specific PipelineResources named picalc-git. We will define it in a moment.

    Parameters exposed by the pipeline are set to specific values.

    A service account named pipeline-account is specified to provide the credentials needed for the pipeline to run successfully. We will define this service account in the next part of the tutorial.
    
You must edit this file to substitute the values of <REGISTRY> and <NAMESPACE> with the information for your private container registry.

    To find the value for <REGISTRY>, enter the command ibmcloud cr region.
    To find the value of <NAMESPACE>, enter the command ibmcloud cr namespace-list.

修改Pipelinerun，指向正确的`<REGISTRY>/<NAMESPACE>`    
- 修改tekton/run/picalc-pipeline-run.yaml    
将文件中的`<REGISTRY>`和`<NAMESPACE>`用您个人account下的private container registry信息替代   
`<REGISTRY>`和`<NAMESPACE>`为实验准备->7 中获得的`registry`和`namespace`   

这是一个Tekton PipelineResource，它定义了picalc-git，指向一个git source。这是一个计算圆周率的go程序。包含了一个Dockerfile来测试，编译代码，build image。[tekton/resources/picalc-git.yaml]（https://github.com/IBM/tekton-tutorial/blob/master/tekton/resources/picalc-git.yaml）

创建pipelineresource。   
`kubectl apply -f tekton/resources/picalc-git.yaml`   

先不要创建PipelineRun，我们接下来还要为它定义service account。   

```
*输出示例:*
`secret/ibm-cr-push-secret annotated`

6. 创建service account
The last step before running the pipeline is to set up a service account so that it can access protected resources. The service account ties together a couple of secrets containing credentials for authentication along with RBAC-related resources for permission to create and modify certain Kubernetes resources.

First you need to enable programmatic access to your private container registry by creating either a registry token or an IBM Cloud Identity and Access Management (IAM) API key. The process for creating a token or an API key is described here.

After you have the token or API key, you can create the following secret.

kubectl create secret generic ibm-cr-push-secret --type="kubernetes.io/basic-auth" --from-literal=username=<USER> --from-literal=password=<TOKEN/APIKEY>
kubectl annotate secret ibm-cr-push-secret tekton.dev/docker-0=<REGISTRY>

替换`<APIKEY>`为实验准备->2 中使用的apikey，执行以下命令创建一个secret
```
kubectl create secret generic ibm-cr-push-secret --type="kubernetes.io/basic-auth" --from-literal=username=iamapikey --from-literal=password=<APIKEY>
```
*输出示例:*
`kubectl annotate secret ibm-cr-push-secret tekton.dev/docker-0=secret/ibm-cr-push-secret created`

替换`<REGISTRY>`为实验准备->7 中获得的`registry`
```
kubectl annotate secret ibm-cr-push-secret tekton.dev/docker-0=<REGISTRY>


where

    <USER> is either token if you are using a token or iamapikey if you are using an API key
    <TOKEN/APIKEY> is either the token or API key that you created
    <REGISTRY> is the URL of your container registry, such as us.icr.io or registry.ng.bluemix.net

Now you can create the service account using the following yaml. You can find this yaml file at tekton/pipeline-account.yaml.

This yaml creates the following Kubernetes resources:

    A ServiceAccount named pipeline-account. This is the name that the PipelineRun seen earlier uses to reference this account. The service account references the ibm-cr-push-secret secret so that the pipeline can authenticate to your private container registry when it pushes a container image.

    A Secret named kube-api-secret which contains an API credential (generated by Kubernetes) for accessing the Kubernetes API. This allows the pipeline to use kubectl to talk to your cluster.

    A Role named pipeline-role and a RoleBinding named pipeline-role-binding which provide the resource-based access control permissions needed for this pipeline to create and modify Knative services.

Apply the file to your cluster to create the service account and related resources.

kubectl apply -f tekton/pipeline-account.yaml


`kubectl apply -f tekton/pipeline-account.yaml`

7. 执行Pipeline  
现在万事俱备，我们来执行这个pipeline。       
`kubectl create -f tekton/run/picalc-pipeline-run.yaml`
