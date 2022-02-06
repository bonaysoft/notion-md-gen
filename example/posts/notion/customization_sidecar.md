---
author: saltbo
categories:
    - ServiceMesh
createat: "2022-01-07T15:19:00+07:00"
date: "2022-01-07T00:00:00+07:00"
lastupdated: "2022-01-27T13:34:00+07:00"
name: customization sidecar
status: "Published \U0001F5A8"
tags:
    - Istio
    - Sidecar
title: 如何自定义Sidecar及设置Sidecar的日志
---

### 场景
想要收集Envoy的Accesslog，希望把日志打到Node上，让Node上的采集程序能采集到

### 步骤
1. 挂载一个日志卷进去
2. 修改istio的配置文件

问题1：如何自定义sidecar的配置？
答案：直接在containers下面新建一个容器，name为istio-proxy，image为auto。然后根据我们的需求设置volumeMounts即可
```yaml
template:
  metadata:
    labels:
      app: httpbin
      version: v1
  spec:
    serviceAccountName: httpbin
    containers:
    - image: docker.io/kennethreitz/httpbin
      imagePullPolicy: IfNotPresent
      name: httpbin
      ports:
      - containerPort: 80
    - image: auto
      name: istio-proxy
      volumeMounts:
        - mountPath: /data/logs
          name: logs
    volumes:
      - name: logs
        emptyDir: {}
```

问题2：如何修改Envoy的Accesslog输出地址？
- 全局的方式可以直接改istio的meshConfig，修改accessLogFile即可
- 修改sidecar级别的需要利用Telemetry来配置
    - 第一步在meshConfig里增加一个provider
    - 第二步增加一个Telemetry使用这个provider

meshConfig: @`kubectl istio-system configmap istio`
```yaml
accessLogFile: /dev/stdout
defaultConfig:
  discoveryAddress: istiod.istio-system.svc:15012
  proxyMetadata: {}
  tracing:
    zipkin:
      address: zipkin.istio-system:9411
extensionProviders:
  - name: logfile
    envoyFileAccessLog:
      path: "/data/logs/istio.log"
enablePrometheusMerge: true
rootNamespace: istio-system
trustDomain: cluster.local
```
Telemetry如下：
```yaml
apiVersion: telemetry.istio.io/v1alpha1
kind: Telemetry
metadata:
  name: frontend-logging
spec:
  selector:
    matchLabels:
      app: httpbin
  accessLogging:
    - providers:
        - name: logfile
```
注：改完meshConfig需要重启sidecar才能生效

### 参考文档
- [https://istio.io/latest/docs/setup/additional-setup/sidecar-injection/#customizing-injection](https://istio.io/latest/docs/setup/additional-setup/sidecar-injection/#customizing-injection)
- [https://github.com/istio/istio.io/issues/7613#issuecomment-1009553832](https://github.com/istio/istio.io/issues/7613#issuecomment-1009553832)
- [https://istio.io/latest/docs/reference/config/istio.mesh.v1alpha1/#MeshConfig-ExtensionProvider-EnvoyFileAccessLogProvider](https://istio.io/latest/docs/reference/config/istio.mesh.v1alpha1/#MeshConfig-ExtensionProvider-EnvoyFileAccessLogProvider)


