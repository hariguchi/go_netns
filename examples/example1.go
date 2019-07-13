package main

import (
	"fmt"
	netns "github.com/hariguchi/go_netns"
	"net"
	"os"
	"runtime"
)

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var (
		defNs netns.NsDesc
		ns    netns.NsDesc
		err   error
		name  string = "nsTest"
	)

	//
	// Save the current namespace
	//
	defNs, err = netns.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Get(): %v\n", err)
		os.Exit(1)
	}
	//
	// Close the current namespace when exit.
	// The namespace will continue to exist.
	//
	defer defNs.Close()

	//
	// Dump interfaces in the default namespace
	//
	if ifs, err := net.Interfaces(); err == nil {
		fmt.Printf("default namespace: interfaces: %v\n", ifs)
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: net.Interfaces(): %v\n", err)
		os.Exit(1)
	}

	//
	// Create a new namespace and switch to it
	//
	ns, err = netns.SetByName(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: netns.SetByName(%s): %v\n", name, err)
		os.Exit(1)
	}
	//
	// Delete namespace when exit.
	// the namespace will continue to exist if ns.Close() is used
	//
	defer ns.Delete()

	//
	// Executions are done under new namespace from here.
	// Let us dump interfaces in the new namespace for example.
	// There must be only loopback interface.
	//
	if ifs, err := net.Interfaces(); err == nil {
		fmt.Printf("namespace %s: interfaces: %v\n", ns.Name, ifs)
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: net.Interfaces(): %v\n", err)
		os.Exit(1)
	}

	//
	// Switch back to the default namespace
	//
	if err := defNs.Set(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Set(): %v\n", err)
		os.Exit(1)
	}

	//
	// Double check if we switched back
	//
	if h, err := netns.GetMyHandle(); err == nil {
		if h.UniqueId() != defNs.UniqueId() {
			fmt.Fprintf(os.Stderr,
				"ERROR: failed to switch back to default namespace\n")
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: GetMyHandle(): %v\n", err)
		os.Exit(1)
	}
}
