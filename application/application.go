package application

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

var (
	ErrExit              = errors.New("exit")
	ErrCommandRegistered = errors.New("command has been registered!")
)

type Context struct {
	terminal *term.Terminal
	app      *Application
}

type HandlerFunc func(*Context) error

type Application struct {
	commands map[string]interface{}
}

type Command struct {
	Name        string
	Handler     HandlerFunc
	Usage       string
	Description string
}

func NewApp(prompt string) *Application {
	app := &Application{
		commands: map[string]interface{}{},
	}
	app.RegisterCommand("exit", "", "quit application", func(_ *Context) error {
		return ErrExit
	})
	app.RegisterCommand("help", "", "list all commands", func(context *Context) error {
		context.info(context.app.Commands())
		return nil
	})
	return app
}

func (app *Application) RegisterCommand(cmd, usage, description string, handler HandlerFunc) {
	if cmd == "" {
		return
	}
	cmdMap := app.commands
	cmds := strings.Split(cmd, " ")
	for i, sc := range cmds {
		c := strings.TrimSpace(sc)
		if i < len(cmds)-1 {
			if m, ok := cmdMap[c]; ok {
				mm, ok := m.(map[string]interface{})
				if !ok {
					panic(ErrCommandRegistered)
				}
				cmdMap = mm
			} else {
				m := make(map[string]interface{})
				cmdMap[c] = m
				cmdMap = m
			}
		} else {
			_, ok := cmdMap[c]
			if ok {
				panic(ErrCommandRegistered)
			}
			cmdMap[c] = Command{
				Name:        sc,
				Handler:     handler,
				Usage:       usage,
				Description: description,
			}
		}
	}
}

func (app *Application) Run() {
	state, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	terminal := term.NewTerminal(os.Stdin, color.GreenString("neo> "))
	context := &Context{
		terminal: terminal,
		app:      app,
	}
	defer func() {
		err = term.Restore(int(os.Stdin.Fd()), state)
		if err != nil {
			panic(err)
		}
	}()
loop:
	for {
		line, err := terminal.ReadLine()
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		line = strings.TrimRight(line, "\n")
		whitespaces := regexp.MustCompile(`\s+`)
		cmd := whitespaces.ReplaceAllString(line, " ")
		if cmd == "" {
			continue
		}
		cmds := strings.Split(cmd, " ")
		cmdMap := app.commands
		for i, subCmd := range cmds {
			if i < len(cmds)-1 {
				if m, ok := cmdMap[subCmd]; ok {
					mm, ok := m.(map[string]interface{})
					if !ok {
						context.info("command not found")
						break
					}
					cmdMap = mm
				} else {
					context.info("command not found")
					break
				}
			} else {
				h, ok := cmdMap[subCmd]
				if !ok {
					context.info("command not found")
					break
				}
				command, ok := h.(Command)
				if !ok {
					context.info("command not found")
					break
				}
				err := command.Handler(context)
				if err != nil {
					if errors.Is(err, ErrExit) {
						break loop
					} else {
						context.error(fmt.Sprintf("%s", err))
					}
				}
			}
		}
	}
}

func (ctx *Context) info(msg string) {
	ctx.terminal.Write([]byte(msg + "\n"))
}

func (ctx *Context) error(msg string) {
	ctx.terminal.Write(append(append(ctx.terminal.Escape.Red, append([]byte("Err: "), ctx.terminal.Escape.Reset...)...), []byte(msg+"\n")...))
}

func (app *Application) Commands() string {
	return commandMapToString(app.commands, 0)
}

func commandMapToString(cmdMap map[string]interface{}, deepth int) string {
	str := ""
	for scmd, m := range cmdMap {
		mm, ok := m.(map[string]interface{})
		if ok {
			str += strings.Repeat(" ", deepth) + scmd + ":\n" + commandMapToString(mm, deepth+1)
		} else {
			command := m.(Command)
			str += strings.Repeat(" ", deepth) + command.String() + "\n"
		}
	}
	return str
}

func (cmd *Command) String() string {
	return fmt.Sprintf("%s %s %s", cmd.Name, cmd.Usage, cmd.Description)
}
