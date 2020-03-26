package conf

import (
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func SetDefaultCommand(root, setDef *cobra.Command) {
	//set node is default command
	if runCommand, args, err := root.Find(os.Args[1:]); err == nil {
		if runCommand == root {
			runCommand.InitDefaultHelpFlag()
			if help, err := runCommand.PersistentFlags().GetBool("help"); err == nil && help {
				//show help
			} else {
				idx := 1
				for _, arg := range args {
					if strings.HasPrefix(arg, "-") {
						flagName := strings.TrimLeft(arg, "-")
						hasValue := strings.Index(flagName, "=")
						if hasValue > 0 {
							flagName = flagName[:hasValue]
						}
						if f := root.PersistentFlags().Lookup(flagName); f != nil {
							if f.Value.Type() == "bool" || hasValue > 0 {
								idx += 1
							} else if f.Value.String() != "" {
								idx += 2
							}
							continue
						} else if f = root.PersistentFlags().ShorthandLookup(flagName); f != nil {
							if f.Value.Type() == "bool" || hasValue > 0 {
								idx += 1
							} else if f.Value.String() != "" {
								idx += 2
							}
							continue
						}
					}
					break
				}
				os.Args = append(os.Args[:idx], append([]string{setDef.Name()}, os.Args[idx:]...)...)
			}
		}
	}
}
