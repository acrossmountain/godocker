package container

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"
)

var (
	oldRootPath = ".pivot_root"
)

func setUpMount() {
	mountPrivate()

	pwd, err := os.Getwd() // 获取当前工作目录 pwd = print work dir ？
	if err != nil {
		logrus.Errorf("Get current dir error: %v", err)
		return
	}

	if err := pivotRoot(pwd); err != nil {
		logrus.Error(err)
		return
	}

	mountProc()
}

func mountPrivate() {
	// 新的 linux kernel, 默认挂载是 share
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		logrus.Error(err)
	}
}

func mountProc() {
	// MS_NOEXEC：在本文件系统中不允许运行其它程序
	// MS_NOSUID：在本系统运行程序的时候，不允许 set-user-id 或 set-group-id
	procFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(procFlags), ""); err != nil {
		logrus.Errorf("Mount proc error: %v", err)
	}

	tmpfsFlags := syscall.MS_NOSUID | syscall.MS_STRICTATIME
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", uintptr(tmpfsFlags), "mode=755"); err != nil {
		logrus.Errorf("Mount tmpfs error: %v", err)
	}
}

func pivotRoot(root string) error {

	// 使当前的 root 的老 root 和新 root 不在同一个文件系统下
	// 把 root 重新 mount 一次，bind mount 是把相同的内容换了一个挂载点的挂载方式。
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}

	// 创建 rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, oldRootPath)
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return fmt.Errorf("mkdir .pivot_root error: %v", err)
	}

	// pivot_root 到新的 rootfs，老的 old_root 现在挂载在 rootfs/.pivot_root 上
	// 挂载点目前依然可以通过 mount 命令查看到
	// logrus.Infof("root: %s, pivot_root: %s", root, pivotDir)
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root error: %v", err)
	}

	// 修改当前当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / error: %v", err)
	}

	pivotDir = filepath.Join("/", oldRootPath)
	// unmount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir error: %v", err)
	}

	return os.Remove(pivotDir) // 删除老的 root
}
