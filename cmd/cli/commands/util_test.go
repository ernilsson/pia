package commands

import "testing"

func Test_ParseKeyValues_GivenSinglePair_ReturnsParsedPair(t *testing.T) {
	kv, err := ParseKeyValues([]string{"username=pia"})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	val, ok := kv["username"]
	if !ok {
		t.Errorf("missing key: username")
		return
	}
	if val != "pia" {
		t.Errorf("unexpected value: %s", val)
	}
}

func Test_ParseKeyValues_GivenMultipleEqualSigns_CutsByFirstEqualSign(t *testing.T) {
	kv, err := ParseKeyValues([]string{"username==pia="})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	val, ok := kv["username"]
	if !ok {
		t.Errorf("missing key: username")
		return
	}
	if val != "=pia=" {
		t.Errorf("unexpected value: %s", val)
	}
}

func Test_ParseKeyValues_GivenBadDelimiter_ReturnsError(t *testing.T) {
	_, err := ParseKeyValues([]string{"username:pia"})
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}
