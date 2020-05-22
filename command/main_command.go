package command

import (
	log "github.com/Sirupsen/logrus"
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
			Run(
				ttl,
				arr, ctx.String("image"),
				ctx.String("v"),
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
	commitCommand = cli.Command{
		Name:            "commit",
		Usage:           "Init container process run user's rocess in container.Do not call it outside",
		SkipFlagParsing: false,
		Flags: []cli.Flag{
			cli.BoolFlag{
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

func Run(tty bool, cmdArray []string, image, volume string, res *subsystems.ResourceConfig) {
	log.Infof("Run.Params | %v | %v | %v | %s", tty, cmdArray, res, image)
	parent, writePipe := container.NewParentProcess(tty, volume, image)
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

	sendInitCommand(cmdArray, writePipe)
	if tty {
		// ti参数前台等待
		// 前台等待台运行
		log.Infof("Run.Wait | %v", parent.Wait())
		_, workUrl := images.GetWriteWorkLayerOverlay()
		container.DeleteWorkSpace(workUrl, volume)
	}
}

var Commands = []cli.Command{
	initCommand,   // 初始化容器
	runCommand,    // 运行容器
	commitCommand, // 镜像打包
	rmCommand,     // 删除容器
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