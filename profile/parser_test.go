package profile

import (
	"testing"
)

func Test_RecordJarParser_GivenSingleProfile_ReturnsProfile(t *testing.T) {
	data := []byte("Name: dev\nURL: http://localhost:8080\n%%\n")
	p := ProviderFunc(func() ([]byte, error) {
		return data, nil
	})

	parser := UnmarshalRecordJar(p)
	profile, err := parser("dev")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	name := profile.Name()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if name != "dev" {
		t.Errorf("expected 'Name' to equal 'dev' but got: %s", name)
		return
	}
	url, err := profile.GetString("URL")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if url != "http://localhost:8080" {
		t.Errorf("expected 'URL' to equal 'http://localhost:8080' but got: %s", url)
		return
	}
}

func Test_RecordJarParser_GivenMultipleProfiles_ReturnsCorrectProfile(t *testing.T) {
	data := []byte("Name: dev\nURL: http://localhost:8080\n%%\nName: prod\nURL: https://prod.example.com\n%%\n")
	p := ProviderFunc(func() ([]byte, error) {
		return data, nil
	})

	parser := UnmarshalRecordJar(p)
	profile, err := parser("dev")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	name := profile.Name()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if name != "dev" {
		t.Errorf("expected 'Name' to equal 'dev' but got: %s", name)
		return
	}
	url, err := profile.GetString("URL")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if url != "http://localhost:8080" {
		t.Errorf("expected 'URL' to equal 'http://localhost:8080' but got: %s", url)
		return
	}
}

func Test_RecordJarParser_GivenProfileValueWithMultipleLines_ReturnsLinesAsSingleValue(t *testing.T) {
	data := []byte("Name: dev\nURL: http://local\\ \n\thost:8080\n%%\n")
	p := ProviderFunc(func() ([]byte, error) {
		return data, nil
	})

	parser := UnmarshalRecordJar(p)
	profile, err := parser("dev")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	url, err := profile.GetString("URL")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if url != "http://localhost:8080" {
		t.Errorf("expected 'URL' to equal 'http://localhost:8080' but got: %s", url)
		return
	}
}

func Test_RecordJarParser_GivenInvalidLineInProfile_ReturnsError(t *testing.T) {
	data := []byte("Name:dev\nURL: http://localhost:8080\n%%\n")
	p := ProviderFunc(func() ([]byte, error) {
		return data, nil
	})

	parser := UnmarshalRecordJar(p)
	_, err := parser("dev")
	if err == nil {
		t.Errorf("expected error but got none")
	}

	data = []byte("Name: dev\n\nURL: http://localhost:8080\n%%\n")
	parser = UnmarshalRecordJar(p)
	_, err = parser("dev")
	if err == nil {
		t.Errorf("expected error but got none")
	}
}
