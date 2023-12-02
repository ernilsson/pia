package profile

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ErrInvalidProfileLine = errors.New("invalid profile line")
	ErrProfileNotFound    = errors.New("profile not found")
)

type ProviderFunc func() ([]byte, error)

func FileProvider(wd string) ProviderFunc {
	return func() ([]byte, error) {
		f, err := os.OpenFile(fmt.Sprintf("%s/.profiles", wd), os.O_RDONLY, os.ModeAppend)
		if err != nil {
			return nil, err
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}(f)
		return io.ReadAll(f)
	}
}

type ParserFunc func(name string) (Profile, error)

func Unmarshal(name string, p ProviderFunc) (Profile, error) {
	return UnmarshalRecordJar(p)(name)
}

func UnmarshalActive() (Profile, error) {
	wd, err := os.Getwd()
	if err != nil {
		return Profile{}, err
	}
	ap, err := ActiveProfileName(wd)
	if err != nil {
		return Profile{}, err
	}
	return Unmarshal(ap, FileProvider(wd))
}

func UnmarshalRecordJar(p ProviderFunc) ParserFunc {
	return func(name string) (Profile, error) {
		content, err := p()
		if err != nil {
			return Profile{}, err
		}
		s := bufio.NewScanner(bytes.NewBuffer(content))
		profile := Profile{}
		for s.Scan() {
			line := s.Text()
			if strings.HasPrefix(line, "%%") && profile.Name() == name {
				return profile, nil
			} else if strings.HasPrefix(line, "%%") {
				profile = Profile{}
				continue
			}
			key, val, err := parseRecordJarLine(s, line)
			if err != nil {
				return Profile{}, err
			}
			profile.Put(key, val)
		}
		return Profile{}, ErrProfileNotFound
	}
}

func parseRecordJarLine(s *bufio.Scanner, line string) (key, val string, err error) {
	key, val, ok := strings.Cut(line, ": ")
	if !ok {
		return "", "", ErrInvalidProfileLine
	}
	val, err = getFullRecordJarVal(s, val)
	if err != nil {
		return "", "", err
	}
	return key, val, nil
}

func getFullRecordJarVal(s *bufio.Scanner, val string) (string, error) {
	for strings.HasSuffix(val, "\\ ") {
		val = strings.TrimSuffix(val, "\\ ")
		if !s.Scan() {
			return "", ErrInvalidProfileLine
		}
		val += strings.TrimSpace(s.Text())
	}
	return val, nil
}
