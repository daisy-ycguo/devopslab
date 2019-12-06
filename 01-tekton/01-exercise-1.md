# 创建Tekton pipeline并手动运行 

## 实验目标
- 了解Tekton pipeline的基本概念
- 创建一个pipeline来build和部署一个Knatvie的应用
- 执行一个pipeline并查看状态
- 问题诊断

## 基本概念

- PipelineResource 定义了输入 (例如一个 git repository) 或输出 (例如 docker image)供pipeline来使用.
- PipelineRun 定义了一个pipeline的执行. 它引用PipelineResources做为输入输出，引用Pipeline来执行，
- Pipeline 定义了一系列的Tasks.
- Task 定义了一系列的build steps例如编译代码, 执行测试, 构建和部署images。

![alt text](https://github.com/IBM/tekton-tutorial/blob/master/doc/source/images/crd.png)


## 前提
- [安装Istio和Knative](https://github.com/daisy-ycguo/devopslab/blob/master/00-install/istio-knative-install.md)   
- [Tekton安装](https://github.com/daisy-ycguo/devopslab/blob/master/00-install/tekton-install.md)  
- [准备Kubernetes集群环境](https://github.com/daisy-ycguo/devopslab/blob/master/00-install/01-k8s-connect.md)

## 实验步骤
### 第一步. clone代码到本地
1. 在您的github代码库页面，点击"Clone or download"，获取URL。
2. 将代码库clone到本地，在CloudShell窗口运行这个命令：    
`git clone https://github.com/<your-git-account>/devopslab.git`
3. 运行setenv.sh配置环境变量。在CloudShell中执行：
```
source devopslab/src/setenv.sh
```
4. 配置Cloud API Key。
打开您本地保存的aipkey.json，拷贝apikey，在CloudShell中执行：
```
export APIKEY=<your_api_key>
```
例如，您的apikey.json中有：
```
"apikey": "abcedef123434"
```
则执行
```
export APIKEY=abcedef123434
```

### 第二步. 创建一个Task， 用于build一个image并push到您的container registry。 

在CloudShell中请执行命令：
```
kubectl apply -f devopslab/src/tekton/basic/tekton/tasks/source-to-image.yaml
```

在这个步骤中， 我们创建了Task：devopslab/src/tekton/basic/tekton/tasks/source-to-image.yaml。

这个Task用于构建一个docker image并把它push到一个registry， 具体内容如下：   

```
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: source-to-image
spec:
  inputs:
    resources:
      - name: git-source
        type: git
    params:
      - name: pathToContext
        description: The path to the build context, used by Kaniko - within the workspace
        default: .
      - name: pathToDockerFile
        description: The path to the dockerfile to build (relative to the context)
        default: Dockerfile
      - name: imageUrl
        description: Url of image repository
      - name: imageTag
        description: Tag to apply to the built image
        default: "latest"
  steps:
    - name: build-and-push
      image: gcr.io/kaniko-project/executor
      command:
        - /kaniko/executor
      args:
        - --dockerfile=$(inputs.params.pathToDockerFile)
        - --destination=$(inputs.params.imageUrl}:${inputs.params.imageTag)
        - --context=/workspace/git-source/$(inputs.params.pathToContext)
```
说明：
- 一个Task可以包含一个或多个`Steps`。每个step定义了一个image用来执行这个step. 这个Task的步骤中使用了[kaniko](https://github.com/GoogleContainerTools/kaniko)项目来build source为一个docker image并把它push到一个registry。      
- 这个Task需要一个git类型的input resource,来定义souce的位置。这个git souce将被clone到本地的/workspace/git-source目录下。在Task中这个resource只是一个引用。后面我们将创建一个PipelineResources来定义真正的resouce资源。
- Task还使用了input parameters。这样做的好处是可以重用Task。   
- 后面我们会看到task是如何获得认证来puhs image到repository的。  


### 第三步. 创建另一个Task来将image部署到Kubernetes cluster。  

在CloudShell中请执行命令：
```
kubectl apply -f devopslab/src/tekton/basic/tekton/tasks/deploy-using-kubectl.yaml
```

在这个步骤中， 我们创建了Task：devopslab/src/tekton/basic/tekton/tasks/deploy-using-kubectl.yaml。  

这个Task有两个步骤。    
- 第一，在container里通过执行sed命令更新yaml文件来部署第1步时通过source-to-image Task创建出来image。   
- 第二，使用Lachlan Evenson的k8s-kubectl container image执行kubectl命令来apply上一步的yaml文件。   

后面我们会看到这个task是如何获得认证来apply这个yaml文件中的resouce的。   
```
....
steps:
    - name: update-yaml
      image: alpine
      command: ["sed"]
      args:
        - "-i"
        - "-e"
        - "s;__IMAGE__;${inputs.params.imageUrl}:${inputs.params.imageTag};g"
        - "/workspace/git-source/${inputs.params.pathToYamlFile}"
    - name: run-kubectl
      image: lachlanevenson/k8s-kubectl
      command: ["kubectl"]
      args:
        - "apply"
        - "-f"
        - "/workspace/git-source/${inputs.params.pathToYamlFile}"
```


### 第四步. 创建一个Pipeline来组合以上两个Task。   

在CloudShell中请执行命令：
```  
kubectl apply -f devopslab/src/tekton/basic/tekton/pipeline/build-and-deploy-pipeline.yaml
```

在这个步骤中， 我们创建了Pipeline：devopslab/src/tekton/basic/tekton/pipeline/build-and-deploy-pipeline.yaml。

这个Pipeline具体内容是：

```
apiVersion: tekton.dev/v1alpha1
kind: Pipeline
metadata:
  name: build-and-deploy-pipeline
spec:
  resources:
    - name: git-source
      type: git
  params:
    - name: pathToContext
      description: The path to the build context, used by Kaniko - within the workspace
      default: src/app
    - name: pathToYamlFile
      description: The path to the yaml file to deploy within the git source
    - name: imageUrl
      description: Url of image repository
    - name: imageTag
      description: Tag to apply to the built image
  tasks:
  - name: source-to-image
    taskRef:
      name: source-to-image
    params:
      - name: pathToContext
        value: "${params.pathToContext}"
      - name: imageUrl
        value: "${params.imageUrl}"
      - name: imageTag
        value: "${params.imageTag}"
    resources:
      inputs:
        - name: git-source
          resource: git-source
  - name: deploy-to-cluster
    taskRef:
      name: deploy-using-kubectl
    runAfter:
      - source-to-image
    params:
      - name: pathToYamlFile
        value:  "${params.pathToYamlFile}"
      - name: imageUrl
        value: "${params.imageUrl}"
      - name: imageTag
        value: "${params.imageTag}"
    resources:
      inputs:
        - name: git-source
          resource: git-source
```    
说明：    
- Pipeline列出了需要执行的task，以及input output resources。    
- Pipeline还定义了每个task需要的input parameters。Task的input可以以多种方式进行定义，通过pipeline里的input parameter定义，或者直接设置，也可以使用task中的default值。在这个pipeline里，source-to-image task中的pathToContext parameter被暴露成为一个parameter 'pathToContext'，而source-to-image task中pathToDockerFile则使用task中的default值。      
- Task之间的顺序用runAfter关键字来定义。在这个例子中，deploy-using-kubectl task需要在source-to-image task之后执行。    


### 第五步. 创建PipelineRun和PipelineResources，并执行Pipeline  

以上Task, Pipeline定义均为统一的模板文件，下面我们创建一个PipelineRun来指定input resource和parameters，用于真正执行这个pipeline。 

#### 5.1 创建Tekton PipelineResource

Pipeline resouce 文件路径：

* *请修改文件* devopslab/src/tekton/basic/tekton/resources/hello-git.yaml, 输入您的git-account信息；
* 保存您的修改， 

在CloudShell中请执行命令：
```
kubectl apply -f devopslab/src/tekton/basic/tekton/resources/hello-git.yaml
```

说明： 

PipelineResource指向一个git source。这git source是一个hello world的go程序。它包含了一个Dockerfile来测试，编译代码，build image。   
您需要更新hello-git.yaml，将下面url的value改为您自己clone的git repo。

```
apiVersion: tekton.dev/v1alpha1
kind: PipelineResource
metadata:
  name: hello-git
spec:
  type: git
  params:
    - name: revision
      value: master
    - name: url
      value: https://github.com/<your-git-account>/devopslab
```

#### 5.2 创建service account
Service account让pipeline可以访问被保护的资源-您私人的IBM container registry。

* 在创建service account之前，我们先要创建一个secret,它含了对您的container registry进行操作所需要的认证信息。  

在CloudShell中请执行命令：
```
kubectl create secret docker-registry ibm-cr-push-secret --docker-server=us.icr.io --docker-username=iamapikey --docker-password=$APIKEY --docker-email=$EMAIL
``` 

其中$APIKEY, $EMAIL的值在环境变量中配置过。

* 创建service account。

在CloudShell中请执行命令：
```
kubectl apply -f devopslab/src/tekton/basic/tekton/pipeline-account.yaml
``` 

在这个步骤中，我们使用了devopslab/src/tekton/basic/tekton/pipeline-account.yaml创建service account. 
```
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pipeline-account
secrets:
- name: ibm-cr-push-secret
imagePullSecrets:
- name: ibm-cr-push-secret

---

apiVersion: v1
kind: Secret
metadata:
  name: kube-api-secret
  annotations:
    kubernetes.io/service-account.name: pipeline-account
type: kubernetes.io/service-account-token

---

kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: pipeline-role
rules:
- apiGroups: ["serving.knative.dev"]
  resources: ["services"]
  verbs: ["get", "create", "update", "patch"]

---

apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: pipeline-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: pipeline-role
subjects:
- kind: ServiceAccount
name: pipeline-account
```

说明：
这个yaml创建了以下资源：   
- 一个名为pipeline-account的ServiceAccount。在之前PipelineRun的定义中我们引用了这个serviceAccount。这个serviceAccount引用了我们之前创建的名为ibm-cr-push-secret的secret。这样就让pipeline获得了向你私人的container registry push image的认证。   
- 一个名为kube-api-secret的Secret,包含了用来访问Kubernetes API的认证信息信息，使得pipeline可以适用kubectl去操作您的kube cluster。   
- 一个名为pipeline-role的Role和一个名为pipeline-role-binding的RoleBinding，提供给pipeline基于resource的访问控制权限来创建和修改Knative services。  

#### 5.5 修改PipelineRun文件，替换`<NAMESPACE>`为具体的值。

PipelineRun文件路径：devopslab/src/tekton/basic/tekton/run/hello-pipeline-run.yaml。 

将文件中的`<NAMESPACE>`用前述的变量值替换。 
```
apiVersion: tekton.dev/v1alpha1
kind: PipelineRun
metadata:
  generateName: hello-pr-
spec:
  pipelineRef:
    name: build-and-deploy-pipeline
  resources:
    - name: git-source
      resourceRef:
        name: hello-git
  params:
    - name: pathToYamlFile
      value: "src/tekton/basic/knative/hello.yaml"
    - name: imageUrl
      value: us.icr.io/<NAMESPACE>/hello
    - name: imageTag
      value: "1.0"
  trigger:
    type: manual
  serviceAccount: pipeline-account
```
说明：
- PipelineRun没有一个固定的名字，每次执行的的时候会使用generateName的内容生成一个名字，例如‘hello-pr-4jrtd’。这样做的好处是可以多次执行PipelineRun。   
- PipelineRun要执行的Pipeline由pipelineRef指定。   
- Pipeline暴露出来的parameters被指定了具体的值。   
- Pipeline需要的resources是我们已经定义的hello-git的PipelineResources。   
- pipeline执行时所需要的认证信息是我们已经定义的pipeline-account的service account。    

#### 5.6 执行Pipeline  

在CloudShell中请执行命令：
```
kubectl create -f devopslab/src/tekton/basic/tekton/run/hello-pipeline-run.yaml
```   

#### 5.7  检查Pipeline Run的执行结果

PipelineRun创建后没有一个固定的名字，每次执行的的时候会使用generateName的内容生成一个名字。kubectl会返回一个新生成的PipelineRun resource名字。   
`pipelinerun.tekton.dev/hello-pr-ktc9j created`  

1. 检查pipelinerun的状态，确保pipelinerun处在Running状态(REASON栏)。
```
$ kubectl get pr
NAME             SUCCEEDED   REASON    STARTTIME   COMPLETIONTIME
hello-pr-llpm4   Unknown     Running   4s
```

1. 检查taskrun的状态,直到的状态都变为True(SUCCEEDED)。（大概需要1-2分钟）
```
$ kubectl get taskrun
NAME                                     SUCCEEDED   REASON      STARTTIME   COMPLETIONTIME
hello-pr-ktc9j-deploy-to-cluster-wvn94   True        Succeeded   11s         1s
hello-pr-ktc9j-source-to-image-ccvfg     True        Succeeded   67s         11s
```

2. 如果看到以上结果，我们就可以查看部署好的Knative service了。READY状态为True说明部署成功了。    
```
$ kubectl get ksvc hello
NAME    URL                                                                      LATESTCREATED   LATESTREADY   READY   REASON
hello   http://hello-default.<CLUSTER-NAME>.us-south.containers.appdomain.cloud   hello-thjf9     hello-thjf9   True
```   
请记录这里的Ingress-doman
```
export INGRESS=<CLUSTER-NAME>.us-south.containers.appdomain.cloud
```

3. 最后，访问应用。
```
curl http://hello-default.$INGRESS
Hello world, this is BLUE-IBM!!!
```

恭喜您，您已经完成了Tekton第一个实验。下面可以开始[实验2](./02-exercise-2.md)

### 参考资料

#### 问题诊断

如果以上步骤遇到问题，尝试下面的方法进行诊断：   
1. 检查task run的状态。   
```
kubectl get taskrun
```
2. 如果有taskrun失败了，检查该taskrun的describe之后的‘Message’。   
```
kubectl describe taskrun hello-pr-kpxbr-source-to-image-hbbw4
...
Message
"step-build-and-push" exited with code 1 (image: "gcr.io/kaniko-project/executor@sha256:9c40a04cf1bc9d886f7f000e0b7fa5300c31c89e2ad001e97eeeecdce9f07a29"); for logs run: kubectl -n default logs hello-pr-7gr8f-source-to-image-sn7r8-pod-02e5da -c step-build-and-push
```
3. 按照提示检查pod的log获取失败的详细信息。   
```
$ kubectl -n default logs hello-pr-fwdz7-source-to-image-6tvck-pod-9a97c6 -c step-build-and-push
error checking push permissions -- make sure you entered the correct tag name, and that you are authenticated correctly, and try again: checking push permission for "us.icr.io/sophy/hello:1.0": DENIED: You have exceeded your storage quota. Delete one or more images, or review your storage quota and pricing plan. For more information, see https://ibm.biz/BdjFwL; [map[Action:pull Class: Name:sophy/hello Type:repository] map[Action:push Class: Name:sophy/hello Type:repository]]
```

4. 检查pipeline的状态
```
kubectl get pipelinerun
```
最后的Events应该是‘Succeeded’状态
```
kubectl describe pipelinerun hello-pr-xxxx
...
  Normal   Succeeded          92s    pipeline-controller  All Tasks have completed executing
```

5. 检查service的状态
```
kubectl get ksvc

```
如果READY不为True,describe ksvc查看Message提供的错误信息。
```
$ kubectl describe ksvc hello
```

#### 清理环境

如果以上步骤无法找到问题并解决，可以尝试清掉以上实验所创建的资源，重新开始。
清理脚本：
```
kubectl delete pipeline --all
kubectl delete task --all
kubectl delete sa pipeline-account
kubectl delete PipelineResource hello-git
kubectl delete secret ibm-cr-push-secret
kubectl delete pipelinerun --all
kubectl delete ksvc hello
```

#### 备忘单

所有实验步骤脚本：(确保需要更新的文件已经更新好）
```
kubectl apply -f devopslab/src/tekton/basic/tekton/tasks/source-to-image.yaml
kubectl apply -f devopslab/src/tekton/basic/tekton/tasks/deploy-using-kubectl.yaml
kubectl apply -f devopslab/src/tekton/basic/tekton/pipeline/build-and-deploy-pipeline.yaml
kubectl create secret docker-registry ibm-cr-push-secret --docker-server=us.icr.io --docker-username=iamapikey --docker-password=<YOURAPIKEY> --docker-email=me@here.com
kubectl apply -f devopslab/src/tekton/basic/tekton/pipeline-account.yaml
kubectl patch sa pipeline-account -p '"imagePullSecrets": [{"name": "ibm-cr-push-secret" }]'
kubectl apply -f devopslab/src/tekton/basic/tekton/resources/hello-git.yaml
kubectl create -f devopslab/src/tekton/basic/tekton/run/hello-pipeline-run.yaml
```
