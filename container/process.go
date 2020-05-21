package container

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"io/ioutil"
	"mydocker/images"
	"mydocker/setting"
	"mydocker/subsystems"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	log.Info("readUserCommand | %s", msgStr)
	return strings.Split(msgStr, " ")
}

func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return errors.New("Run container get user command error, cmdArray is nil")
	}

	subsystems.SetUpMount()

	p, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("Exec loop path error %v", err)
		return err
	}
	log.Infof("Find path %s | %v", p, cmdArray)
	//dmf := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	//
	//_ = syscall.Mount("/proc", "/proc", "proc", uintptr(dmf), "")
	//
	if err := syscall.Exec(p, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf("RunContainerInitProcess.Exec.Error | %s", err.Error())
	}

	return nil
}

func NewParentProcess(tty bool, volume string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

	}
	cmd.ExtraFiles = []*os.File{readPipe}

	rootUrl := setting.EImagesPath
	// images.NewWorkSpaceAufs
	// aufs是老版本了，所以就没试过了，这里只用了overlay2
	upperDir, workDir := images.NewWorkSpaceOverlay(rootUrl)
	log.Infof("NewParentProcess.CreateMountPointOverlay | %s | %s", upperDir, workDir)
	cmd.Dir = workDir
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

func DeleteWorkSpace(rootUrl,mntUrl,volume string)  {

}

func NewWorkSpace(rootUrl, volume string)  {
	upperDir, workDir := images.NewWorkSpaceOverlay(rootUrl)
	if len(volume) != 0{
		vu := volumesUrlExtract(volume)
		if len(vu) != 2{
			log.Errorf("NewWorkSpace.Split.Len | %v", volume)
			return
		}
		left, right := vu[0], vu[1]
		if len(left) != 0 && len(right) != 0{
			MountVolume(rootUrl, workDir, left, right)
		}
	}
}
func volumesUrlExtract(volume string)[]string  {
	return  strings.Split(volume,":")
}


func MountVolume(rootUrl,mntUrl,left, right string){
	if err := os.Mkdir(left, 0777); err != nil{
		log.Infof("MountVolume.Mkdir.Left | %s | %v", left, err)
	}
	cvu := path.Join(mntUrl,right)
	if err := os.Mkdir(cvu, 0777); err != nil{
		log.Infof("MountVolume.Mkdir.Right | %s | %v", right, err)
	}
	// 把宿主机文件目录挂载到容器挂载点

	cmd := exec.Command("mount", left, cvu)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err  := cmd.Run();err != nil{
		log.Errorf("MountVolume.Command.Run | %v", err)
	}
}