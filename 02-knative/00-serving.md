# 金丝雀升级 Canary Upgrade 

## 1.检查service: 
```
$ kn service list
$ kn revision list
$ kn service update picalc –tag picalc-4rnfw=current

## 2.把所有请求都路由到service的当前版本
```
$ kn service update picalc –traffic current=100
$ kn service describe picalc
$ curl http://picalc-default.testcluster.us-south.containers.appdomain.cloud
```

## 3.构建一个新的green版本
```
$ kn service update –image us.icr.io/emilyregistry/picalc:green1120
$ kn service update picalc –tag picalc-xxx=candi
```
把50%的请求路由到当前版本，50%的请求路由到green版本：
```
$ kn service update picalc –traffic current=50 –traffic candi=50
```

## 4.验证金丝雀升级
```
$ While true; do curl http://picalc-default.mycluster6.us-south.containers.appdomain.cloud; done
```
```
$ kn route list
$ kn route describe picalc
```

## 5.升级后，把100%的请求路由到green版本
```
$ kn service update picalc –traffic candi=100
```

## 6.验证金丝雀升级
```
$ While true; do curl http://picalc-default.mycluster6.us-south.containers.appdomain.cloud; done
```
