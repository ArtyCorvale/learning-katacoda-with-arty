package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"io/ioutil"
	"path/filepath"
	"strconv"

)

// docker run image <cmd> <params>
// go run main.go (compiles and run executable)
// needs a command run, process <cmd> and arbitrary <params>

func main () {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("bad command")
	}
}

func run () {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr {
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS, // Need to unshare mount namespace because mounts are shared with host by default. This prevents mount command clutter.
	}
	
	cmd.Run()

}

func child () {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cg()

	syscall.Sethostname([]byte("container"))
	syscall.Chroot("/home/arty/ubuntu-fs")
	syscall.Chdir("/")
	syscall.Mount("proc", "proc", "proc", 0, "")

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	cmd.Run()

	syscall.Unmount("/proc", 0)

}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	err := os.Mkdir(filepath.Join(pids, "arty"), 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	must(ioutil.WriteFile(filepath.Join(pids, "arty/pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exists
	must(ioutil.WriteFile(filepath.Join(pids, "arty/notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "arty/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must (err error) {
		if err != nil {
			panic(err)
		}
}