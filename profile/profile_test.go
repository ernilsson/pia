package profile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
)

func Test_SetActiveProfile_WhenNoProfileIsSet_WritesActiveProfileToFile(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to get current working directory: %s", err)
		return
	}
	if err := Bootstrap(wd); err != nil {
		t.Errorf("failed to bootstrap profile: %s", err)
		return
	}
	err = SetActiveProfileName(wd, "dev")
	if err != nil {
		t.Errorf("unexpected error when setting profile: %s", err)
		return
	}
	defer func() {
		if err := os.Remove(fmt.Sprintf("%s/.profile", wd)); err != nil {
			panic(err)
		}
	}()

	f, err := os.Open(fmt.Sprintf("%s/.profile", wd))
	if err != nil {
		t.Errorf("failed to open profile file after setting: %s", err)
		return
	}
	content, err := io.ReadAll(f)
	if err != nil {
		t.Errorf("failed to read .profile content: %s", err)
		return
	}
	if string(content) != "dev" {
		t.Errorf("unexpected active profile content: %s", string(content))
	}
}

func Test_SetActiveProfile_WhenProfileIsAlreadySet_ReplacesExistingProfile(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to get current working directory: %s", err)
		return
	}
	if err := Bootstrap(wd); err != nil {
		t.Errorf("failed to bootstrap profile: %s", err)
		return
	}
	defer func() {
		if err := os.Remove(fmt.Sprintf("%s/.profile", wd)); err != nil {
			panic(err)
		}
	}()

	err = SetActiveProfileName(wd, "dev")
	if err != nil {
		t.Errorf("unexpected error when setting profile: %s", err)
		return
	}
	err = SetActiveProfileName(wd, "prod")
	if err != nil {
		t.Errorf("unexpected error when setting profile: %s", err)
		return
	}

	f, err := os.Open(fmt.Sprintf("%s/.profile", wd))
	if err != nil {
		t.Errorf("failed to open profile file after setting: %s", err)
		return
	}
	content, err := io.ReadAll(f)
	if err != nil {
		t.Errorf("failed to read .profile content: %s", err)
		return
	}
	if string(content) != "prod" {
		t.Errorf("unexpected active profile content: %s", string(content))
	}
}

func Test_ActiveProfile_WhenNoProfileIsSet_ReturnsNoActiveProfileError(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to get current working directory: %s", err)
		return
	}
	if err := Bootstrap(wd); err != nil {
		t.Errorf("failed to bootstrap profile: %s", err)
		return
	}
	defer func() {
		if err := os.Remove(fmt.Sprintf("%s/.profile", wd)); err != nil {
			panic(err)
		}
	}()

	_, err = ActiveProfileName(wd)
	if err == nil {
		t.Errorf("expected error but got nil")
		return
	}
	if !errors.Is(err, ErrNoActiveProfileSet) {
		t.Errorf("unexpected error returned: %s", err)
	}
}

func Test_ActiveProfile_WhenProfileIsSet_ReturnsActiveProfileName(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to get current working directory: %s", err)
		return
	}
	if err := Bootstrap(wd); err != nil {
		t.Errorf("failed to bootstrap profile: %s", err)
		return
	}
	defer func() {
		if err := os.Remove(fmt.Sprintf("%s/.profile", wd)); err != nil {
			panic(err)
		}
	}()
	if err := SetActiveProfileName(wd, "dev"); err != nil {
		t.Errorf("failed to set active profile: %s", err)
		return
	}

	profile, err := ActiveProfileName(wd)
	if err != nil {
		t.Errorf("got unexpected error: %s", err)
		return
	}
	if profile != "dev" {
		t.Errorf("expected active profile to be 'dev' but got: %s", profile)
	}
}

func Test_ActiveProfile_GivenBadlyFormattedFile_ReturnsBadActiveProfileFileFormatError(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to get current working directory: %s", err)
		return
	}
	if err := Bootstrap(wd); err != nil {
		t.Errorf("failed to bootstrap profile: %s", err)
		return
	}
	defer func() {
		if err := os.Remove(fmt.Sprintf("%s/.profile", wd)); err != nil {
			panic(err)
		}
	}()
	f, err := os.OpenFile(fmt.Sprintf("%s/.profile", wd), os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Errorf("unexpected error when opening profile file for writing: %s", err)
		return
	}
	if _, err := f.WriteString("this is\npoor\tformatting"); err != nil {
		t.Errorf("unexpected error when writing profile file: %s", err)
		return
	}

	_, err = ActiveProfileName(wd)
	if err == nil {
		t.Errorf("expected an error but got none")
		return
	}
	if !errors.Is(err, ErrBadActiveProfileFileFormat) {
		t.Errorf("expected bad active profile format error but got: %s", err)
	}
}

func Test_Profile_SubstituteLine_GivenLineWithMultipleSubstitutions_ReturnsSubstitutedLine(t *testing.T) {
	profile := Profile{}
	profile.Put("env", "dev")
	profile.Put("age", 25)

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
	profile.Put("env", "dev")
	profile.Put("age", 25)

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
	profile.Put("env", "dev")

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
