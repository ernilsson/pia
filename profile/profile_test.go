package profile

import (
	"testing"
)

func Test_Profile_SubstituteLine_GivenLineWithMultipleSubstitutions_ReturnsSubstitutedLine(t *testing.T) {
	profile := Profile{}
	profile["env"] = "dev"
	profile["age"] = 25

	line := "This test is in ${profile.env} and I am ${profile.age} years old"
	actual, err := profile.SubstituteLine(line)
	if err != nil {
		t.Errorf("unexpected error returned: %s", err)
		return
	}
	expected := "This test is in dev and I am 25 years old"
	if actual != expected {
		t.Errorf("expected '%s' to be returned but got: %s", expected, actual)
	}
}

func Test_Profile_SubstituteLine_GivenLineWithSingleSubstitution_ReturnsSubstitutedLine(t *testing.T) {
	profile := Profile{}
	profile["env"] = "dev"
	profile["age"] = 25

	line := "This test is in ${profile.env}"
	actual, err := profile.SubstituteLine(line)
	if err != nil {
		t.Errorf("unexpected error returned: %s", err)
		return
	}
	expected := "This test is in dev"
	if actual != expected {
		t.Errorf("expected '%s' to be returned but got: %s", expected, actual)
	}
}

func Test_Profile_SubstituteLine_GivenLineWithNoSubstitutions_ReturnsOriginalString(t *testing.T) {
	profile := Profile{}

	line := "This test is in dev"
	actual, err := profile.SubstituteLine(line)
	if err != nil {
		t.Errorf("unexpected error returned: %s", err)
		return
	}
	expected := "This test is in dev"
	if actual != expected {
		t.Errorf("expected '%s' to be returned but got: %s", expected, actual)
	}
}

func Test_Profile_SubstituteLine_WhenSubstitutionVariableDoesNotExist_ReturnsError(t *testing.T) {
	profile := Profile{}
	profile["env"] = "dev"

	line := "This test is in ${profile.env} and I am ${profile.age} years old"
	actual, err := profile.SubstituteLine(line)
	if err == nil {
		t.Errorf("expected error to be returned but got none")
		return
	}
	if actual != "" {
		t.Errorf("expected empty value to be returned but got: %s", actual)
	}
}
