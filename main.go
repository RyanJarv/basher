package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mvdan/sh/syntax"
)

var (
	parser *syntax.Parser
	cmds   chan string
)

const commanddb = "/usr/share/command-not-found/commands.db"

func main() {
	flag.Parse()
	cmds = make(chan string)
	go findPrograms(cmds)

	parser = syntax.NewParser(syntax.KeepComments, syntax.Variant(syntax.LangBash))
	anyErr := false
	for _, path := range flag.Args() {
		if err := parseFile(path); err != nil {
			fmt.Fprintln(os.Stderr, err)
			anyErr = true
		}
	}
	if anyErr {
		os.Exit(1)
	}
}

func parseFile(path string) error {
	cont, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	prog, err := parser.Parse(bytes.NewReader(cont), path)
	if err != nil {
		return err
	}

	syntax.Walk(prog, getCommands)
	return nil
}

func getCommands(node syntax.Node) bool {
	switch node.(type) {
	case *syntax.CallExpr:
		p := node.(*syntax.CallExpr).Args[0].Parts[0]
		r := reflect.ValueOf(p)
		v := reflect.Indirect(r).FieldByName("Value")
		if v.IsValid() == false {
			log.Fatal("invalid value")
		}
		cmds <- v.String()
	}
	return true
}

func findPrograms(cmds chan string) {
	db, err := sql.Open("sqlite3", commanddb)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for {
		select {
		case cmd := <-cmds:
			typeCmd := exec.Command("bash", "-c", fmt.Sprintf("type %s", cmd))
			err := typeCmd.Run()
			if err == nil {
				fmt.Printf("\u2713 %s\n", cmd)
				if err != nil {
					log.Fatal(err)
				}
				continue
			}

			sqlStmt := fmt.Sprintf(`select p.name from packages as p join commands as c on p.pkgID = c.pkgID where c.command = "%s" limit 10;`, cmd)
			rows, err := db.Query(sqlStmt)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()
			for rows.Next() {
				var name string
				err = rows.Scan(&name)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("\u2717 %s\n", name)
			}
			err = rows.Err()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
