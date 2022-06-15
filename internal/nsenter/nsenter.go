package nsenter

/*
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

// __attribut__((constructor)) means this function will be called right after the package is imported
// in other words, this function will run before the program run
__attribute__((constructor)) void enter_namespace(void) {
	char *go_docker_pid;
	// get pid from env
	go_docker_pid = getenv("go_docker_pid");
	if (go_docker_pid) {
		fprintf(stdout, "got env: go_docker_pid=%s\n", go_docker_pid);
	} else {
		// using env to control whether run this bunch of codes
		// if env not exist, than this function will not run
		// so that command other than "docker exec" will not run this bunch of cgo code
		fprintf(stdout, "missing env: go_docker_pid\n");
		return;
	}
	char *go_docker_cmd;
	go_docker_cmd = getenv("go_docker_cmd");
	if (go_docker_cmd) {
		fprintf(stdout, "got env: go_docker_cmd=%s\n", go_docker_cmd);
	} else {
		fprintf(stdout, "missing env: go_docker_cmd\n");
		return;
	}


	// five namespaces that need to enter
	char *namespaces[] = {"ipc", "uts", "net", "pid", "mnt"};

	char nspath[1024];

	int i; // old c compiler style
	for (i = 0; i < 5; i++) {
	    sprintf(nspath, "/proc/%s/ns/%s", go_docker_pid, namespaces[i]);
	    int fd = open(nspath, O_RDONLY);
	    // call setns to enter namespace
	    if (setns(fd, 0) == -1) {
	        fprintf(stderr, "setns %s fails: %s\n", namespaces[i], strerror(errno));
	    } else {
	        fprintf(stdout, "setns %s succeed\n", namespaces[i]);
	    }
	    close(fd);
	}
	// after enter the namespaces, run the cmd
	int res = system(go_docker_cmd);
	exit(0);
	return;
}

*/
import "C"
