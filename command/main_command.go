package command

import (
	"mydocker/cni"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"mydocker/container"
	"mydocker/images"
	"mydocker/subsystems"
)

var (
	// 运行
	runCommand = cli.Command{
		Name: "run",
		Usage: `Create a container with namespace and cgroups limit
								 mydocker run -ti [command]`,
		SkipFlagParsing: false,
		Flags: []cli.Flag{
			// 后台运行
			cli.BoolFlag{
				Name:  "d",
				Usage: "detach container",
			},
			// 前台等待运行
			cli.BoolFlag{
				Name:  "ti",
				Usage: "enable tty",
			},
			// 指定镜像名称
			cli.StringFlag{
				Name:  "image",
				Usage: "image name",
			},
			// 卷
			cli.StringFlag{
				Name:  "v",
				Usage: "volumes",
			},
			// 容器名称
			cli.StringFlag{
				Name:  "name",
				Usage: "container name",
			},

			cli.StringFlag{
				Name:  "m",
				Usage: "memory limit",
			},
			cli.StringFlag{
				Name:  "cpushare",
				Usage: "cpushare limit",
			},
			cli.StringFlag{
				Name:  "cpuset",
				Usage: "cpuset limit",
			},
			cli.StringSliceFlag{
				Name:  "e",
				Usage: "set environment",
			},
		},
		Action: func(ctx *cli.Context) error {
			// 启动容器

			ttl := ctx.Bool("ti")
			detach := ctx.Bool("d")
			if ttl && detach {
				return errors.New("ttl && detach")
			}

			if len(ctx.Args()) < 1 {
				return errors.New("Missing container command")
			}
			var arr []string
			for _, v := range ctx.Args() {
				arr = append(arr, v)
			}

			Run(
				&RunArgs{
					Tty:      ttl,
					CmdArray: arr,
					Image:    ctx.String("image"),
					Volume:   ctx.String("v"),
					Name:     ctx.String("name"),
					Env:      ctx.StringSlice("e"),
					ResourceConfig: &subsystems.ResourceConfig{
						MemoryLimit: ctx.String("m"),
						CpuSet:      ctx.String("cpuset"),
						CpuShare:    ctx.String("cpushare"),
					},
				})
			return nil
		},
	}

	// 初始化容器
	initCommand = cli.Command{
		Name:            "init",
		Usage:           "Init container process run user's rocess in container.Do not call it outside",
		SkipFlagParsing: false,
		Action: func(ctx *cli.Context) error {
			// 容器初始化
			log.Infof("init come on | %v", ctx.Args())

			return container.RunContainerInitProcess()
		},
	}

	// 删除
	rmCommand = cli.Command{
		Name:            "rm",
		Usage:           "rm container",
		SkipFlagParsing: false,
		Action: func(ctx *cli.Context) error {
			// 容器删除
			log.Infof("rm come on | %v", ctx.Args())

			args := ctx.Args()
			if len(args) != 1 {
				log.Errorf("Remove.Len.Args!=1")
				return errors.Errorf("Missing container name")
			}
			return container.RemoveContainer(args[0])
		},
	}

	// ps命令
	listCommand = cli.Command{
		Name:            "ps",
		Usage:           "list container",
		SkipFlagParsing: false,
		Action: func(ctx *cli.Context) error {
			log.Infof("ps come on | %v", ctx.Args())
			return container.ListContainer()
		},
	}

	// 侵入容器执行命令
	execCommand = cli.Command{
		Name:            "exec",
		Usage:           "exec",
		SkipFlagParsing: false,
		Action: func(ctx *cli.Context) error {
			log.Infof("exec come on | %v", ctx.Args())

			// 如果环境变量不为空就返回
			if os.Getenv(container.ENV_EXEC_PID) != "" {
				log.Errorf("pid callback pid | %d", os.Getgid())
				return nil
			}
			args := ctx.Args()
			if len(args) < 2 {
				log.Errorf("Exec.Len.Args!=2")
				return errors.Errorf("Missing container name or command")
			}
			name := args[0]

			return container.ExecContainer(name, args.Tail())
		},
	}
	// 查看日志
	logsCommand = cli.Command{
		Name:            "logs",
		Usage:           "show container log",
		SkipFlagParsing: false,
		Flags: []cli.Flag{

			// 容器名称
			cli.StringFlag{
				Name:  "name",
				Usage: "container name",
			},
		},

		Action: func(ctx *cli.Context) error {
			name := ctx.String("name")
			if len(name) == 0 {
				return errors.New("LogsContainer.Action.Name.Null")
			}
			// 容器删除
			log.Infof("logs come on | %v", ctx.Args())
			return container.LogsContainer(name)
		},
	}
	// 构建镜像
	commitCommand = cli.Command{
		Name:            "commit",
		Usage:           "build images",
		SkipFlagParsing: false,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "name",
				Usage: "image name",
			},
		},

		Action: func(ctx *cli.Context) error {
			imageName := ctx.String("name")
			// 容器删除
			log.Infof("commit come on | %v", ctx.Args())
			images.CommitImage(imageName)
			return nil
		},
	}

	// 停止容器
	stopCommand = cli.Command{
		Name:            "stop",
		Usage:           "stop container",
		SkipFlagParsing: false,


		Action: func(ctx *cli.Context) error {
			// 容器删除
			log.Infof("stop come on | %v", ctx.Args())
			args := ctx.Args()
			if len(args) != 1 {
				log.Errorf("Stop.Len.Args!=1")
				return errors.Errorf("Missing container name")
			}

			return container.StopContainer(args[0])
		},
	}
	// 创建网络
	networkCommand = cli.Command{
		Name:         "network",
		Usage:        "network",

		Subcommands:            []cli.Command{
			{
				Name: "create",
				Usage: "create a container network",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "driver",
						Usage: "network driver",
					},
					cli.StringFlag{
						Name:  "subnet",
						Usage: "subnet cidr",
					},
				},
				Action:func(context *cli.Context) error {
					if len(context.Args()) < 1 {
						return errors.Errorf("Missing network name")
					}
					if err := cni.InitNetworkList();err != nil{
						return err
					}
					err := cni.CreateNetwork(context.String("driver"), context.String("subnet"), context.Args()[0])
					if err != nil {
						return errors.WithMessage(err, "create network error")
					}
					return nil
				},
			},
			{
				Name: "list",
				Usage: "list container network",
				Action:func(context *cli.Context) error {
					if err := cni.InitNetworkList();err != nil{
						return err
					}
					cni.ListNetwork()
					return nil
				},
			},
			{
				Name: "remove",
				Usage: "remove container network",
				Action:func(context *cli.Context) error {
					if len(context.Args()) < 1 {
						return errors.Errorf("Missing network name")

					}
					if err := cni.InitNetworkList();err != nil{
						return err
					}
					err := cni.DeleteNetwork(context.Args()[0])
					if err != nil {
						return errors.WithMessage(err, "remove network error")
					}
					return nil
				},
			},
		},
	}
)

