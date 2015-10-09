package main

import (
	//    "bufio"
	"bytes"
	//  "flag"
	"fmt"
	"io/ioutil"
	"os"
	//"path"
	"strings"
	"text/template"
	"time"

    "gopkg.in/alecthomas/kingpin.v2"
)

type NewCommand struct {
}

type TemplateParams struct {
	Name   string
	Action string
}

func (c *NewCommand) Help() string {
	helpText := `
    Usage: sql-migrate new <name> [options] ...

    Generate new migration scripts.

    Options:

    -config=dbconfig.yml   Configuration file to use.
    -env="development"     Environment.

    `
	return strings.TrimSpace(helpText)
}

func (c *NewCommand) Synopsis() string {
	return "Create new migration scripts"
}

func BuildTemplate(params *TemplateParams) string {
	text := `-- This is generated from a template
-- +migrate Up
--{{.Action}} <TYPE> {{.Name}}
----------------------------------------
-- +migrate Down
--{{.Action}} <TYPE> {{.Name}}
`
	t := template.New("something")
	t, err := t.Parse(text)
	check(err)
	buf := new(bytes.Buffer)
	t.Execute(buf, params)
	return buf.String()
}

var (
	newCmd = kingpin.Command("new", "Generate a new Migration script").Default()
	name   = newCmd.Arg("name", "Name of SQL entity to migrate").Required().String()
	action = newCmd.Arg("action", "Action of the migration [create|alter|drop]").Enum("create", "alter", "drop")
	naming = newCmd.Flag("naming", "Output file naming convention [epoch|iso|counter|none]").Short('n').Default("epoch").Enum("epoch", "iso", "counter", "none")
)

func (c *NewCommand) Run(args []string) int {
	//kingpin.UsageTemplate(kingpin.DefaultUsageTemplate).Version("1.0").Author("Joe Bloggs")
    KingpinConfigFlags()
	kingpin.Parse()

	params := TemplateParams{Name: *name, Action: *action}
	t := BuildTemplate(&params)
	filename := GetFilename(&params)
	var dir string
	env, err := GetEnvironment()
	if err != nil {
		dir = "./"
	} else {
		dir = env.Dir
	}
	if !strings.HasSuffix(dir, fmt.Sprintf("%c", os.PathSeparator)) {
		dir += fmt.Sprintf("%c", os.PathSeparator)
	}
	filepath := dir + filename
	err = ioutil.WriteFile(filepath, []byte(t), 0644)
	check(err)
	fmt.Printf("New migration script generated: %s\n\n", filepath)

	return 0
}

func GetFilename(params *TemplateParams) string {
	var prefix string
	prefix = fmt.Sprintf("%v", time.Now().Unix())
	switch *naming {
	case "epoch":
	default:
		prefix = fmt.Sprintf("%v", time.Now().Unix())
	case "iso":
		prefix = time.Now().Format(time.RFC3339)
	case "counter":
		env, err := GetEnvironment()
		if err != nil {
			prefix = fmt.Sprintf("%08d", 1)
			break
		}
		files, _ := ioutil.ReadDir(env.Dir)
		prefix = fmt.Sprintf("%08d", len(files))
	case "none":
		return fmt.Sprintf("%s-%s.sql", params.Name, params.Action)
	}

	return fmt.Sprintf("%s-%s-%s.sql", prefix, params.Name, params.Action)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
