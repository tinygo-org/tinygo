package compileopts

import (
	"os"
	"reflect"
	"testing"
)

func TestLoadTarget(t *testing.T) {
	_, err := LoadTarget(&Options{Target: "arduino"})
	if err != nil {
		t.Error("LoadTarget test failed:", err)
	}

	_, err = LoadTarget(&Options{Target: "notexist"})
	if err == nil {
		t.Error("LoadTarget should have failed with non existing target")
	}

	if !os.IsNotExist(err) {
		t.Error("LoadTarget failed for wrong reason:", err)
	}
}

func TestOverrideProperties(t *testing.T) {
	baseAutoStackSize := true
	base := &TargetSpec{
		GOOS:             "baseGoos",
		CPU:              "baseCpu",
		CFlags:           []string{"-base-foo", "-base-bar"},
		BuildTags:        []string{"bt1", "bt2"},
		Emulator:         []string{"be1", "be2"},
		DefaultStackSize: 42,
		AutoStackSize:    &baseAutoStackSize,
	}
	childAutoStackSize := false
	child := &TargetSpec{
		GOOS:             "",
		CPU:              "chlidCpu",
		CFlags:           []string{"-child-foo", "-child-bar"},
		Emulator:         []string{"ce1", "ce2"},
		AutoStackSize:    &childAutoStackSize,
		DefaultStackSize: 64,
	}

	base.overrideProperties(child)

	if base.GOOS != "baseGoos" {
		t.Errorf("Overriding failed : got %v", base.GOOS)
	}
	if base.CPU != "chlidCpu" {
		t.Errorf("Overriding failed : got %v", base.CPU)
	}
	if !reflect.DeepEqual(base.CFlags, []string{"-base-foo", "-base-bar", "-child-foo", "-child-bar"}) {
		t.Errorf("Overriding failed : got %v", base.CFlags)
	}
	if !reflect.DeepEqual(base.BuildTags, []string{"bt1", "bt2"}) {
		t.Errorf("Overriding failed : got %v", base.BuildTags)
	}
	if !reflect.DeepEqual(base.Emulator, []string{"ce1", "ce2"}) {
		t.Errorf("Overriding failed : got %v", base.Emulator)
	}
	if *base.AutoStackSize != false {
		t.Errorf("Overriding failed : got %v", base.AutoStackSize)
	}
	if base.DefaultStackSize != 64 {
		t.Errorf("Overriding failed : got %v", base.DefaultStackSize)
	}

	baseAutoStackSize = true
	base = &TargetSpec{
		AutoStackSize:    &baseAutoStackSize,
		DefaultStackSize: 42,
	}
	child = &TargetSpec{
		AutoStackSize:    nil,
		DefaultStackSize: 0,
	}
	base.overrideProperties(child)
	if *base.AutoStackSize != true {
		t.Errorf("Overriding failed : got %v", base.AutoStackSize)
	}
	if base.DefaultStackSize != 42 {
		t.Errorf("Overriding failed : got %v", base.DefaultStackSize)
	}

}
