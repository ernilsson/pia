package profile

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	ErrNoActiveProfileSet     = errors.New("no active profile set")
	ErrBadActiveProfileFormat = errors.New("bad active profile file format")
)

func Bootstrap(wd string) error {
	f, err := os.Create(fmt.Sprintf("%s/.profile", wd))
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	return nil
}

func SetActiveProfileName(wd string, profile string) error {
	f, err := os.OpenFile(fmt.Sprintf("%s/.profile", wd), os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

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
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	raw, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	profile := strings.TrimSpace(string(raw))
	if profile == "" {
		return "", ErrNoActiveProfileSet
	}
	if len(strings.Split(profile, " ")) > 1 {
		return "", ErrBadActiveProfileFormat
	}
	return profile, nil
}

func New(name string) Profile {
	profile := make(Profile)
	profile["profile"] = name
	return profile
}

type Profile map[string]any

func (p Profile) Name() string {
	for k, v := range p {
		if k == "profile" {
			return v.(string)
		}
	}
	return ""
}

func (p Profile) SetName(name string) {
	p["profile"] = name
}

func (p Profile) Get(key string) (any, error) {
	v, ok := p[key]
	if !ok {
		return nil, fmt.Errorf("profile value '%s' not found", key)
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

func (p Profile) SubstituteLine(line string) (string, error) {
	regx := regexp.MustCompile("\\$\\{profile\\..+?}")
	if !regx.MatchString(line) {
		return line, nil
	}
	regx = regexp.MustCompile("\\$\\{profile\\.(.+?)}")
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

func (p Profile) SubstituteLines(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	scn := bufio.NewScanner(bytes.NewReader(data))
	for scn.Scan() {
		line, err := p.SubstituteLine(scn.Text())
		if err != nil {
			return nil, err
		}
		if _, err := fmt.Fprintln(buf, line); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
