package nsenter

// todo 注意去linux下编译

/*
#define _GNU_SOURCE
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>

// 一个包被引用，函数自动执行
__attribute__((constructor)) void enter_namespace(void) {
	char *mydocker_pid;
	// 从环境变量获取要进入的pid
	mydocker_pid = getenv("mydocker_pid");
	if (mydocker_pid) {
		fprintf(stdout, "got mydocker_pid=%s\n", mydocker_pid);
	} else {
		// 没有执行pid直接退出不向下执行
		fprintf(stdout, "missing mydocker_pid env skip nsenter");
		return;
	}
	char *mydocker_cmd;
	mydocker_cmd = getenv("mydocker_cmd");
	// 从环境变量获取要执行的命令
	if (mydocker_cmd) {
		fprintf(stdout, "got mydocker_cmd=%s\n", mydocker_cmd);
	} else {
		// 没有指定命令直接退出
		fprintf(stdout, "missing mydocker_cmd env skip nsenter");
		return;
	}
	int i;
	char nspath[1024];
	// 要进入的ns
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };

	for (i=0; i<5; i++) {
		// 拼接对应的路径
		sprintf(nspath, "/proc/%s/ns/%s", mydocker_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);

		if (setns(fd, 0) == -1) {
			// 调用setns进入
			fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	// 执行命令
	int res = system(mydocker_cmd);
	exit(0);
	return;
}
*/
import "C"