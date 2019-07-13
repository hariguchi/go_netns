// Extension of github.com/vishvananda/netns
//  Added the following functions:
//    - create a namespace by name (NewByName)
//    - close (delete) a namespace by name (CloseByName)
//    - related functions with the name of namespace
//
// The following functions can be used as it is
//
// (ns *NsHandle) Close()
// (ns *NsHandle) Equal()
// (ns *NsHandle) IsOpen()
// (ns *NsHandle) String()
// (ns *NsHandle) UniqueId()

package netns

import (
	"fmt"
	vnetns "github.com/vishvananda/netns"
	"os"
	"path"
	"syscall"
)

const (
	NsRunDir  = "/var/run/netns"
	SelfNetNs = "/proc/self/ns/net"
)

type NsHandle = vnetns.NsHandle
type NsDesc struct {
	Handle NsHandle
	Name   string
}

// DeleteByName deletes the namespace whose name is `name'
// in: name Name of the namespace to be deleted
// return: nil if success
//         non-nil otherwise
func DeleteByName(name string) error {
	banner := fmt.Sprintf("CloseByName(%s): ", name)
	path := path.Join(NsRunDir, name)
	err := syscall.Unmount(path, syscall.MNT_DETACH)
	if err != nil {
		return fmt.Errorf("%sUnmount(): %v", banner, err)
	}
	if err = syscall.Unlink(path); err != nil {
		return fmt.Errorf("%sUnlink(): %v", banner, err)
	}
	return nil
}

// IsNotExist returns true if the string of `err' is
// "no such file or directory"
// in: err Error
// return true if the string of `err' is "no such file or directory"
//        false otherwise
func IsNotExist(err error) bool {
	if fmt.Sprint(err) == "no such file or directory" {
		return true
	}
	return false
}

// Get retrieves a descriptor to the current thread's network namespace
// return: 1. Valid NsDesc if success
//            Invalid otherwise
//         2. nil if success
//            non-nil otherwise
func Get() (NsDesc, error) {
	var (
		d   NsDesc
		err error
	)
	d.Handle, err = GetMyHandle()
	return d, err
}

// GetByName retrieves a descriptor to the network namespace
// associated with the given name
// in: name Name of the namespace
// return: 1. Valid NsDesc if success
//            Invalid otherwise
//         2. nil if success
//            non-nil otherwise
func GetByName(name string) (NsDesc, error) {
	var err error
	ns := NsDesc{Name: name}
	ns.Handle, err = GetHandleByName(name)
	return ns, err
}

// GetFromDocker retrieves a handle to the network namespace
// associated with the given docker identifier.
// in: id Docker Identifier
// return: 1. Valid NsHandle if success
//            Invalid otherwise
//         2. nil if success
//            non-nil otherwise
func GetFromDocker(id string) (NsHandle, error) {
	return vnetns.GetFromDocker(id)
}

// GetMyHandle retrieves a handle to the current thread's network namespace
// return: 1. NsHandle if success
//            Invalid otherwise
//         2. nil if success
//            non-nil otherwise
func GetMyHandle() (NsHandle, error) {
	return vnetns.Get()
}

// GetHandleByName retrieves a handle to the network namespace
// associated with the given name
// in: name Name of the namespace
// return: 1. NsHandle if success
//            Invalid otherwise
//         2. nil if success
//            non-nil otherwise
func GetHandleByName(name string) (NsHandle, error) {
	path := path.Join(NsRunDir, name)
	return vnetns.GetFromPath(path)
}

// GetFromPath retrieves a handle to the network namespace
// associated with the given full path
// in: name The full path of the namespace
// return: 1. NsHandle if success
//            Invalid otherwise
//         2. nil if success
//            non-nil otherwise
func GetFromPath(path string) (NsHandle, error) {
	return vnetns.GetFromPath(path)
}

// GetFromPid retrieves a handle to the network namespace
// associated with the given process ID
// in: pid Process ID
// return: 1. NsHandle if success
//            Invalid otherwise
//         2. nil if success
//            non-nil otherwise
func GetFromPid(pid int) (NsHandle, error) {
	return vnetns.GetFromPid(pid)
}

// GetFromThread retrieves a handle to the network namespace
// associated with the given process and thread ID
// in: pid Process ID
//     tid Thread ID
// return: 1. NsHandle if success
//            Invalid otherwise
//         2. nil if success
//            non-nil otherwise
func GetFromThread(pid, tid int) (NsHandle, error) {
	return vnetns.GetFromThread(pid, tid)
}

// New creates a new network namespace, sets it as the current network
// namespace, and returns a handle to it
// return: 1. Valid NsHandle if success
//            Invalid otherwise
//         2. nil if success
//            non-nil otherwise
func New() (NsHandle, error) {
	return vnetns.New()
}

