
### proc 目录下重要的部分

- /proc/N　          PID为N的进程信息
- /proc/N/cmdline　  进程启动命令
- /proc/N/cwd　      链接到进程当前工作目录
- /proc/N/environ　  进程环境变量列表
- /proc/N/exe　      链接到进程的执行命令文件
- /proc/N/fd　       包含进程相关的所有文件描述符
- /proc/N/maps　     与进程相关的内存映射信息
- /proc/N/mem　      指代进程持有的内存，不可读
- /proc/N/root　     链接到进程的根目录
- /proc/N/stat　     进程的状态
- /proc/N/statm　    进程使用的内存状态
- /proc/N/status　   进程状态信息，比stat/statm更具可读性
- /proc/self/　      链接到当前正在运行的进程

### 相关知识

- pivot_root 是什么？
```text
pivot_root 是一个系统调用，主要功能是去改变当前的 root 文件系统。
pivot_root 可以将当前进程的 root 文件系统移动到 put_old 文件夹中，然后使new_root成为新的root 文件系统。
new_root 和 put_old 必须不能同时存在当前 root 的同一个文件系统中。
```

- pivot_root 和 chroot 区别是什么？
```text
pivot_root 和 chroot 的主要区别是，pivot_root 是把整个系统切换到一个新的 root 目录，而移除对之前 root 文件系统的依赖，这样你就能够 umount 原先的 root 文件系统。
而 chroot 是针对某个进程，系统的其他部分依旧运行于老的 root 目录中。
```

- aufs 是什么？挂载方式是什么？
```
https://segmentfault.com/a/1190000008489207
```

- runContainer 入参设计模式
```text
https://blog.kunlunjun.net/posts/golang-option-design-pattern/
```
