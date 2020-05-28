package setting

import "path"

const (
	EImagesPath                      = "/root/images"                    // 镜像存储路径
	EContainerPath                   = "/root/container"                 // 容器存储路径
	EContainerMetaDataPath           = "/root/container_meta"            // 容器元信息
	EContainerLogsDataPath           = "/root/container_log"             // 日志路径
	EContainerNetworkDataPath        = "/root/container_network"         // 网络路径
	EContainerNetworkManagerDataPath = "/root/container_network_manager" // 网络分配信息

	EContainerLogName     = "run.log"
	EContainerNetworkFile = "network.json"
)

var (
	EContainerNetworkManagerDataFileName = path.Join(EContainerNetworkManagerDataPath, EContainerNetworkFile)
)
