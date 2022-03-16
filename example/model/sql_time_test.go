package model_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/go-msvc/api/example/model"
)

func Test1(t *testing.T) {
	T1 := "2022-04-01 10:11:12"
	expValue := []uint8{50, 48, 50, 50, 45, 48, 52, 45, 48, 49, 32, 49, 48, 58, 49, 49, 58, 49, 50}

	t1, _ := time.Parse("2006-01-02 15:04:05", T1)

	s1 := model.SqlTime(t1)
	v1, ok := s1.Value().([]uint8)
	if !ok {
		t.Fatalf("Value -> %T != []uint8", s1.Value())
	}
	t.Logf("v1: %+v", v1)
	t.Logf("expValue: %+v", expValue)
	if len(v1) != len(expValue) {
		t.Fatalf("len %d != %d", len(v1), len(expValue))
	}
	for i, b := range expValue {
		if v1[i] != b {
			t.Fatalf("byte[%d] %v != %v", i, b, v1[i])
		}
	}
	j1, err := json.Marshal(s1)
	if err != nil {
		t.Fatal(err)
	}
	if string(j1) != "\""+T1+"\"" {
		t.Fatalf("json %s != %s", string(j1), T1)
	}

	var s2 model.SqlTime
	if err := json.Unmarshal([]byte("\""+T1+"\""), &s2); err != nil {
		t.Fatal(err)
	}
	if s2 != s1 {
		t.Fatalf("%v != %v", s1, s2)
	}

	if err := s2.Scan(s1.Value()); err != nil {
		t.Fatal(err)
	}
	if s2 != s1 {
		t.Fatalf("%v != %v", s1, s2)
	}
}
