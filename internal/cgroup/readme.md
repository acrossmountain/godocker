### 问题 
- write /sys/fs/cgroup/cpuset/xxx/tasks: no space left on device
```
在添加 tasks 之前，cpuset.cpus 和 cpuset.mems 需要提前配置。相关文章：https://blog.csdn.net/xftony/article/details/80536562
```
- 限制 cpu.shares 时，发现会占用所有的CPU
```
当前 slice 下只有当前进程比较繁忙，所以占用全部的CPU。 
```



### 相关文章

- https://icloudnative.io/posts/understanding-cgroups-part-1-basics/
- https://icloudnative.io/posts/understanding-cgroups-part-2-cpu/
- https://icloudnative.io/posts/understanding-cgroups-part-3-memory/
