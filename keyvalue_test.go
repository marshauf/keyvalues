package keyvalues

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	files, err := filepath.Glob("./testdata/PICSProductInfo_App_*")
	if err != nil {
		t.Fatal(err)
	}
	for i, file := range files {
		t.Logf("Testing file(%d/%d) %s\n", i, len(files), file)

		b, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		kv, err := Unmarshal(b)
		if err != nil {
			t.Error(err)
		}

		if kv == nil {
			t.Error("KeyValue is nil")
		}

		if kv.Key != "appinfo" {
			t.Errorf("Root key is not appinfo: %s", kv)
		}
	}
}

func TestUnmarshalBinary(t *testing.T) {
	files, err := filepath.Glob("./testdata/PICSProductInfo_Package_*")
	if err != nil {
		t.Fatal(err)
	}
	for i, file := range files {
		t.Logf("Testing file(%d/%d) %s\n", i, len(files), file)

		r, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}

		// For some reason the ProductInfo has a leading uint32
		_, err = ReadUint32(r)
		if err != nil {
			t.Fatal(err)
		}

		kv, err := UnmarshalBinary(r)
		if err != nil {
			t.Error(err)
		}

		if kv == nil {
			t.Error("KeyValue is nil")
		}

		//t.Log(kv.String())
	}

	// From steam cache
	files, err = filepath.Glob("./testdata/*.bin")
	if err != nil {
		t.Fatal(err)
	}
	for i, file := range files {
		t.Logf("Testing file(%d/%d) %s\n", i, len(files), file)

		b, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}

		kv, err := UnmarshalBinary(b)
		if err != nil {
			t.Error(err)
		}

		if kv == nil {
			t.Error("KeyValue is nil")
		}

		t.Log(kv.String())
	}
}
