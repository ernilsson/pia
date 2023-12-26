package commands

import (
	"errors"
	"fmt"
	"github.com/ernilsson/pia/profile"
	"github.com/spf13/cobra"
	"os"
)

var prof = &cobra.Command{
	Use:     "profile",
	Short:   "manages current profile values",
	Aliases: []string{"prof", "pf"},
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		store := profile.Must(profile.NewFileStore(wd))
		p, err := store.LoadActive()
		if err != nil {
			return err
		}
		if err := SetProfileValuesHandler(cmd, p); err != nil {
			return err
		}
		if err := DeleteProfileValuesHandler(cmd, p); err != nil {
			return err
		}
		if err := store.Save(p); err != nil {
			return err
		}
		pr, err := cmd.Flags().GetBool("print")
		if err != nil {
			return err
		}
		if !pr {
			return nil
		}
		for key, val := range p {
			fmt.Printf("%s: %v\n", key, val)
		}
		return nil
	},
}

func SetProfileValuesHandler(cmd *cobra.Command, p profile.Profile) error {
	set, err := cmd.Flags().GetStringSlice("set")
	if err != nil {
		return err
	}
	sets, err := ParseKeyValues(set)
	if err != nil {
		return err
	}
	for key, val := range sets {
		p[key] = val
	}
	return nil
}

func DeleteProfileValuesHandler(cmd *cobra.Command, p profile.Profile) error {
	del, err := cmd.Flags().GetStringSlice("delete")
	if err != nil {
		return err
	}
	for _, key := range del {
		delete(p, key)
	}
	return nil
}

var sw = &cobra.Command{
	Use:        "switch",
	Aliases:    []string{"sw"},
	Short:      "sets the active profile for the current project",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"name"},
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		name := args[0]
		store := profile.Must(profile.NewFileStore(wd))
		p, err := store.Load(name)
		if err != nil && errors.Is(err, profile.ErrProfileNotFound) {
			p = profile.New(name)
		} else if err != nil {
			return err
		}
		if err := store.Save(p); err != nil {
			return err
		}
		if err := store.SetActive(p.Name()); err != nil {
			return err
		}
		return nil
	},
}

var cp = &cobra.Command{
	Use:        "copy",
	Short:      "copies the source profiles values to the destination profile",
	Aliases:    []string{"cp"},
	Args:       cobra.ExactArgs(2),
	ArgAliases: []string{"src", "dst"},
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		store := profile.Must(profile.NewFileStore(wd))
		sp, err := store.Load(args[0])
		if err != nil {
			return err
		}
		dst := args[1]
		merge, err := cmd.Flags().GetBool("merge")
		if err != nil {
			return err
		}
		var dp profile.Profile
		if merge {
			dp, err = store.Load(dst)
			if err != nil && errors.Is(err, profile.ErrProfileNotFound) {
				dp = profile.New(dst)
			} else if err != nil {
				return err
			}
		} else {
			dp = profile.New(dst)
		}
		for key, val := range sp {
			dp[key] = val
		}
		dp.SetName(dst)
		if err := store.Save(dp); err != nil {
			return err
		}
		return nil
	},
}

var rm = &cobra.Command{
	Use:        "remove",
	Aliases:    []string{"rm"},
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"name"},
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		store := profile.Must(profile.NewFileStore(wd))
		return store.Delete(args[0])
	},
}

func init() {
	Root.AddCommand(prof)
	prof.Flags().StringSliceP("set", "s", nil, "key-value pairs to be set on the currently active profile, ex: -s username=Root")
	prof.Flags().StringSliceP("delete", "d", nil, "keys to be deleted from the currently active profile, ex: -d username")
	prof.Flags().BoolP("print", "p", false, "prints the profile after all other potential changes have been applied")
	prof.AddCommand(sw)
	prof.AddCommand(rm)

	cp.Flags().BoolP("merge", "m", false, "merges the source profile into the destination profile if the destination profile already exists")
	prof.AddCommand(cp)
}
