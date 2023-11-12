package environment

import (
	"strings"
	"testing"
)

func Test_Environment_SubstituteLines_GivenNothingToSubstituteThenReturnsCopyOfSource(t *testing.T) {
	env := Environment{}
	src := "{\n\t\"name\": \"Test\",\n\t\"age\": 56\n}\n"
	out, err := env.SubstituteLines(src)
	if err != nil {
		t.Error(err)
		return
	}
	if out != src {
		t.Errorf("pristine source does not equal output:\n%s\n%s", src, out)
	}
}

func Test_Environment_SubstituteLines_GivenSingleLevelSubstitutionThenSubstitutesWithEnvironmentValue(t *testing.T) {
	env := Environment{
		"age": 25,
	}
	src := "{\n\t\"name\": \"Test\",\n\t\"age\": ${env.age}\n}\n"
	out, err := env.SubstituteLines(src)
	if err != nil {
		t.Error(err)
		return
	}
	expected := "{\n\t\"name\": \"Test\",\n\t\"age\": 25\n}\n"
	if out != expected {
		t.Errorf("processed output does not equal expected:\n%s\n%s", expected, out)
	}
}

func Test_Environment_SubstituteLines_GivenMultiLevelSubstitutionThenSubstitutesWithEnvironmentValue(t *testing.T) {
	env := Environment{
		"person": map[string]any{
			"age": 25,
		},
	}
	src := "{\n\t\"name\": \"Test\",\n\t\"age\": ${env.person.age}\n}\n"
	out, err := env.SubstituteLines(src)
	if err != nil {
		t.Error(err)
		return
	}
	expected := "{\n\t\"name\": \"Test\",\n\t\"age\": 25\n}\n"
	if out != expected {
		t.Errorf("processed output does not equal expected:\n%s\n%s", expected, out)
	}
}

func Test_Environment_SubstituteLines_GivenMissingSingleLevelKeyThenReturnsError(t *testing.T) {
	env := Environment{}
	src := "{\n\t\"name\": \"Test\",\n\t\"age\": ${env.age}\n}\n"
	out, err := env.SubstituteLines(src)
	if err == nil {
		t.Errorf("expected error to be returned but none was")
	}
	if out != "" {
		t.Errorf("expected empty string but got '%s'", out)
	}
}

func Test_Environment_SubstituteLines_GivenMissingNestedKeyThenReturnsError(t *testing.T) {
	env := Environment{}
	src := "{\n\t\"name\": \"Test\",\n\t\"age\": ${env.person.age}\n}\n"
	out, err := env.SubstituteLines(src)
	if err == nil {
		t.Errorf("expected error to be returned but none was")
	}
	if out != "" {
		t.Errorf("expected empty string but got '%s'", out)
	}
}

func Test_Environment_SubstituteLines_GivenMultiLineSubstitutionThenSubstitutesAllLines(t *testing.T) {
	env := Environment{
		"person": map[string]any{
			"name": "E. R. Nilsson",
			"age":  25,
		},
	}
	src := "{\n\t\"name\": \"${env.person.name}\",\n\t\"age\": ${env.person.age}\n}\n"
	out, err := env.SubstituteLines(src)
	if err != nil {
		t.Error(err)
		return
	}
	expected := "{\n\t\"name\": \"E. R. Nilsson\",\n\t\"age\": 25\n}\n"
	if out != expected {
		t.Errorf("processed output does not equal expected:\n%s\n%s", expected, out)
	}
}

func Test_Load_GivenInvalidSource_ReturnsError(t *testing.T) {
	json := "{\"env\": \"dev\""
	env, err := Load(strings.NewReader(json))
	if err == nil {
		t.Errorf("expected non-nil error but got %s", err)
	}
	if env != nil {
		t.Errorf("expected nil result but got %+v", env)
	}
}

func Test_Load_GivenValidSource_ReturnsParsedEnvironment(t *testing.T) {
	json := "{\"env\": \"dev\", \"person\": {\"age\": 40}}"
	env, err := Load(strings.NewReader(json))
	if err != nil {
		t.Error(err)
	}
	if env["env"] != "dev" {
		t.Errorf("expected key 'env' to equal 'dev' but got %s", env["env"])
	}
	person := env["person"].(map[string]any)
	if person["age"] != float64(40) {
		t.Errorf("expected key 'person.age' to equal float64(40) but got %v (%T)", person["age"], person["age"])
	}
}
