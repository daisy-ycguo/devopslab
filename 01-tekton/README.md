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
以上我们定义了可以重用的Pipeline和Task，下面我们创建一个PipelineRun来为它指定input resource和parameters，并执行这个pipeline。     
PipelineRun文件：[tekton/run/picalc-pipeline-run.yaml](https://github.com/IBM/tekton-tutorial/blob/master/tekton/run/picalc-pipeline-run.yaml)    
我们需要修改PipelineRun文件，替换`<REGISTRY>/<NAMESPACE>`为具体的值。   
使用**您自己的**ibm account登录   
`ibmcloud login --apikey <YOURAPIKEY>`   
登录您的私人container registry   
`ibmcloud cr login`   
列出您的`namespace`   
`ibmcloud cr namespaces`   
如果您还没有一个namespace,创建一个。   
`ibmcloud cr namespace-add <yourspacename>`
执行以下命令获得registry，在以下例子中registry为us.icr.io。      
      ```
      $ ibmcloud cr region
      You are targeting region 'us-south', the registry is 'us.icr.io'.
      ```   
      将文件中的`<REGISTRY>`和`<NAMESPACE>`用以上的值代替。
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
      PipelineRun没有一个固定的名字，每次执行的的时候会使用generateName的内容生成一个名字，例如‘picalc-pr-4jrtd’。这样做的好处是可以多次执行PipelineRun。   
PipelineRun要执行的Pipeline由pipelineRef指定。   
Pipeline暴露出来的parameters被指定了具体的值。   
关于Pipeline需要的resources，我们之后会定义一个名为picalc-git的PipelineResources。   
名为pipeline-account的service account用来提供pipeline执行时所需要的认证信息。我们后面将会创建这个service account。
下面我们来创建Tekton PipelineResource。名为picalc-git的PipelineResource指向一个git source。这git source是一个计算圆周率的go程序。它包含了一个Dockerfile来测试，编译代码，build image。[tekton/resources/picalc-git.yaml]（https://github.com/IBM/tekton-tutorial/blob/master/tekton/resources/picalc-git.yaml)     
下面我们创建这个Pipelineresource。   
`kubectl apply -f tekton/resources/picalc-git.yaml`   
下面我们来创建service account。Service account让pipeline可以访问被保护的资源-您私人的container registry。在创建service account之前，我们先要创建一个secret,它含了对您的container registry进行操作所需要的认证信息。   
`kubectl create secret docker-registry ibm-cr-push-secret --docker-server=<REGISTRY> --docker-username=iamapikey --docker-password=<YOURAPIKEY> --docker-email=me@here.com`   
其中`<YOURAPIKEY>`和`<REGISTRY>`的值，请参考实验步骤5。      
现在可以创建service account了。Service account的文件在这里[tekton/pipeline-account.yaml](https://github.com/IBM/tekton-tutorial/blob/master/tekton/pipeline-account.yaml)。   
`kubectl apply -f tekton/pipeline-account.yaml`   
这个yaml创建了一下资源：   
一个名为pipeline-account的ServiceAccount。在之前PipelineRun的定义中我们引用了这个serviceAccount。这个serviceAccount引用了我们之前创建的名为ibm-cr-push-secret的secret。这样就让pipeline获得了向你私人的container registry push image的认证。   
一个名为kube-api-secret的Secret,包含了用来访问Kubernetes API的认证信息信息，使得pipeline可以适用kubectl去操作您的kube cluster。   
一个名为pipeline-role的Role和一个名为pipeline-role-binding的RoleBinding，提供给pipeline基于resource的访问控制权限来创建和修改Knative services。   

6. 执行Pipeline  
现在万事俱备，我们来执行这个pipeline。        
`kubectl create -f tekton/run/picalc-pipeline-run.yaml`   
就像前面说过的，PipelineRun没有一个固定的名字，每次执行的的时候会使用generateName的内容生成一个名字。kubectl会返回一个新生成的PipelineRun resource名字。   
`pipelinerun.tekton.dev/picalc-pr-rqzgp created`   
可以用一下命令检查pipeline的状态。   
`kubectl describe pipelinerun picalc-pr-rqzgp`   
多检查几次直到你看到类似下面的状态。   
      ```
      Status:
        Completion Time:  2019-11-28T09:21:09Z
        Conditions:
          Last Transition Time:  2019-11-28T09:21:09Z
          Message:               All Steps have completed executing
          Reason:                Succeeded
          Status:                True
          Type:                  Succeeded
        Pod Name:                picalc-pr-rqzgp-source-to-image-dng6x-pod-97cd19
        Start Time:              2019-11-28T09:20:11Z
      ```   
      如果看到以上结果，我们就可以查看部署好的Knative service了。READY状态应该为True。    
      ```
      $ kubectl get ksvc picalc
      NAME     URL                                                             LATESTCREATED   LATESTREADY    READY   REASON
      picalc   http://picalc-default.cdl-performance-3c3cell.us-south.containers.appdomain.cloud   picalc-zgqkq    picalc-zgqkq   True
      ```   
      如果Pipeline没有执行成功，状态可能是这样：   
      ```
      Status:
        Conditions:
          Last Transition Time:  2019-04-15T14:30:46Z
          Message:               TaskRun picalc-pr-db6p6-deploy-to-cluster-7h8pm has failed
          Reason:                Failed
          Status:                False
          Type:                  Succeeded
        Start Time:              2019-04-15T14:29:23Z
      ```   
      在task run的状态下面，会有一条信息告诉您如何去查看失败的task的log。请根据log提示查看问题。      
你也可以通过下面的命令查看taskrun的状态，和失败的task的描述信息。   
`kubectl get taskruns`   
`kubectl describe taskrun <failed-task-run-name>`   

7. （可选）访问service   
Patch default service account,添加"imagePullSecrets"为我们前面创建的"ibm-cr-push-secret" 
      ```
      $ kubectl patch sa default -p '"imagePullSecrets": [{"name": "ibm-cr-push-secret" }]'
      serviceaccount/default patched
      ```   
      获得 istio ingressgateway的ip。    
`kubectl get svc istio-ingressgateway --namespace istio-system --output jsonpath="{.status.loadBalancer.ingress[*].ip}"`   
获得picalc service的domain URL。    
`kubectl get route picalc --output jsonpath="{.status.url}"| awk -F/ '{print $3}'`   
curl service。    
`curl -H "Host: <service-domain-url>" http://<istio-ingressgateway-ip>?iterations=20000000`   
你将得到返回结果：        
`3.1415926036`   
如果curl命令没有返回正确的结果，添加-vvv获得详细的信息。      
`curl -H "Host: <service-domain-url>" http://<istio-ingressgateway-ip>?iterations=20000000 -vvv`   
