package cni

import (
	"github.com/sirupsen/logrus"
	"os"
)

var defaultNetworkDriverManager NetworkDriverManager

// 网络驱动管理器
type NetworkDriverManager map[string]INetworkDriver

// 获取一个网络驱动（因为是命令行模式，所以获取不到退出即可）
func (nd NetworkDriverManager) Get(name string) INetworkDriver {
	d, ok := nd[name]
	if !ok {
		logrus.Errorf("Not found driver: %s", name)
		os.Exit(1)
	}
	return d
}

// 删除一个网络驱动

func (nd NetworkDriverManager) Delete(name string) {
	delete(nd, name)
}

// 重置网络驱动
func (nd NetworkDriverManager) Clear() {
	nd = map[string]INetworkDriver{}
}

// 添加一个网络驱动
func (nd NetworkDriverManager) Add(name string, driver INetworkDriver) {
	nd[name] = driver
}

func GetNetworkDriverManager() NetworkDriverManager {
	if defaultNetworkDriverManager == nil {
		defaultNetworkDriverManager = map[string]INetworkDriver{}
	}
	return defaultNetworkDriverManager
}

func init() {
	GetNetworkDriverManager()
	defaultNetworkDriverManager.Add("bridge", &BridgeNetworkDriver{})
}
