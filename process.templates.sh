#!/bin/sh

# Takes a set of templates stored in files and creates Go code that specifies the
# same templates inline.  Normally the scaffolder uses these inline versions and so
# the user doesn't need to specify where to find the tenplates when they run the 
# scaffolder.
#
# usage:
#    ./process_templates.sh

process() {
    echo 'package main' >$1
    echo >>$1
    echo 'import (' >>$1
    echo '	"io/ioutil"' >>$1
    echo '	"strings"' >>$1
    echo '	"text/template"' >>$1
    echo '	"log"' >>$1
    echo '	"os"' >>$1
    echo ')' >>$1
    echo >>$1
    echo '// substituteGraves replaces each occurence of the sequence "%%GRAVE%%" with a' >>$1
    echo '// single grave (backtick) rune.  In this source file, all templates are quoted in' >>$1
    echo '// graves, but some templates contain graves, and a grave within a grave causes a' >>$1
    echo '// syntax error.  The solution is to replace the graves in the template with' >>$1
    echo '// "%%GRAVE%% and then pre-process the template before use.' >>$1
    echo 'func substituteGraves(s string) string {' >>$1
    echo '	return strings.Replace(s, "%%GRAVE%%", "\x60", -1)' >>$1
    echo '}' >>$1
    echo >>$1
    echo '// createTemplateFromFile creates a template from a file.  The file is in the' >>$1
    echo '// templates directory wherever the scaffolder is installed, and that is out of our' >>$1
    echo '// control, so this should only be called when the "templatedir" command line' >>$1
    echo '// argument is specified. ' >>$1
    echo 'func createTemplateFromFile(templateName string) *template.Template {' >>$1
    echo '	log.SetPrefix("createTemplate() ")' >>$1
    echo '	templateFile := templateDir + templateName' >>$1
    echo '	buf, err := ioutil.ReadFile(templateFile)' >>$1
    echo '	if err != nil {' >>$1
    echo '		log.Printf("cannot open template file %s - %s ",' >>$1
    echo '			templateFile, err.Error())' >>$1
    echo '		os.Exit(-1)' >>$1
    echo '	}' >>$1
    echo '	tp := string(buf)' >>$1
    echo '	tp = substituteGraves(tp)' >>$1
    echo '	return template.Must(template.New(templateName).Parse(tp))' >>$1
    echo '}' >>$1
    echo >>$1
    echo 'func createTemplates(useBuiltIn bool) {' >>$1
    first=1
    for file in *
    do
        echo
        if test $first -eq 1
        then
            echo 'templateName := "'$file'"'
        else
    	        echo 'templateName = "'$file'"'
        fi 
        first=0 
        echo '	if useBuiltIn {' 
        echo '		if verbose {' 
        echo '			log.Printf("creating template %s from builtin template", templateName)' 
        echo '		}' 
        echo '		templateText := `' 
        cat $file 
        echo '`' 
        echo '		templateText = substituteGraves(templateText)' 
        echo '		templateMap[templateName] =' 
        echo '			template.Must(template.New(templateName).Parse(templateText))' 
        echo '	} else {' 
        echo '		if verbose {' 
        echo '			log.Printf("creating template %s from file %s", templateName, templateDir+templateName)' 
        echo '		}' 
        echo '		templateMap[templateName] = createTemplateFromFile(templateName)' 
        echo '	}' 
    done >>$1
    echo '}' >>$1
}

cd $GOPATH/src/github.com/goblimey/scaffolder/templates

process "../create_templates.go"
