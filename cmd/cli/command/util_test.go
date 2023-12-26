package command

import (
	"os"
	"path"
	"testing"
)

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

func Test_DiscoverExchangeFile_GivenExistingFileAsInput_ReturnsAbsolutePathOfInput(t *testing.T) {
	f, err := os.Create("./configuration")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			t.Errorf("failed to close test file: %s", err)
		}
	}(f)
	defer func() {
		if err := os.Remove("./configuration"); err != nil {
			t.Errorf("failed to clean up test file: %s", err)
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	actual, err := DiscoverExchangeFile(path.Join(wd, "configuration"))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	expected := path.Join(wd, "configuration")
	if actual != expected {
		t.Errorf("expected '%s' but got: %s", expected, actual)
	}
}

func Test_DiscoverExchangeFile_WhenYMLIsOmitted_ReturnsAbsolutePathOfFileWithExtension(t *testing.T) {
	f, err := os.Create("./configuration.yml")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			t.Errorf("failed to close test file: %s", err)
		}
	}(f)
	defer func() {
		if err := os.Remove("./configuration.yml"); err != nil {
			t.Errorf("failed to clean up test file: %s", err)
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	actual, err := DiscoverExchangeFile(path.Join(wd, "configuration"))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	expected := path.Join(wd, "configuration.yml")
	if actual != expected {
		t.Errorf("expected '%s' but got: %s", expected, actual)
	}
}

func Test_DiscoverExchangeFile_WhenYAMLIsOmitted_ReturnsAbsolutePathOfFileWithExtension(t *testing.T) {
	f, err := os.Create("./configuration.yaml")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			t.Errorf("failed to close test file: %s", err)
		}
	}(f)
	defer func() {
		if err := os.Remove("./configuration.yaml"); err != nil {
			t.Errorf("failed to clean up test file: %s", err)
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	actual, err := DiscoverExchangeFile(path.Join(wd, "configuration"))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	expected := path.Join(wd, "configuration.yaml")
	if actual != expected {
		t.Errorf("expected '%s' but got: %s", expected, actual)
	}
}

func Test_DiscoverExchangeFile_GivenDirectoryWithConfigYAMLInside_ReturnsAbsolutePathOfConfig(t *testing.T) {
	if err := os.Mkdir("./req", 0777); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	f, err := os.Create("./req/config.yml")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			t.Errorf("failed to close test file: %s", err)
		}
	}(f)
	defer func() {
		if err := os.Remove("./req/config.yml"); err != nil {
			t.Errorf("failed to clean up test file: %s", err)
		}
		if err := os.Remove("./req"); err != nil {
			t.Errorf("failed to clean up test file: %s", err)
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	actual, err := DiscoverExchangeFile(path.Join(wd, "req"))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	expected := path.Join(wd, "req", "config.yml")
	if actual != expected {
		t.Errorf("expected '%s' but got: %s", expected, actual)
	}
}

func Test_DiscoverExchangeFile_GivenDirectoryWithoutConfigYAMLInside_ReturnsError(t *testing.T) {
	if err := os.Mkdir("./req", 0777); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	defer func() {
		if err := os.Remove("./req"); err != nil {
			t.Errorf("failed to clean up test file: %s", err)
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	_, err = DiscoverExchangeFile(path.Join(wd, "req"))
	if err == nil {
		t.Error("expected error but got nil")
		return
	}
}

func Test_DiscoverExchangeFile_GivenNonExistentFile_ReturnsError(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	_, err = DiscoverExchangeFile(path.Join(wd, "config"))
	if err == nil {
		t.Error("expected error but got nil")
		return
	}
}
