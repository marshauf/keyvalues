package keyvalues

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	files, err := filepath.Glob("PICSProductInfo*")
	if err != nil {
		t.Fatal(err)
	}
	for i, file := range files {
		t.Logf("Testing file[%d:%d] %s\n", i, len(files), file)

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
