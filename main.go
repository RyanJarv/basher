package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"

	"database/sql"

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
		cmds <- v
	}
	return true
}

func findPrograms(cmds chan string) {
	db, err := sql.Open("sqlite3", commanddb)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for cmd := range <-cmds {
		sqlStmt := fmt.Sprintf(`select p.name from packages as p join commands as c on p.pkgID = c.pkgID where c.command = "%s";`, cmd)
		rows, err = db.Query(sqlStmt)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var id int
			var name string
			err = rows.Scan(&id, &name)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(id, name)
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}
