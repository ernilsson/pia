package exchange

import "testing"

func Test_VariableSet_SubstituteLine_GivenNothingToSubstitute_ReturnsPristineLine(t *testing.T) {
	set := make(VariableSet)
	set["name"] = "tester"

	line := "this is a line with no substitution"
	processed, err := set.SubstituteLine(line)

	if err != nil {
		t.Errorf("received unexpected error: %s", err)
		return
	}
	if line != processed {
		t.Errorf("expected line '%s' but got: %s", line, processed)
	}
}

func Test_VariableSet_SubstituteLine_GivenNonExistentKeyWithDefault_SubstitutesWithDefault(t *testing.T) {
	set := make(VariableSet)
	line := "this is a ${var.type:line} with default substitution"
	processed, err := set.SubstituteLine(line)

	if err != nil {
		t.Errorf("received unexpected error: %s", err)
		return
	}
	expected := "this is a line with default substitution"
	if processed != expected {
		t.Errorf("expected line '%s' but got: %s", expected, processed)
	}
}

func Test_VariableSet_SubstituteLine_GivenExistingKeyWithoutDefault_SubstitutesWithValue(t *testing.T) {
	set := make(VariableSet)
	set["type"] = "line"
	line := "this is a ${var.type} with var substitution"
	processed, err := set.SubstituteLine(line)

	if err != nil {
		t.Errorf("received unexpected error: %s", err)
		return
	}
	expected := "this is a line with var substitution"
	if processed != expected {
		t.Errorf("expected line '%s' but got: %s", expected, processed)
	}
}

func Test_VariableSet_SubstituteLine_GivenExistingKeyWithDefault_SubstitutesWithValue(t *testing.T) {
	set := make(VariableSet)
	set["type"] = "line"
	line := "this is a ${var.type:default} with var substitution"
	processed, err := set.SubstituteLine(line)

	if err != nil {
		t.Errorf("received unexpected error: %s", err)
		return
	}
	expected := "this is a line with var substitution"
	if processed != expected {
		t.Errorf("expected line '%s' but got: %s", expected, processed)
	}
}

func Test_VariableSet_SubstituteLine_GivenNonExistentKeyWithoutDefault_ReturnsError(t *testing.T) {
	set := make(VariableSet)
	line := "this is a ${var.type} with var substitution"
	_, err := set.SubstituteLine(line)

	if err == nil {
		t.Errorf("expected error to be returned but got nil")
		return
	}
}

func Test_VariableSet_SubstituteLine_GivenMalformedKey_ReturnsError(t *testing.T) {
	set := make(VariableSet)
	line := "this is a ${var.type:default:another_default} with var substitution"
	_, err := set.SubstituteLine(line)

	if err == nil {
		t.Errorf("expected error to be returned but got nil")
		return
	}
}

func Test_VariableSet_SubstituteLine_GivenSubstitutionWithAndWithoutDefault_Substitutes(t *testing.T) {
	set := make(VariableSet)
	set["sub"] = "substitution"
	line := "this is a ${var.type:line} with var ${var.sub}"
	processed, err := set.SubstituteLine(line)

	if err != nil {
		t.Errorf("received unexpected error: %s", err)
		return
	}
	expected := "this is a line with var substitution"
	if processed != expected {
		t.Errorf("expected line '%s' but got: %s", expected, processed)
	}
}