type RunArgs struct {
	Tty            bool
	CmdArray       []string
	Image          string
	Volume         string
	Name           string
	Env            []string
	ResourceConfig *subsystems.ResourceConfig
}

func Run(args *RunArgs) {
	log.Infof("Run.Params | %v | %v | %v | %s", args.Tty, args.CmdArray, args.ResourceConfig, args.Image)
	meta, _ := container.GetContainerMeta(args.Name)

	if meta != nil {
		log.Warnf("Run.ContainerExist | %v | %v | %v | %s | %v", args.Tty, args.CmdArray, args.ResourceConfig, args.Image, meta)
		return
	}

	parent, writePipe := container.NewParentProcess(args.Tty, args.Volume, args.Image, args.Name, args.Env)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("Run.Error | %v | %v", err, parent)
		return
	}
	//cgroupManger := cgroup.NewCGroupManager("mydocker-cgroup")
	//defer func() {
	//
	//	if err := cgroupManger.Destroy(); err != nil {
	//		log.Errorf("Run.Destroy | %v", err)
	//	}
	//}()
	//if err := cgroupManger.Set(res); err != nil {
	//	log.Errorf("Run.Set | %v", err)
	//
	//}
	//if err := cgroupManger.Apply(parent.Process.Pid); err != nil {
	//	log.Errorf("Run.Apply | %v", err)
	//
	//}
	if _, err := container.RecordContainerMeta(parent.Process.Pid, args.Name, args.Image, args.Volume, args.CmdArray); err != nil {
		log.Errorf("Run.RecordContainerMeta | %v", err)
		return
	}

	sendInitCommand(args.CmdArray, writePipe)
	if args.Tty {
		// ti参数前台等待
		// 前台等待台运行
		log.Infof("Run.Wait | %v", parent.Wait())
		_, workUrl := images.GetWriteWorkLayerOverlay(args.Name)
		container.DeleteWorkSpace(workUrl, args.Volume)
		_ = container.UpdateContainerStatus(args.Name, container.STOP)
		//container.DeleteContainerMeta(cname)
	}
}

var Commands = []cli.Command{
	initCommand,   // 初始化容器
	runCommand,    // 运行容器
	commitCommand, // 镜像打包
	rmCommand,     // 删除容器
	listCommand,   // 列表
	logsCommand,   // 查看日志
	execCommand,   // 执行命令
	stopCommand,   // 停止容器
	networkCommand,   // 创建网络
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	n, err := writePipe.WriteString(command)
	if err != nil {
		log.Errorf("sendInitCommand.ERROR | %v | %v", n, err)
	}
	log.Infof("sendInitCommand.Info | %v", writePipe.Close())
}
