---
author: saltbo
categories:
    - Linux
cover: /images/posts/learn-iptables/www.notion.so_solid_beige.png
createat: "2022-01-25T06:46:00+07:00"
date: "2022-01-25T00:00:00+07:00"
lastupdated: "2022-01-27T16:20:00+07:00"
name: learn-iptables
status: "Published \U0001F5A8"
tags:
    - iptables
title: iptables基本概念和操作
---

# 简介
iptables本质上只是一个命令行工具，真正起作用的是Linux内核中的netfilter，他有通过hook的方式进行相关的控制，如下图：
![](/images/posts/learn-iptables/s3.us-west-2.amazonaws.com_89d0f15f-c24c-40b2-97c7-d46f9c0f8d95.png)
# 基础知识
## 四表
1. filter表：负责过滤功能，防火墙；内核模块：iptables_filter
2. nat表：network address translation，网络地址转换功能；内核模块：iptable_nat
3. mangle表：拆解报文，做出修改，并重新封装 的功能；内核模块：iptable_mangle
4. raw表：关闭nat表上启用的连接追踪机制；内核模块：iptable_raw
## 五链
![](/images/posts/learn-iptables/s3.us-west-2.amazonaws.com_e5432620-abf4-4b3a-bcb2-aa199b615c55.png)
1. PREROUTING：在对数据包做路由选择之前，将应用此链中的规则
2. POSTROUTING：在对数据包做路由选择之后，将应用此链中的规则
3. INPUT：当收到访问本机地址的数据包时，将应用此链中的规则
4. OUTPUT：当本机向外发送数据包时，将应用此链中的规则
5. FORWARD：当收到需要通过防火墙转发给其他地址的数据包时，将应用此链中的规则

## 表链关系
|  | raw | mangle | nat | filter |
| :-----: | :-----: | :-----: | :-----: | :-----: |
| PREROUTING | ✅ | ✅ | ✅ | ❌ |
| POSTROUTING | ❌ | ✅ | ✅ | ❌ |
| INPUT | ❌ | ✅ | ❌ | ✅ |
| OUTPUT | ✅ | ✅ | ✅ | ✅ |
| FORWARD | ❌ | ✅ | ❌ | ✅ |
## 规则
- 匹配条件
    - 基本匹配条件
        - source/destination：IP地址
    - 扩展匹配条件
        - protocol：指定协议后-h可以看到更多支持的条件，eg: `iptables -p tcp -h`
        - sport/dport：端口，指定协议后可用
- 控制类型
    - ACCEPT：允许数据包通过
    - DROP：直接丢弃数据包，不给任何回应信息
    - REJECT：拒绝数据包通过，必要时会给数据发送端一个响应的信息
    - SNAT：源地址转换，解决内网用户用同一个公网地址上网的问题
    - MASQUERADE：是SNAT的一种特殊形式，适用于动态的、临时会变的ip上
    - DNAT：目标地址转换
    - REDIRECT：在本机做端口映射
    - LOG：在/var/log/messages文件中记录日志信息，然后将数据包传递给下一条规则
# 实操
### 创建一条拦截规则，阻止所有来自本机的访问
```bash
root@lima-default:~# iptables -I INPUT -s 127.0.0.1 -j REJECT
root@lima-default:~# iptables -vnL INPUT
Chain INPUT (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination
    0     0 REJECT     all  --  *      *       127.0.0.1            0.0.0.0/0            reject-with icmp-port-unreachable
```
### 创建一条拦截规则，阻止访问本机8000端口
```bash
root@lima-default:~# iptables -I INPUT -p tcp --dport 8000 -j REJECT
root@lima-default:~# iptables -vnL INPUT
Chain INPUT (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination
    0     0 REJECT     tcp  --  *      *       0.0.0.0/0            0.0.0.0/0            tcp dpt:8000 reject-with icmp-port-unreachable
```
### 创建一条路由规则，将80端口的流量转向8000端口
```bash
root@lima-default:~# iptables -t nat -I PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8000
root@lima-default:~# iptables -t nat -vnL PREROUTING
Chain PREROUTING (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination
    0     0 REDIRECT   tcp  --  *      *       0.0.0.0/0            0.0.0.0/0            tcp dpt:80 redir ports 8000
```
这是你会发现，你在本机访问80端口仍然不通，而在本机之外却通了。这是因为本机访问并没有走PREROUTING链。如果想要在本机也通，那么我们应该配置OUTPUT。
```bash
root@lima-default:~# iptables -t nat -I OUTPUT -d localhost -p tcp --dport 80 -j REDIRECT --to-ports 8000
root@lima-default:~# iptables -t nat -vnL OUTPUT
Chain OUTPUT (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination
    1    60 REDIRECT   tcp  --  *      *       0.0.0.0/0            127.0.0.1            tcp dpt:80 redir ports 8000
```
这里需要注意指定destination，否则你所有对外的请求都被转到8000端口了。

# 参考资料
- [http://www.yunweipai.com/35053.html](http://www.yunweipai.com/35053.html)
- [https://tinychen.com/20200414-iptables-principle-introduction](https://tinychen.com/20200414-iptables-principle-introduction/)

