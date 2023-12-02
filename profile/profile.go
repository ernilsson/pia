package profile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	ErrProfileValueNotFound       = errors.New("profile value not found")
	ErrNoActiveProfileSet         = errors.New("no active profile set")
	ErrBadActiveProfileFileFormat = errors.New("bad active profile file format")
)

func Bootstrap(wd string) error {
	f, err := os.Create(fmt.Sprintf("%s/.profile", wd))
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func SetActiveProfileName(wd string, profile string) error {
	f, err := os.OpenFile(fmt.Sprintf("%s/.profile", wd), os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	// TODO: Handle error on closing
	defer f.Close()

	// TODO: Handle if all bytes were not written
	if _, err := f.WriteString(profile); err != nil {
		return err
	}
	return nil
}

func ActiveProfileName(wd string) (string, error) {
	f, err := os.Open(fmt.Sprintf("%s/.profile", wd))
	if err != nil {
		return "", err
	}
	// TODO: Handle error on closing
	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	profile := strings.TrimSpace(string(raw))
	if profile == "" {
		return "", ErrNoActiveProfileSet
	}
	if len(strings.Split(profile, " ")) > 1 {
		return "", ErrBadActiveProfileFileFormat
	}
	return profile, nil
}

type Profile map[string]any

func (p Profile) Name() string {
	for k, v := range p {
		if k == "Name" {
			return v.(string)
		}
	}
	return ""
}

func (p Profile) Get(key string) (any, error) {
	v, ok := p[key]
	if !ok {
		return nil, ErrProfileValueNotFound
	}
	return v, nil
}

func (p Profile) GetString(key string) (string, error) {
	v, err := p.Get(key)
	if err != nil {
		return "", err
	}
	s, ok := v.(string)
	if !ok {
		return "", errors.New("tried to fetch non-string value as string")
	}
	return s, nil
}

func (p Profile) Put(key string, value any) {
	p[key] = value
}

func (p Profile) SubstituteLine(line string) (string, error) {
	regx := regexp.MustCompile("\\$\\{profile\\..+?\\}")
	if !regx.MatchString(line) {
		return line, nil
	}
	regx = regexp.MustCompile("\\$\\{profile\\.(.+?)\\}{1}")
	matches := regx.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		val, err := p.Get(match[1])
		if err != nil {
			return "", err
		}
		regx = regexp.MustCompile(fmt.Sprintf("\\$\\{profile.%s\\}", match[1]))
		line = regx.ReplaceAllString(line, fmt.Sprintf("%v", val))
	}
	return line, nil
}
