package profile

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type Store interface {
	Must(Profile, error) Profile
	SetActive(name string) error
	LoadActive() (Profile, error)

	Load(name string) (Profile, error)
	Save(profile Profile) error
	Delete(name string) error
}

func Marshall(profile Profile) ([]byte, error) {
	buf := new(bytes.Buffer)
	for key, val := range profile {
		if _, err := fmt.Fprintln(buf, fmt.Sprintf("%s: %s", key, val)); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func Must(store FileStore, err error) FileStore {
	if err != nil {
		panic(err)
	}
	return store
}

func NewFileStore(wd string) (FileStore, error) {
	if err := bootstrap(wd, ".profile"); err != nil {
		return FileStore{}, err
	}
	if err := bootstrap(wd, ".profiles"); err != nil {
		return FileStore{}, err
	}
	return FileStore{wd: wd}, nil
}

func bootstrap(wd string, filename string) error {
	_, err := os.Stat(path.Join(wd, filename))
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	f, err := os.Create(path.Join(wd, filename))
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}(f)
	return nil
}

type FileStore struct {
	wd string
}

func (f FileStore) SetActive(name string) error {
	file, err := os.OpenFile(f.activeFilePath(), os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}(file)
	if err := file.Truncate(0); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(file, name); err != nil {
		return err
	}
	return nil
}

func (f FileStore) active() (string, error) {
	file, err := os.Open(f.activeFilePath())
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}(file)

	raw, err := io.ReadAll(file)
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

func (f FileStore) activeFilePath() string {
	return fmt.Sprintf("%s/.profile", f.wd)
}

func (f FileStore) LoadActive() (Profile, error) {
	active, err := f.active()
	if err != nil {
		return Profile{}, err
	}
	return f.Load(active)
}

func (f FileStore) Must(profile Profile, err error) Profile {
	if err != nil {
		panic(err)
	}
	return profile
}

func (f FileStore) Load(name string) (Profile, error) {
	profiles, err := f.load()
	if err != nil {
		return Profile{}, err
	}
	profile, ok := profiles[name]
	if !ok {
		return Profile{}, ErrProfileNotFound
	}
	return profile, nil
}

func (f FileStore) load() (map[string]Profile, error) {
	file, err := os.Open(f.storeFilePath())
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}(file)
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	scn := bufio.NewScanner(bytes.NewBuffer(content))
	profiles := make(map[string]Profile)
	profile := make(Profile)
	for scn.Scan() {
		line := scn.Text()
		if strings.HasPrefix(line, "%%") {
			profiles[profile.Name()] = profile
			profile = make(Profile)
			continue
		}
		key, val, err := f.parse(scn, line)
		if err != nil {
			return nil, err
		}
		profile[key] = val
	}
	return profiles, nil
}

func (f FileStore) parse(scn *bufio.Scanner, line string) (string, string, error) {
	key, val, ok := strings.Cut(line, ": ")
	if !ok {
		return "", "", ErrInvalidProfileLine
	}
	for strings.HasSuffix(val, "\\ ") {
		val = strings.TrimSuffix(val, "\\ ")
		if !scn.Scan() {
			return "", "", ErrInvalidProfileLine
		}
		val += strings.TrimSpace(scn.Text())
	}
	return key, val, nil
}

func (f FileStore) Save(profile Profile) error {
	profiles, err := f.load()
	if err != nil {
		return err
	}
	profiles[profile.Name()] = profile
	return f.write(profiles)
}

func (f FileStore) Delete(name string) error {
	profiles, err := f.load()
	if err != nil {
		return err
	}
	delete(profiles, name)
	return f.write(profiles)
}

func (f FileStore) write(profiles map[string]Profile) error {
	file, err := os.OpenFile(f.storeFilePath(), os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}(file)
	if err := file.Truncate(0); err != nil {
		return err
	}
	for _, p := range profiles {
		data, err := Marshall(p)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprint(file, string(data)); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(file, "%%"); err != nil {
			return err
		}
	}
	return nil
}

func (f FileStore) storeFilePath() string {
	return fmt.Sprintf("%s/.profiles", f.wd)
}
