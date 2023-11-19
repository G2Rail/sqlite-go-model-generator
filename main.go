package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/dave/jennifer/jen"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var commonInitialisms = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SSH":   true,
	"TLS":   true,
	"TTL":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"NTP":   true,
	"DB":    true,
}

var flagDbFile = flag.String("db-file", "./example.db", "path to the DB")
var flagOut = flag.String("out", "./gen", "output file for generated files")
var flagIgnoreColumns = flag.String("skip", "rowid,_rowid_,_rid,rid", "list of columns to be excluded from struct generation")
var flagGenJson = flag.Bool("json", true, "generate JSON annotation")
var flagGenDb = flag.Bool("db", true, "generate DB annotation")
var flagGenGorm = flag.Bool("gorm", true, "generate GORM annotation")
var flagPkgName = flag.String("pkg", "def", "specify package name")

var ignoreColumns []string

var intToWordMap = []string{
	"zero",
	"one",
	"two",
	"three",
	"four",
	"five",
	"six",
	"seven",
	"eight",
	"nine",
}

func main() {
	flag.Parse()

	db, err := sql.Open("sqlite3", *flagDbFile)

	ignoreColumns = strings.Split(strings.ToLower(*flagIgnoreColumns), ",")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	tableNames := getTableNames(db)
	outPath := filepath.Clean(*flagOut)
	errOs := os.MkdirAll(outPath, 0770)

	if errOs != nil {
		log.Fatal(errOs)
	}

	os.Create(outPath)
	c := 0
	fmt.Printf("Generating code for the following tables (%d)\n", len(tableNames))
	for _, tableName := range tableNames {
		c++
		fmt.Printf("[%d] %s\n", c, tableName)

		file := scanTableStructure(db, tableName, outPath, *flagPkgName)
		structureName := formatFieldName(tableName)
		fileName := filepath.Join(outPath, fmt.Sprintf("%s.go", structureName))
		err = file.Save(fileName)
	}

	if err != nil {
		panic(err)
	}
}

func scanTableStructure(db *sql.DB, tableName string, outPath string, packageName string) *jen.File {
	file := jen.NewFilePathName(outPath, packageName)
	structureName := formatFieldName(tableName)
	file.Comment(fmt.Sprintf("// %s represent database table (%s)", structureName, tableName))
	file.Type().Id(structureName).Struct(
		*generateTableFields(db, tableName)...,
	)
	file.Comment(fmt.Sprintf("// TableName represent the database table name of %s", structureName))
	file.Func().Parens(jen.Id(structureName)).Id("TableName").Params().String().Block(
		jen.Return(jen.Lit(tableName)),
	)
	return file
}

func getTableNames(db *sql.DB) []string {
	var tableNames []string
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var name string

		err = rows.Scan(&name)
		if err != nil {
			log.Fatal(err)
		}
		tableNames = append(tableNames, name)
	}
	return tableNames
}

func generateTableFields(db *sql.DB, tableName string) *[]jen.Code {
	var fields []jen.Code

	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s);", tableName))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	//fmt.Println(getTableNames(db))
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull string
		var dfltValue sql.NullString
		var pk string
		var ignore bool

		err = rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			log.Fatal(err)
		}

		ignore = isIgnoreField(name)

		//fmt.Println(cid, name, ctype, notnull, dfltValue, pk)
		name2 := formatFieldName(name)
		if ignore {
			name2 = `// ` + name2
		}
		field := jen.Id(name2)
		setFieldType(field, ctype)
		setFieldTags(field, name)
		fields = append(fields, field)
	}

	return &fields
}

func setFieldTags(field *jen.Statement, name string) {
	m := map[string]string{}

	if *flagGenDb {
		m["db"] = name
	}

	if *flagGenGorm {
		m["gorm"] = fmt.Sprintf("column:%s", name)
	}

	if *flagGenJson {
		m["json"] = name
	}

	field.Tag(m)
}

func setFieldType(field *jen.Statement, ctype string) {
	dbType := strings.ToUpper(strings.Split(ctype, "(")[0])
	switch dbType {
	case "VARCHAR", "TEXT":
		field.String()
	case "BOOL", "BOOLEAN":
		field.Bool()
	case "TINYINT", "SMALLINT":
		field.Int32()
	case "INTEGER", "INT", "INT2", "MEDIUMINT", "BIGINT", "UNSIGNED BIG INT", "INT8":
		field.Int64()
	case "REAL", "DOUBLE", "DOUBLE PRECISION", "FLOAT":
		field.Float32()
	case "NUMERIC", "DECIMAL", "DECIMAL(10,5)":
		field.Float64()
	default:
		field.String()
	}
}

func formatFieldName(s string) string {
	runes := []rune(s)
	for len(runes) > 0 && !unicode.IsLetter(runes[0]) && !unicode.IsDigit(runes[0]) {
		runes = runes[1:]
	}
	if len(runes) == 0 {
		return "_"
	}

	s = stringifyFirstChar(string(runes))
	name := lintFieldName(s)
	runes = []rune(name)
	for i, c := range runes {
		ok := unicode.IsLetter(c) || unicode.IsDigit(c)
		if i == 0 {
			ok = unicode.IsLetter(c)
		}
		if !ok {
			runes[i] = '_'
		}
	}
	s = string(runes)
	s = strings.Trim(s, "_")
	if len(s) == 0 {
		return "_"
	}
	return s
}

func lintFieldName(name string) string {
	// Fast path for simple cases: "_" and all lowercase.
	if name == "_" {
		return name
	}

	allLower := true
	for _, r := range name {
		if !unicode.IsLower(r) {
			allLower = false
			break
		}
	}

	if allLower {
		runes := []rune(name)
		if u := strings.ToUpper(name); commonInitialisms[u] {
			copy(runes[0:], []rune(u))
		} else {
			runes[0] = unicode.ToUpper(runes[0])
		}
		return string(runes)
	}

	allUpperWithUnderscore := true
	for _, r := range name {
		if !unicode.IsUpper(r) && r != '_' {
			allUpperWithUnderscore = false
			break
		}
	}

	if allUpperWithUnderscore {
		name = strings.ToLower(name)
	}

	// Split camelCase at any lower->upper transition, and split on underscores.
	// Check each word for common initialisms.
	runes := []rune(name)
	w, i := 0, 0 // index of start of word, scan
	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word

		if i+1 == len(runes) {
			eow = true
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}

			// Leave at most one underscore if the underscore is between two digits
			if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
				n--
			}

			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i) is a word.
		word := string(runes[w:i])
		if u := strings.ToUpper(word); commonInitialisms[u] {
			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))

		} else if strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		}
		w = i
	}
	return string(runes)
}

func stringifyFirstChar(str string) string {
	first := str[:1]

	i, err := strconv.ParseInt(first, 10, 8)

	if err != nil {
		return str
	}

	return intToWordMap[i] + "_" + str[1:]
}

func isIgnoreField(fieldName string) bool {
	lowerFieldName := strings.ToLower(fieldName)
	for _, v := range ignoreColumns {
		if v == lowerFieldName {
			return true
		}
	}
	return false
}
