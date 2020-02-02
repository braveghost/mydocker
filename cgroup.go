package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

const cgroupMemoryHierarchyMount = "/sys/fs/cgroup/memory"

func main() {
	if os.Args[0] == "/proc/self/exe" {
		fmt.Printf("current pid %d", syscall.Getpid())
		fmt.Println()
		cmd := exec.Command("sh", "-c", `stress --vm-bytes 200m --vm-keep -m 1`)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatal("Error stress", err)
			os.Exit(1)
		}
	}

	cmd := exec.Command("/proc/self/exe")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 11,
				HostID:      0,
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 11,
				HostID:      0,
				Size:        1,
			},
		},
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatal("Error exe", err)
	}

	fmt.Println("%v", cmd.Process.Pid)

	_ = os.Mkdir(path.Join(cgroupMemoryHierarchyMount, "testMemoryLimit"), 0755)
	_ = ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testMemoryLimit", "tasks"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
	_ = ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testMemoryLimit", "memory.limit_in_bytes"), []byte("100m"), 0644)
	_, _ = cmd.Process.Wait()
}
