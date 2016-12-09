#!/bin/sh

# Takes a set of templates stored in files and creates Go code that specifies the
# same templates inline.  Normally the scaffolder uses these inline versions and so
# the user doesn't need to specify where to find the tenplates when they run the 
# scaffolder.
#
# usage:
#    cd templates
#    ../template.sh >../src/github.com/goblimey/scaffolder/create_templates.go

echo 'package main'
echo
echo 'import ('
echo '	"io/ioutil"'
echo '	"strings"'
echo '	"text/template"'
echo '	"log"'
echo '	"os"'
echo ')'
echo
echo '// substituteGraves replaces each occurence of the sequence "%%GRAVE%%" with a'
echo '// single grave (backtick) rune.  In this source file, all templates are quoted in'
echo '// graves, but some templates contain graves, and a grave within a grave causes a'
echo '// syntax error.  The solution is to replace the graves in the template with'
echo '// "%%GRAVE%% and then pre-process the template before use.'
echo 'func substituteGraves(s string) string {'
echo '	return strings.Replace(s, "%%GRAVE%%", "\x60", -1)'
echo '}'
echo
echo '// createTemplateFromFile creates a template from a file.  The file is in the'
echo '// templates directory wherever the scaffolder is installed, and that is out of our'
echo '// control, so this should only be called when the "templatedir" command line'
echo '// argument is specified. '
echo 'func createTemplateFromFile(templateName string) *template.Template {'
echo '	log.SetPrefix("createTemplate() ")'
echo '	templateFile := templateDir + templateName'
echo '	buf, err := ioutil.ReadFile(templateFile)'
echo '	if err != nil {'
echo '		log.Printf("cannot open template file %s - %s ",'
echo '			templateFile, err.Error())'
echo '		os.Exit(-1)'
echo '	}'
echo '	tp := string(buf)'
echo '	tp = substituteGraves(tp)'
echo '	return template.Must(template.New(templateName).Parse(tp))'
echo '}'
echo
echo 'func createTemplates(useBuiltIn bool) {'
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
done
echo '}'
