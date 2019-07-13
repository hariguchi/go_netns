package netns

import (
	"net"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestGetNewSetDelete(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var (
		err    error
		newNs  NsHandle
		origNs NsHandle
	)

	if origNs, err = GetMyHandle(); err == nil {
		t.Logf("origNs: %s", origNs.String())
	} else {
		t.Fatal(err)
	}
	if newNs, err = New(); err == nil {
		t.Logf("newNs: %s", newNs.String())
		if origNs.Equal(newNs) {
			t.Fatalf("old: %s, new: %s", origNs.String(), newNs.String())
		}
		ifs, _ := net.Interfaces()
		t.Logf("interfaces: %v", ifs)
		if err := SetByHandle(origNs); err != nil {
			t.Fatal(err)
		}
		if err := newNs.Close(); err != nil {
			t.Fatal(err)
		}
		if newNs.IsOpen() {
			t.Fatal("newNs is still open after close", newNs)
		}
		ns, err := GetMyHandle()
		if err != nil {
			t.Fatal(err)
		}
		if !ns.Equal(origNs) {
			t.Fatal("Failed to reset namespace", origNs, newNs, ns)
		}
	} else {
		t.Fatal(err)
	}
}

func testDeleteByName(t *testing.T, name string) {
	path := path.Join(NsRunDir, name)
	if _, err := os.Stat(path); err == nil {
		if err := DeleteByName(name); err == nil {
			t.Logf("deleted namespace %s", name)
		} else {
			t.Fatal(err)
		}
		if ns, err := GetByName(name); err == nil {
			t.Fatalf("namespace %s still exists: %v", name, ns.String())
		}
	} else if os.IsNotExist(err) {
		t.Logf("GetFromPath(%s): %v", name, err)
	} else {
		t.Fatalf("os.Stat(%s): %v", path, err)
	}
}

func TestGetNewDeleteByName(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	name := "nsTest"

	testDeleteByName(t, name)
	desc, err := NewByName(name)
	if err == nil {
		t.Logf("new netns: %v", desc)
	} else {
		t.Fatal(err)
	}
	if ns, err := GetByName(name); err == nil {
		t.Logf("namespace %s (NsHandle %v)", name, ns.String())
	} else {
		t.Fatal(err)
	}
	testDeleteByName(t, name)
}

func TestNsDesc(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	name := "nsDescTest"
	var origNs NsHandle

	if h, err := GetMyHandle(); err == nil {
		origNs = h
		t.Logf("origNs: %s", origNs.String())
	} else {
		t.Fatal(err)
	}
	testDeleteByName(t, name)
	if _, err := GetByName(name); err == nil {
		t.Fatalf("namespace %s still exists", name)
	} else if !IsNotExist(err) {
		t.Fatal(err)
	}

	desc, err := NewByName(name)
	if err == nil {
		t.Logf("NsDesc.String(): %s", desc.String())
	} else {
		t.Fatal(err)
	}

	other := desc.Copy()
	if !desc.Equal(&other) {
		t.Fatalf("orig: %v, new: %v", desc, other)
	}
	s1 := desc.UniqueId()
	s2 := other.UniqueId()
	if s1 != s2 {
		t.Fatalf("orig: %s, new: %s", s1, s2)
	}
	if err := SetByHandle(origNs); err != nil {
		t.Fatalf("Failed to switch namespace: %v", err)
	}
	if err := desc.Close(); err != nil {
		t.Fatalf("Failed to close %s: %v", desc, err)
	}
	if desc.IsOpen() {
		t.Fatalf("namespace %s is still open after close", name)
	}
	if err := desc.Delete(); err == nil {
		t.Logf("namespace %s is deleted", name)
	} else {
		t.Fatal(err)
	}
	if ns, err := GetMyHandle(); err == nil {
		if !ns.Equal(origNs) {
			t.Fatal("Failed to reset namespace", origNs, desc, ns)
		}
	} else {
		t.Fatal(err)
	}
}

func TestSetByName(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	name := "testSetByName"
	origH, err := GetMyHandle()
	if err != nil {
		t.Fatal(err)
	}
	origIfs, _ := net.Interfaces()
	t.Logf("original namespace: interfaces: %v", origIfs)

	//
	// switch to new namesapace
	//
	var ns NsDesc
	ns, err = SetByName(name)
	if err != nil {
		t.Fatal(err)
	}
	defer ns.Delete()

	nsIfs, _ := net.Interfaces()
	t.Logf("namespace %s: interfaces: %v", ns.Name, nsIfs)

	if h, err := GetMyHandle(); err == nil {
		if h.UniqueId() == origH.UniqueId() {
			t.Fatalf("failed to switch namespace")
		}
	} else {
		t.Fatal(err)
	}
	err = SetByHandle(origH)
	if err != nil {
		t.Fatalf("failed to switch back to original namespace")
	}
	if h, err := GetMyHandle(); err == nil {
		if h.UniqueId() != origH.UniqueId() {
			t.Fatalf("unique id: failed to switch namespace")
		}
	} else {
		t.Fatal(err)
	}
}