// AddByName adds a network namespace `name' and returns its descriptor.
// If it already exists, it simply returns its descriptor.
// AddByName does *NOT* switch the namespace.
// in: name Name of the namespace to be added
// return: 1. Descriptor to network namespace `name' if success
//            Invalid otherwise
//         2. nil if success
//            non-nil otherwise
func AddByName(name string) (NsDesc, error) {
	var ns NsDesc
	banner := fmt.Sprintf("AddByName(%s): ", name)

	h, err := GetMyHandle()
	if err != nil {
		return NsDesc{}, fmt.Errorf("%sGetMyHandle(): %v", banner, err)
	}
	ns, err = SetByName(name)
	if err != nil {
		return NsDesc{}, fmt.Errorf("%sSetByName(): %v", banner, err)
	}
	err = SetByHandle(h)
	if err != nil {
		return ns, fmt.Errorf("%sSetByHandle(): %v", banner, err)
	}
	return ns, nil
}

// NewByName creates a new network namespace, sets it as the current
// network namespace, binds the given name to it, switches to the new
// network namespace, and returns a descriptor to the new network namespace
// in: name Name of the network namespace to be created
// return: 1. Valid NsDesc if success
//            Invalid otherwise
//         2. nil if success
//            non-nil otherwise
func NewByName(name string) (NsDesc, error) {
	banner := fmt.Sprintf("NewByName(%s): ", name)
	netNsPath := path.Join(NsRunDir, name)
	desc := NsDesc{Name: name}

	os.Mkdir(NsRunDir, 0755)
	fd, err := syscall.Open(netNsPath, syscall.O_RDONLY|syscall.O_CREAT|syscall.O_EXCL, 0)
	if err != nil {
		return desc, fmt.Errorf("%s%v", banner, err)
	}
	syscall.Close(fd)
	if desc.Handle, err = New(); err != nil {
		return desc, fmt.Errorf("%s%v", banner, err)
	}
	if err := syscall.Mount(SelfNetNs, netNsPath, "none", syscall.MS_BIND, ""); err != nil {
		return desc, fmt.Errorf("%s%v", banner, err)
	}
	return desc, nil
}

// None retrieves a closed NsHandle
// return: Closed NsHandle (-1)
func None() NsHandle {
	return vnetns.None()
}

// Set switches to the namespace specified by `ns'
// in: ns Descriptor to be set to the current namespace
// return: nil if success
//         non-nil otherwise
func Set(ns NsDesc) error {
	return SetByHandle(ns.Handle)
}

// SetByHandle switches to the namespace specified by `h'
// return: nil if success
//         non-nil otherwise
func SetByHandle(h NsHandle) error {
	return vnetns.Set(h)
}

// SetByName switches to the namespace specified by `name'
// return: nil if success
//         non-nil otherwise
func SetByName(name string) (NsDesc, error) {
	ns, err := GetByName(name)
	if err == nil {
		err = ns.Set()
		return ns, err
	} else if IsNotExist(err) {
		return NewByName(name) // NewByName() sets namespace
	} else {
		return ns, err
	}
}

// Copy returns a copy of `ns'
// return: a copy of `ns'
func (ns *NsDesc) Copy() NsDesc {
	return NsDesc{Name: ns.Name, Handle: ns.Handle}
}

// Close closes the namespace associated with `ns'.
// The namespace continues to exist.
// return: nil if success
//         non-nil otherwise
func (ns *NsDesc) Close() error {
	return ns.Handle.Close()
}

// Delete the namespace associated with `ns'
// return: nil if success
//         non-nil otherwise
func (ns *NsDesc) Delete() error {
	ns.Close()
	return DeleteByName(ns.Name)
}

// Equal returns true if `ns' and `other' are the same network namespace
// in: other Pointer to a namespace
// return: true if `ns' and `other' are identical
//         false otherwise
func (ns *NsDesc) Equal(other *NsDesc) bool {
	return ns.Handle.Equal(other.Handle)
}

// IsOpen returns true if `ns' is not closed
// return: true If `ns' is closed
//         false Otherwise
func (ns *NsDesc) IsOpen() bool {
	return ns.Handle.IsOpen()
}

// Set switches to the namespace specified by `ns'
// return: nil if success
//         non-nil otherwise
func (ns *NsDesc) Set() error {
	return SetByHandle(ns.Handle)
}

// String returns the name of namespace `ns', file descriptor,
// ifindex, and inode
func (ns *NsDesc) String() string {
	return fmt.Sprintf("Name: %s, Handle: %s", ns.Name, ns.Handle)
}

// UniqueId returns a string that uniquely identifies the namespace
// associated with `ns'
func (ns *NsDesc) UniqueId() string {
	return ns.Handle.UniqueId()
}
