package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/go-jsonfile/jsonfile"
)

// The Goblimey scaffolder reads a specification file written in JSON describing
// a set of database tables.  It generates a web application server that implements
// the Create, Read, Update and Delete (CRUD) operations on those tables.  The
// idea is based on the Ruby-on-Rails scaffold generator.
//
// The examples directory contains example JSON specification files.
//
// Run the scaffolder like so:
//    scaffolder	// uses spec file scaffold.json
// or
//    scaffolder <json file>

type Field struct {
	Name               string   `json:"name"`
	Type               string   `json: "type"`
	ExcludeFromDisplay bool     `json: "excludeFromDisplay"`
	Mandatory          bool     `json: "mandatory"`
	TestValues         []string `json: "testValues"`
	GoType             string
	NameWithUpperFirst string
	NameWithLowerFirst string
	NameAllLower       string
	LastItem           bool
}

func (f Field) String() string {
	testValues := ""
	for _, s := range f.TestValues {
		if f.Type == "string" {
			testValues += fmt.Sprintf("\"%s\",", s)
		} else {
			testValues += s
		}
	}

	status := "optional"
	if f.Mandatory {
		status = "mandatory"
	}
	return fmt.Sprintf("{Name=%s,Type=%s,GoType=%s, ExcludeFromDisplay=%v,%s,TestValues=%s,NameWithLowerFirst=%s,NameWithUpperFirst=%s,NameAllLower=%s,LastItem=%v}",
		f.Name, f.Type, f.GoType, f.ExcludeFromDisplay, status, testValues,
		f.NameWithLowerFirst, f.NameWithUpperFirst, f.NameAllLower, f.LastItem)
}

type Resource struct {
	Name                      string `json:"name"`
	PluralName                string `json:"plural"`
	TableName                 string `json:"tableName"`
	NameWithUpperFirst        string
	NameWithLowerFirst        string
	NameAllLower              string
	PluralNameWithUpperFirst  string
	PluralNameWithLowerFirst  string
	ProjectName               string // copied from the name field of the spec record
	ProjectNameWithUpperFirst string
	Imports                   string
	SourceBase                string // copied from the spec record
	DB                        string // copied from the spec record
	DBURL                     string // copied from the spec record
	Fields                    []Field
}

func (r Resource) String() string {
	var fields string
	for _, f := range r.Fields {
		fields += f.String() + "\n"
	}
	return fmt.Sprintf("{Name=%s,PluralName=%s,TableName=%s,NameWithLowerFirst=%s,NameWithUpperFirst=%s,PluralNameWithLowerFirst=%s,PluralNameWithUpperFirst=%s,NameAllLower=%s,ProjectName=%s,imports=%s,DB=%s,DBURL=%s,fields=%s}",
		r.Name, r.PluralName, r.TableName,
		r.NameWithLowerFirst, r.NameWithUpperFirst,
		r.PluralNameWithLowerFirst, r.PluralNameWithUpperFirst, r.NameAllLower,
		r.ProjectName, r.Imports, r.DB, r.DBURL, fields)
}

type Spec struct {
	Name               string `json:"name"`
	SourceBase         string `json:"sourceBase"`
	DB                 string `json:"db"`
	DBUser             string `json:"dbuser"`
	DBPassword         string `json:"dbpassword"`
	DBServer           string `json:dbserver`
	DBPort             string `json:dbport`
	ORM                string `json:orm`
	DBURL              string
	CurrentDir         string
	NameWithUpperFirst string
	NameWithLowerFirst string
	NameAllUpper       string
	Imports            string
	Resources          []Resource
}

func (s Spec) String() string {
	var resources string
	for _, r := range s.Resources {
		resources += r.String() + "\n"
	}
	return fmt.Sprintf("{name=%s sourceBase=%s db=%s dbserver=%s dbport=s dbuser=%s dbpassword=%s dburl=s %d resources={%s}}",
		s.Name, s.SourceBase, s.DB, s.DBServer, s.DBPort, s.DBUser, s.DBPassword,
		s.DBURL, len(s.Resources), resources)
}

var templateMap map[string]*template.Template

var verbose bool
var overwriteMode bool
var templateDir string
var workspaceDir string

