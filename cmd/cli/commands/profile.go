package commands

import (
	"errors"
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
		store := profile.NewFileStore(wd)
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
		return nil
	},
}

func SetProfileValuesHandler(cmd *cobra.Command, p profile.Profile) error {
	set, err := cmd.Flags().GetStringSlice("set")
	if err != nil {
		return err
	}
	sets, err := ExtractKeyValues(set)
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
		store := profile.NewFileStore(wd)
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
	Aliases:    []string{"cp"},
	Args:       cobra.ExactArgs(2),
	ArgAliases: []string{"src", "dst"},
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		store := profile.NewFileStore(wd)
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

func init() {
	root.AddCommand(prof)
	prof.Flags().StringSliceP("set", "s", nil, "defines a key-value pair to be set on the currently active profile, ex: --set username=root")
	prof.Flags().StringSliceP("delete", "d", nil, "defines a key to be deleted from the currently active profile, ex: --del username")
	prof.AddCommand(sw)

	cp.Flags().BoolP("merge", "m", false, "merges the source profile into the destination profile if the destination profile already exists")
	prof.AddCommand(cp)
}
