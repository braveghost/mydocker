package command

import (
	log "github.com/sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"mydocker/container"
	"mydocker/images"
	"mydocker/subsystems"
	"os"
	"strings"
)

var (
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
		},
		Action: func(ctx *cli.Context) error {
			// 启动容器
			if len(ctx.Args()) < 1 {
				return errors.New("Missing container command")
			}
			var arr []string
			for _, v := range ctx.Args() {
				arr = append(arr, v)
			}

			ttl := ctx.Bool("ti")
			detach := ctx.Bool("d")
			if ttl && detach {
				return errors.New("ttl && detach")
			}
			name := ctx.String("name")
			Run(
				ttl,
				arr, ctx.String("image"),
				ctx.String("v"),
				name,
				&subsystems.ResourceConfig{
					MemoryLimit: ctx.String("m"),
					CpuSet:      ctx.String("cpuset"),
					CpuShare:    ctx.String("cpushare"),
				})
			return nil
		},
	}

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

	rmCommand = cli.Command{
		Name:            "rm",
		Usage:           "Init container process run user's rocess in container.Do not call it outside",
		SkipFlagParsing: false,
		Action: func(ctx *cli.Context) error {
			// 容器删除
			log.Infof("rm come on | %v", ctx.Args())
			return container.RemoveContainer()
		},
	}

	// ps命令
	listCommand = cli.Command{
		Name:            "ps",
		Usage:           "Init container process run user's rocess in container.Do not call it outside",
		SkipFlagParsing: false,
		Action: func(ctx *cli.Context) error {
			// 容器删除
			log.Infof("ps come on | %v", ctx.Args())
			return container.ListContainer()
		},
	}	// ps命令
	execCommand = cli.Command{
		Name:            "exec",
		Usage:           "Init container process run user's rocess in container.Do not call it outside",
		SkipFlagParsing: false,
		Action: func(ctx *cli.Context) error {
			// 容器删除
			log.Infof("exec come on | %v", ctx.Args())

			// 如果环境变量不为空就返回
			if os.Getenv(container.ENV_EXEC_PID) != ""{
				log.Errorf("pid callback pid | %d", os.Getgid())
				return nil
			}
			args := ctx.Args()
			if len(args) < 2{
				log.Errorf("Exec.Len.Args!=2")
				return errors.Errorf("Missing container name or command")
			}
			name := args[0]

			return container.ExecContainer(name ,args.Tail())
		},
	}
	// 查看日志
	logsCommand = cli.Command{
		Name:            "logs",
		Usage:           "Init container process run user's rocess in container.Do not call it outside",
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
		Usage:           "Init container process run user's rocess in container.Do not call it outside",
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
)

func Run(tty bool, cmdArray []string, image, volume, name string, res *subsystems.ResourceConfig) {
	log.Infof("Run.Params | %v | %v | %v | %s", tty, cmdArray, res, image)
	parent, writePipe := container.NewParentProcess(tty, volume, image, name)
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
	cname, err := container.RecordContainerMeta(parent.Process.Pid, name, cmdArray)
	if err != nil {
		log.Errorf("Run.RecordContainerMeta | %v", err)
		return
	}

	sendInitCommand(cmdArray, writePipe)
	if tty {
		// ti参数前台等待
		// 前台等待台运行
		log.Infof("Run.Wait | %v", parent.Wait())
		_, workUrl := images.GetWriteWorkLayerOverlay()
		container.DeleteWorkSpace(workUrl, volume)
		container.DeleteContainerMeta(cname)
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