func init() {
	const (
		defaultVerbose = false
		usage          = "enable verbose logging"
	)
	flag.BoolVar(&verbose, "verbose", defaultVerbose, usage)
	flag.BoolVar(&verbose, "v", defaultVerbose, usage+" (shorthand)")

	flag.BoolVar(&overwriteMode, "overwrite", false, "overwrite all files, not just the generated directory")
	flag.StringVar(&templateDir, "templatedir", "", "the directory containing the scaffold templates (normally this is not specified and built in templates are used)")
	flag.StringVar(&workspaceDir, "workspace", "", "the Go workspace directory")

	templateMap = make(map[string]*template.Template)
}

func main() {
	log.SetPrefix("main() ")

	flag.Parse()

	// Find the scaffold spec.  By default it's "scaffold.json" but it can be
	// specified by the first (and only) command line argument.

	var specFile string
	if len(flag.Args()) >= 1 {
		specFile = flag.Args()[0]
	} else {
		specFile = "scaffold.json"
	}

	// Check that file exists and can be read
	jsonFile, err := os.Open(specFile)
	if err != nil {
		log.Printf("cannot open JSON specification file %s - %s", specFile,
			err.Error())
		os.Exit(-1)
	}
	jsonFile.Close()

	var spec Spec

	jsonfile.ReadJSONFromFile(specFile, &spec)
	if err != nil {
		log.Printf("cannot read JSON from specification file %s - %s", specFile, err.Error())
		os.Exit(-1)
	}

	if verbose {
		log.Printf("specification\n%s", spec.String())
	}

	data, err := json.MarshalIndent(&spec, "", "    ")
	if err != nil {
		log.Printf("internal error - cannot convert specification structure back to JSON - %s",
			err.Error())
		os.Exit(-1)
	}
	if verbose {
		log.Printf("initial specification\n%s\n", data)
	}

	// If the templateDir is not specified, produce templates from the built-in
	// prototypes.  Otherwise produce templates using the files in the templateDir
	// directory as prototypes.  The second choice is intended for use only during
	// development of the scaffolder.

	if templateDir == "" {
		createTemplates(true)
	} else {
		createTemplates(false)
	}

	// Enhance the data by setting the derived fields.

	if spec.DBPort == "" {
		if spec.DB == "mysql" {
			spec.DBPort = "3306"
		}
	}

	// "webuser:secret@tcp(localhost:3306)/animals"
	spec.DBURL = spec.DBUser + ":" + spec.DBPassword + "@tcp(" +
		spec.DBServer + ":" + spec.DBPort + ")/" + spec.Name

	// "animals" => "Animals"
	spec.NameWithUpperFirst = upperFirstRune(spec.Name)
	//"animals" => "ANIMALS"
	spec.NameAllUpper = strings.ToUpper(spec.Name)
	spec.CurrentDir, err = os.Getwd()
	if err != nil {
		log.Printf("internal error - cannot find the name of the current directory - %s",
			err.Error())
		os.Exit(-1)
	}

	for i, _ := range spec.Resources {
		// Set the last item flag in each field list.  For all but the last
		// field in a resource, the LastItem flag is false.  For the last field
		// it's true.  This helps the templates to construct things like lists
		// where the fields are separated by commas but the last field is not
		// followed by a comma, for example:
		// return MakeInitialisedPerson(source.ID(), source.Forename(), source.Surname())
		for j, _ := range spec.Resources[i].Fields {
			// Set LastItem true, then set it false on the next iteration.
			if j > 0 {
				spec.Resources[i].Fields[j].LastItem = true
				spec.Resources[i].Fields[j-1].LastItem = false
			}
		}

		spec.NameWithUpperFirst = upperFirstRune(spec.Name)
		spec.NameWithLowerFirst = lowerFirstRune(spec.Name)

		// These are supplied once in the spec, but each resource needs them,
		// so copy them into each resource record.
		spec.Resources[i].ProjectName = spec.Name
		spec.Resources[i].SourceBase = spec.SourceBase
		// "animals" => "Animals"
		spec.Resources[i].ProjectNameWithUpperFirst =
			upperFirstRune(spec.Name)
		spec.Resources[i].DB = spec.DB
		spec.Resources[i].DBURL = spec.DBURL

		// "CatAndDog" => "catAndDog"
		spec.Resources[i].NameWithLowerFirst = lowerFirstRune(spec.Resources[i].Name)
		// "cat" => "Cat"
		spec.Resources[i].NameWithUpperFirst = upperFirstRune(spec.Resources[i].Name)

		// "CatAndDog" => "catanddog"
		spec.Resources[i].NameAllLower =
			strings.ToLower(spec.Resources[i].Name)

		if spec.Resources[i].PluralName == "" {
			// "cat" => "cats"
			spec.Resources[i].PluralName = spec.Resources[i].NameWithLowerFirst + "s"
		}

		if spec.Resources[i].PluralName == "" {
			// "cat" => "cats"
			spec.Resources[i].PluralName = spec.Resources[i].NameWithLowerFirst + "s"
		}

		// "CatAndDogs" => "catAndDogs"
		spec.Resources[i].PluralNameWithLowerFirst =
			lowerFirstRune(spec.Resources[i].PluralName)

		// "catAndDogs" => "CatAndDogs"
		spec.Resources[i].PluralNameWithUpperFirst =
			upperFirstRune(spec.Resources[i].PluralName)

		// The table name is the plural of the lowered resource name (eg "cats")
		// but the JSON can specify it (eg resource name is "mouse" and table
		// name is "mice".
		if spec.Resources[i].TableName == "" {
			spec.Resources[i].TableName = spec.Resources[i].PluralName
		}

		// Set the fields that are set from other fields.
		nextTestValue := 1

		for j, _ := range spec.Resources[i].Fields {

			// In the JSON, the types are "int", "uint", "float",or "bool".  In
			// the generated Go code use int64 for int, unit64 for uint and
			// float64 for float.  Other types are OK.
			if spec.Resources[i].Fields[j].Type == "int" {
				spec.Resources[i].Fields[j].GoType = "int64"
			} else if spec.Resources[i].Fields[j].Type == "uint" {
				spec.Resources[i].Fields[j].GoType = "uint64"
			} else if spec.Resources[i].Fields[j].Type == "float" {
				spec.Resources[i].Fields[j].GoType = "float64"
			} else {
				spec.Resources[i].Fields[j].GoType =
					spec.Resources[i].Fields[j].Type
			}

			spec.Resources[i].Fields[j].NameWithUpperFirst =
				upperFirstRune(spec.Resources[i].Fields[j].Name)
			spec.Resources[i].Fields[j].NameWithLowerFirst =
				lowerFirstRune(spec.Resources[i].Fields[j].Name)
			spec.Resources[i].Fields[j].NameAllLower =
				strings.ToLower(spec.Resources[i].Fields[j].Name)

			// The test values are optional.  We need two values for each
			// field, because some tests create two objects.  If only one
			// value is supplied, then use that and create the second. If
			// none are supplied, then create both.  To create all values,
			// use a sequence such as:
			// {"s1", "s2}, {"s3", "s4"}, {"s5", "s6"} for three string types,
			// or {"1.1", "2.1"}, {"s3", "s4"} for a float type followed by a
			// string type.
			//
			// For booleans, generate {true, false}, {true, false} ....

			CreateFirstTestValue := false
			CreateSecondTestValue := false
			if spec.Resources[i].Fields[j].TestValues == nil {
				spec.Resources[i].Fields[j].TestValues = make([]string, 2)
				CreateFirstTestValue = true
				CreateSecondTestValue = true
			} else {
				if len(spec.Resources[i].Fields[j].TestValues) == 0 {
					CreateFirstTestValue = true
					CreateSecondTestValue = true
				} else if len(spec.Resources[i].Fields[j].TestValues) == 1 {
					// Got the first value, need the second.
					CreateSecondTestValue = true
				} else {
					// Got both values already
				}
			}

			switch spec.Resources[i].Fields[j].Type {
			case "string":
				if CreateFirstTestValue {
					spec.Resources[i].Fields[j].TestValues[0] =
						fmt.Sprintf("s%d", nextTestValue)
				}
				if CreateSecondTestValue {
					spec.Resources[i].Fields[j].TestValues[1] =
						fmt.Sprintf("s%d", nextTestValue+1)
				}
			case "int":
				if CreateFirstTestValue {
					spec.Resources[i].Fields[j].TestValues[0] =
						fmt.Sprintf("%d", nextTestValue)
				}
				if CreateSecondTestValue {
					spec.Resources[i].Fields[j].TestValues[1] =
						fmt.Sprintf("%d", nextTestValue+1)
				}
			case "uint":
				if CreateFirstTestValue {
					spec.Resources[i].Fields[j].TestValues[0] =
						fmt.Sprintf("%d", nextTestValue)
				}
				if CreateSecondTestValue {
					spec.Resources[i].Fields[j].TestValues[1] =
						fmt.Sprintf("%d", nextTestValue+1)
				}
			case "float":
				if CreateFirstTestValue {
					spec.Resources[i].Fields[j].TestValues[0] =
						fmt.Sprintf("%d.1", nextTestValue)
				}
				if CreateSecondTestValue {
					spec.Resources[i].Fields[j].TestValues[1] =
						fmt.Sprintf("%d.1", nextTestValue+1)
				}
			case "bool":
				if CreateFirstTestValue {
					spec.Resources[i].Fields[j].TestValues[0] = "true"
				}
				if CreateSecondTestValue {
					spec.Resources[i].Fields[j].TestValues[1] = "false"
				}
			default:
				log.Printf("cannot handle type %s ", spec.Resources[i].Fields[j].Type)
				os.Exit(-1)
			}

			nextTestValue += 2 // 1, 3, 5 ...
		}
	}

	data, err = json.MarshalIndent(&spec, "", "    ")
	if err != nil {
		log.Printf("internal error - cannot convert the spec structure back to JSON after enhancement - %s",
			err.Error())
		os.Exit(-1)
	}

	if verbose {
		log.Printf("enhanced spec:\n%s\n", data)
	}

	// Build the project from the templates and the JSON spec.

	// By default, the projectDir is the current directory.  If the workspace
	// directory is specified, the project directory is specified by the
	// sourceBase, something like {workspaceDir}/src/github.com/goblimey/animals.

	projectDir := "."
	if strings.TrimSpace(workspaceDir) != "" {
		projectDir = workspaceDir + "/src/" + spec.SourceBase
		err = os.Chdir(projectDir)
		if err != nil {
			log.Printf("cannot change directory to project directory %s - %s",
				err.Error(), workspaceDir)
			os.Exit(-1)
		}
	}

	// install.sh script with permission u+rwx
	templateName := "script.install.sh.template"
	targetName := "install.sh"
	createFileFromTemplateAndSpec(projectDir, targetName, templateName, spec,
		overwriteMode)

	var permisssions os.FileMode = 0700
	os.Chmod(projectDir+"/"+targetName, permisssions)

	// test.sh script with permission u+rwx
	templateName = "script.test.sh.template"
	targetName = "test.sh"
	createFileFromTemplateAndSpec(projectDir, targetName, templateName, spec,
		overwriteMode)
	os.Chmod(projectDir+"/"+targetName, permisssions)

	// Windoze batch files
	templateName = "script.install.bat.template"
	targetName = "install.bat"
	createFileFromTemplateAndSpec(projectDir, targetName, templateName, spec,
		overwriteMode)

	templateName = "script.test.bat.template"
	targetName = "test.bat"
	createFileFromTemplateAndSpec(projectDir, targetName, templateName, spec,
		overwriteMode)

	// Build the main program.
	templateName = "main.go.template"
	targetName = spec.NameWithLowerFirst + ".go"

	spec.Imports = `
	import (
		"flag"
		"fmt"
		"log"
		"net/http"
		"os"
		"regexp"
		"strconv"
		"strings"
		restful "github.com/emicklei/go-restful"
		retrofitTemplate "` + spec.SourceBase +
		"/generated/crud/retrofit/template" + `"
		"` + spec.SourceBase + "/generated/crud/services" + `"
		"` + spec.SourceBase + "/generated/crud/utilities" + `"
		`

	for _, resource := range spec.Resources {
		// personForms "github.com/goblimey/films/generated/crud/forms/people"
		spec.Imports += resource.NameWithLowerFirst + `Forms "` +
			spec.SourceBase + "/generated/crud/forms/" + resource.NameWithLowerFirst + `"
		`
		// personController "github.com/goblimey/films/generated/crud/controllers/person"
		spec.Imports += resource.NameWithLowerFirst + `Controller "` +
			spec.SourceBase + "/generated/crud/controllers/" +
			resource.NameWithLowerFirst + `"
		`
		// personRepository "github.com/goblimey/films/generated/crud/repositories/person/gorpmysql"
		spec.Imports += resource.NameWithLowerFirst + `Repository "` +
			spec.SourceBase + "/generated/crud/repositories/" +
			resource.NameWithLowerFirst + `/gorpmysql"
		`
	}

	spec.Imports += `
	)`
	createFileFromTemplateAndSpec(projectDir, targetName, templateName, spec,
		overwriteMode)

	// Build the static views.  It's assumed that the user may want to edit
	// these and add their own stuff, so they are not overwritten.

	// views/stylesheets/scaffold.css - the static stylesheet.
	stylesheetDir := projectDir + "/views/stylesheets"
	targetName = "scaffold.css"
	templateName = "view.stylesheets.scaffold.css.template"
	createFileFromTemplateAndSpec(stylesheetDir, targetName, templateName, spec,
		overwriteMode)

	// views/html/index.html - the static application home page.
	htmlDir := projectDir + "/views/html"
	targetName = "index.html"
	templateName = "view.index.ghtml.template"
	createFileFromTemplateAndSpec(htmlDir, targetName, templateName, spec,
		overwriteMode)

	// views/html/error.html - the static error page.
	targetName = "error.html"
	templateName = "view.error.html.template"
	createFileFromTemplateAndSpec(htmlDir, targetName, templateName, spec,
		overwriteMode)

	// These files are always overwritten.

	// views/generated/crud/templates/_base.ghtml - the prototype for all generated
	// pages.
	generatedDir := projectDir + "/views/generated/crud/templates"
	targetName = "_base.ghtml"
	templateName = "view.base.ghtml.template"
	createFileFromTemplateAndSpec(generatedDir, targetName, templateName, spec,
		true)

	// Generate the sql scripts.
	sqlDir := projectDir + "/generated/sql"
	templateName = "sql.create.db.template"
	targetName = "create.db.sql"
	createFileFromTemplateAndSpec(sqlDir, targetName, templateName, spec, true)

	// Generate the utilities.

	crudBase := projectDir + "/generated/crud"
	utilitiesDir := crudBase + "/utilities"
	templateName = "utilities.go.template"
	targetName = "utilities.go"

	spec.Imports = `
		import (
			"fmt"
			"html/template"
			"log"
			"net/http"
			"strings"
			restful "github.com/emicklei/go-restful"
			retrofitTemplate "` + spec.SourceBase +
		"/generated/crud/retrofit/template" + `"
			)`
	createFileFromTemplateAndSpec(utilitiesDir, targetName, templateName, spec,
		true)

	retrofitDir := crudBase + "/retrofit/template"
	templateName = "retrofit.template.go.template"
	targetName = "template.go"
	createFileFromTemplateAndSpec(retrofitDir, targetName, templateName,
		spec, true)

	// Generate the services object.

	// Interface.
	servicesDir := crudBase + "/services"
	templateName = "services.go.template"
	targetName = "services.go"

	spec.Imports = `
		import (
			retrofitTemplate "` + spec.SourceBase +
		"/generated/crud/retrofit/template" + `"
			`
	for _, resource := range spec.Resources {
		// personForms "github.com/goblimey/films/generated/crud/forms/people"
		spec.Imports += resource.NameWithLowerFirst + `Forms "` +
			spec.SourceBase + "/generated/crud/forms/" +
			resource.NameWithLowerFirst + `"
			`
		// "github.com/goblimey/films/generated/crud/models/person"
		spec.Imports += `"` + spec.SourceBase + "/generated/crud/models/" +
			resource.NameWithLowerFirst + `"
			`
		// peopleRepo "github.com/goblimey/films/generated/crud/repositories/people"
		spec.Imports += resource.NameWithLowerFirst + `Repo "` + spec.SourceBase +
			"/generated/crud/repositories/" + resource.NameWithLowerFirst + `"
			`
	}
	spec.Imports += ")"

	createFileFromTemplateAndSpec(servicesDir, targetName, templateName, spec,
		true)

	// Concrete type.
	templateName = "services.concrete.go.template"
	targetName = "concrete_services.go"

	spec.Imports = `
		import (
			retrofitTemplate "` + spec.SourceBase +
		"/generated/crud/retrofit/template" + `"
			`
	for _, resource := range spec.Resources {
		// personForms "github.com/goblimey/films/generated/crud/forms/people"
		spec.Imports += resource.NameWithLowerFirst + `Forms "` +
			spec.SourceBase + "/generated/crud/forms/" +
			resource.NameWithLowerFirst + `"
			`
		// "github.com/goblimey/films/generated/crud/models/person"
		spec.Imports += `"` + spec.SourceBase +
			"/generated/crud/models/" + resource.NameWithLowerFirst + `"
			`
		// gorpPerson "github.com/goblimey/films/generated/crud/models/person/gorp"
		spec.Imports += "gorp" + resource.NameWithUpperFirst + ` "` +
			spec.SourceBase + "/generated/crud/models/" +
			resource.NameWithLowerFirst + `/gorp"
			`
		// peopleRepo "github.com/goblimey/films/generated/crud/repositories/people"
		spec.Imports += resource.NameWithLowerFirst + `Repo "` + spec.SourceBase +
			"/generated/crud/repositories/" + resource.NameWithLowerFirst + `"
			`
	}
	spec.Imports += ")"
	createFileFromTemplateAndSpec(servicesDir, targetName, templateName, spec,
		true)

	// Generate the  models.

	for _, resource := range spec.Resources {

		// Generate the interface and concrete objects for the model.

		modelDir := crudBase + "/models/" + resource.NameAllLower
		targetName = resource.NameAllLower + ".go"
		templateName = "model.interface.go.template"
		createFileFromTemplateAndResource(modelDir, targetName, templateName,
			resource)

		// concrete model object
		targetName = "concrete_" + resource.NameAllLower + ".go"
		templateName = "model.concrete.go.template"
		createFileFromTemplateAndResource(modelDir, targetName, templateName,
			resource)

		// test for concrete model object
		targetName = "concrete_" + resource.NameAllLower + "_test.go"
		templateName = "model.concrete.test.go.template"
		createFileFromTemplateAndResource(modelDir, targetName, templateName,
			resource)

		// concrete model object using gorp to access the database
		modelDir += "/gorp"
		targetName = "concrete_" + resource.NameAllLower + ".go"
		templateName = "gorp.concrete.go.template"

		resource.Imports = `
			import (
				"errors"
				"fmt"
				"strings"
				"` +
			spec.SourceBase + "/generated/crud/models/" +
			resource.NameWithLowerFirst + `"
			)`

		createFileFromTemplateAndResource(modelDir, targetName, templateName,
			resource)

		// The test for the gorp version of the model is the same test as for the
		// concrete model object, but in the appropriate directory.
		targetName = "concrete_" + resource.NameAllLower + "_test.go"
		templateName = "model.concrete.test.go.template"
		createFileFromTemplateAndResource(modelDir, targetName, templateName,
			resource)

		// Generate the repository.

		// interface
		interfaceDir := crudBase + "/repositories/" + resource.NameAllLower
		targetName = "repository.go"
		templateName = "repository.interface.go.template"

		resource.Imports = `
			import ("` + spec.SourceBase + "/generated/crud/models/" +
			resource.NameAllLower + `")`

		createFileFromTemplateAndResource(interfaceDir, targetName, templateName,
			resource)

		// concrete repository using gorp to access the mysql database
		interfaceDir += "/gorpmysql"
		targetName = "concrete_repository.go"
		templateName = "repository.concrete.gorp.go.template"

		resource.Imports = `
			import (
				"database/sql"
				"errors"
				"fmt"
				"log"
				"strconv"
				"strings"
				// This import must be present to satisfy a dependency in the GORP library.
				_ "github.com/go-sql-driver/mysql"
				gorp "gopkg.in/gorp.v1"
				` +
			resource.NameWithLowerFirst + ` "` +
			spec.SourceBase + "/generated/crud/models/" +
			resource.NameAllLower + `"
				` +
			"gorp" + resource.NameWithUpperFirst + ` "` +
			spec.SourceBase + "/generated/crud/models/" + resource.NameAllLower +
			`/gorp"
				` +
			resource.NameWithLowerFirst + "Repo " + `"` +
			spec.SourceBase + "/generated/crud/repositories/" +
			resource.NameWithLowerFirst + `"
			)`

		createFileFromTemplateAndResource(interfaceDir, targetName, templateName,
			resource)

		// Unit test
		targetName = "concrete_repository_test.go"
		templateName = "repository.concrete.gorp.test.go.template"

		resource.Imports = `
			import (
				"fmt"
				"log"
				"os"
				"strconv"
				"testing"
				gorp` + resource.NameWithUpperFirst +
			` "` +
			spec.SourceBase + "/generated/crud/models/" +
			resource.NameAllLower + `/gorp"
				"` + spec.SourceBase + "/generated/crud/repositories/" +
			resource.NameWithLowerFirst + `"
			)`

		createFileFromTemplateAndResource(interfaceDir, targetName, templateName,
			resource)

		// generated the forms.

		// interface for single object form - single_object_form.go

		formsDir := crudBase + "/forms/" + resource.NameAllLower
		targetName = "single_item_form.go"
		templateName = "form.single.item.go.template"

		// import ("github.com/goblimey/films/models/person")
		resource.Imports = `
			import (
				"` +
			spec.SourceBase + "/generated/crud/models/" +
			resource.NameAllLower + `"
			)`

		createFileFromTemplateAndResource(formsDir, targetName, templateName,
			resource)

		// interface for list form - list_form.go

		targetName = "list_form.go"
		templateName = "form.list.go.template"
		// import ("github.com/goblimey/films/generated/crud/models/person")
		resource.Imports = `import ("` + spec.SourceBase +
			"/generated/crud/models/" + resource.NameWithLowerFirst + `")`

		createFileFromTemplateAndResource(formsDir, targetName, templateName,
			resource)

		// concrete structure for single item form concrete_person_form.go
		targetName = "concrete_single_item_form.go"
		templateName = "form.concrete.single.item.go.template"

		resource.Imports = `
			import (
				"fmt"
				"strings"
				"` + spec.SourceBase + `/generated/crud/utilities"
				"` + spec.SourceBase + "/generated/crud/models/" +
			resource.NameAllLower + `"
			)`

		createFileFromTemplateAndResource(formsDir, targetName, templateName,
			resource)

		// test for concrete single item form - concrete_single_item_form_test.go
		targetName = "concrete_single_item_form_test.go"
		templateName = "form.concrete.single.item.test.go.template"

		resource.Imports = `
			import (
				"testing"
				` + resource.NameAllLower + `Model "` + spec.SourceBase +
			"/generated/crud/models/" + resource.NameAllLower + `"
		)`

		createFileFromTemplateAndResource(formsDir, targetName, templateName,
			resource)

		resource.Imports = `
			import (
				"testing"
				"` + spec.SourceBase + "/generated/crud/models/" +
			resource.NameAllLower + `"
				"` + spec.SourceBase + "/generated/crud/repositories/" +
			resource.PluralNameWithLowerFirst + `"
			)`

		// concrete structure for list form concrete_list_form.go
		targetName = "concrete_list_form.go"
		templateName = "form.concrete.list.go.template"
		// import ("github.com/goblimey/films/generated/crud/models/person")
		resource.Imports = `import ("` + spec.SourceBase +
			`/generated/crud/models/` + resource.NameAllLower + `")`
		createFileFromTemplateAndResource(formsDir, targetName, templateName,
			resource)

		// Generate the controller.
		controllerDir := crudBase + "/controllers/" + resource.NameAllLower
		targetName = "controller.go"
		templateName = "controller.go.template"

		resource.Imports = `
			import (
				"fmt"
				"log"
				restful "github.com/emicklei/go-restful"
				"` + spec.SourceBase + "/generated/crud/utilities" + `"
				` + resource.NameWithLowerFirst + `Forms "` + spec.SourceBase +
			"/generated/crud/forms/" + resource.NameWithLowerFirst + `"
				"` + spec.SourceBase + "/generated/crud/services" + `"
			)`

		createFileFromTemplateAndResource(controllerDir, targetName, templateName,
			resource)

		// Controller test.
		targetName = "controller_test.go"
		templateName = "controller.test.go.template"

		resource.Imports = `
			import (
				"errors"
				"fmt"
				"log"
				"net/http"
				"net/url"
				"strings"
				"testing"
				restful "github.com/emicklei/go-restful"
				"github.com/petergtz/pegomock"
				retrofitTemplate "` + spec.SourceBase +
			"/generated/crud/retrofit/template" + `"
				"` + spec.SourceBase + "/generated/crud/services" + `"
				mocks "` + spec.SourceBase + "/generated/crud/mocks/pegomock" + `"
				mock` + resource.NameWithUpperFirst + ` "` +
			spec.SourceBase + "/generated/crud/mocks/pegomock/" +
			resource.NameWithLowerFirst + `"
				` +
			// personForms "github.com/goblimey/films/generated/crud/forms/person"
			resource.NameWithLowerFirst + `Forms "` + spec.SourceBase +
			"/generated/crud/forms/" + resource.NameWithLowerFirst + `"
						` +
			// person "github.com/goblimey/films/generated/crud/models/person"
			resource.NameWithLowerFirst + ` "` + spec.SourceBase +
			"/generated/crud/models/" + resource.NameWithLowerFirst + `"
			)`

		createFileFromTemplateAndResource(controllerDir, targetName, templateName,
			resource)

		// Build the views for each model.

		// views/generated/crud/templates/index.ghtml - html template for the index
		// page for the model.
		ghtmlDir := projectDir + "/views/generated/crud/templates/" +
			resource.NameAllLower
		targetName = "index.ghtml"
		templateName = "view.resource.index.ghtml.template"
		createFileFromTemplateAndResource(ghtmlDir, targetName, templateName,
			resource)

		// views/generated/crud/templates/create.ghtml - html template for the
		// create page for the model.
		targetName = "create.ghtml"
		templateName = "view.resource.create.ghtml.template"
		createFileFromTemplateAndResource(ghtmlDir, targetName, templateName,
			resource)

		// views/generated/crud/templates/edit.ghtml - html template for the edit
		// page for the model.
		targetName = "edit.ghtml"
		templateName = "view.resource.edit.ghtml.template"
		createFileFromTemplateAndResource(ghtmlDir, targetName, templateName,
			resource)

		// views/generated/crud/templates/edit.ghtml - html template for the show
		// page for each model.
		targetName = "show.ghtml"
		templateName = "view.resource.show.ghtml.template"
		createFileFromTemplateAndResource(ghtmlDir, targetName, templateName,
			resource)
	}
}

