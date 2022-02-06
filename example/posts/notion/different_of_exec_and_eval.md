---
author: saltbo
categories: []
createat: "2021-06-09T08:56:00+07:00"
date: "2022-01-09T00:00:00+07:00"
lastupdated: "2022-01-27T13:34:00+07:00"
name: different of exec and eval
status: "Published \U0001F5A8"
tags:
    - Shell
title: Shell中exec和eval的区别
---


默认情况下，如果直接执行bash -c command，command会以子进程方式运行，执行完成后返回父进程继续执行。
## exec
使用exec bash -c command，父进程的pid会转移给command，这时实际上父级shell已经退出，所以无法执行exec后面的脚本。
## eval
假设command中包含export之类的命令，如果采用bash -c的方式，export的变量是无法在父级shell中获取到的。这时采用eval就可以了。和exec相同的是：进程pid没有变。但它没有替换老的shell，而是在老的shell里执行新的命令。

## bash -l
login: 加载bashrc和profile等文件

