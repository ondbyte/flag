 # turbo_flag

a drop in replacement for flag package which is included in the core go, but with additional capabilities like 
- Writing command-line apps with subcommands
- Loading configuration file like json,yaml,toml.
- Binding variable/s to values from a configuration file
- Loading `.env` files
- Binding variable/s to environment variable/s
- Enumeration of the values of the flag
- Short alias for a flag

etc.
### **Sub-commands**
example: a git program with commit and remote sub-commands
```go
import "github.com/ondbyte/turbo_flag"

func main() {
	flag.MainCmd(
		"git",
		"a version control implemented in golang",
		flag.ContinueOnError,
		func(git flag.Cmd, args []string) {
			help := git.Bool("help", false, "help message")
			git.SubCmd("commit", "commits the changes with a message", func(commitCmd flag.Cmd, args []string) {
				var branch string
				commitCmd.StringVar(&branch, "branch", "", "branch name to work", commitCmd.Cfg("branch.name"), commitCmd.Alias("b"), commitCmd.Env("BRANCH", "MAIN_BRNCH"))

				err := commitCmd.Parse(args)
				if err != nil {
					panic(err)
				}
				if branch == "" {
					fmt.Println("'branch' is required")
				} else {
					fmt.Println("commited branch ", branch)
				}
			})
			git.SubCmd(
				"remote",
				"adds a remote with name",
				func(remoteCmd flag.Cmd, args []string) {
					var name string
					remoteCmd.StringVar(&name, "name", "", "remote name work with", remoteCmd.Alias("n"))
					err := remoteCmd.Parse(args)
					if err != nil {
						panic(err)
					}
					if name == "" {
						fmt.Println("'name' is required")
					} else {
						fmt.Println("added remote ", name)
					}
				},
			)
			//lets try to commit with branch as argument
			err := git.Parse(args)
			if err != nil {
				panic(err)
			}
			if *help {
				helpStr, err := git.GetDefaultUsage()
				if err != nil {
					panic(err)
				}
				fmt.Println(helpStr)
			}
		},
	)
}


```