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