func createFileFromTemplateAndSpec(targetDir string, targetName string,
	templateName string, spec Spec, overwrite bool) {

	log.SetPrefix("createFileFromTemplateAndSpec ")

	file, err := createAndOpenFile(targetDir, targetName, overwrite)
	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}

	// Special case, only happens if overwrite is true and file exists - nothing
	// to do.

	if file == nil {
		return
	}

	defer file.Close()

	err = templateMap[templateName].Execute(file, spec)
	if err != nil {
		log.Printf("error creating file %s from template %s - %s ",
			targetDir+"/"+targetName, templateName, err.Error())
		os.Exit(-1)
	}
}

func createFileFromTemplateAndResource(targetDir string, targetName string,
	templateName string, resource Resource) {

	log.SetPrefix("createFileFromTemplateAndResource ")

	targetPathName := targetDir + "/" + targetName
	if verbose {
		log.Printf("creating file %s from template %s", targetPathName,
			templateName)
	}
	conn, err := createAndOpenFile(targetDir, targetName, true)
	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}
	defer conn.Close()

	err = templateMap[templateName].Execute(conn, resource)
	if err != nil {
		log.Printf("error creating file %s from template %s - %s ",
			targetDir+"/"+targetName, templateName, err.Error())
		os.Exit(-1)
	}
}

