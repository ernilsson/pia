package environment

import "testing"

func Test_SubstituteLines_GivenNothingToSubstitute_ReturnsCopyOfSource(t *testing.T) {
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

func Test_GivenSingleLevelSubstitution_SubstitutesWithEnvironmentValue(t *testing.T) {
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

func Test_GivenMultiLevelSubstitution_SubstitutesWithEnvironmentValue(t *testing.T) {
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

func Test_GivenMissingSingleLevelKey_ReturnsError(t *testing.T) {
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

func Test_GivenMissingNestedKey_ReturnsError(t *testing.T) {
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

func Test_MultiLineSubstitution_SubstitutesAllLines(t *testing.T) {
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
