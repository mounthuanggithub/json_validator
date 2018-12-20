package json

import "testing"

func TestValidate(t *testing.T) {
	str := `{"a":"b","c":false,"d":2}`
	err :=Validate(str)
	if err != nil {
		t.Error(err)
	}

	str = `{"a:b"}`
	err =Validate(str)
	t.Log(err)
	if err == nil {
		t.Error("valid failed")
	}
}