// CreateAndOpenfile creates a file if it doesn't exist, opens it and returns
// a file descriptor, or any error.  An existing file is only overwritten
// if overwrite is true.
func createAndOpenFile(targetDir string, targetName string,
	overwrite bool) (*os.File, error) {

	log.SetPrefix("createAndOpenFile ")

	if verbose {
		log.Printf("%s/%s verbose %v", targetDir, targetName, verbose)
	}

	// Ensure that the target directory exists.
	err := os.MkdirAll(targetDir, 0777)
	if err != nil {
		log.Printf("cannot create target directory %s - %s ", targetDir, err.Error())
		return nil, err
	}

	// If the file already exists, do not write to it except in overwrite
	// mode.

	if !overwrite {
		dir, err := os.Open(targetDir)
		if err != nil {
			log.Printf("cannot open target directory %s - %s ",
				targetDir, err.Error())
		}

		defer dir.Close()

		// Get the contents of the target directory
		fileInfoList, err := dir.Readdir(0)
		if err != nil {
			log.Printf("cannot scan target directory %s - %s ",
				targetDir, err.Error())
			return nil, err
		}

		// Scan the target directory to see if the file already exists.
		for _, fileInfo := range fileInfoList {
			if fileInfo.Name() == targetName {
				if verbose {
					log.Printf("file %s/%s already exists and overwrite mode is off.",
						targetDir, targetName)
				}
				return nil, nil
			}
		}
	}

	targetPathName := targetDir + "/" + targetName

	if verbose {
		log.Printf("Creating file %s - overwrite %v.",
			targetPathName, overwrite)
	}
	file, err := os.Create(targetPathName)
	if err != nil {
		log.Println("cannot create target file %s for writing - %s ",
			targetPathName, err.Error())
		return nil, err
	}

	return file, nil
}

// lowerFirstRune takes a string and ensures that the first rune is lower case.
// From https://play.golang.org/p/D8cYDgfZr8 via
// https://groups.google.com/forum/#!topic/golang-nuts/WfpmVDQFecU
func lowerFirstRune(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToLower(r)) + s[n:]
}

// upperFirstRune takes a string and ensures that the first rune is upper case.
// For origins, see lowerFirstRune.
func upperFirstRune(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[n:]
}
