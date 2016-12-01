package main

import (
	"io/ioutil"
	"strings"
	"text/template"
	"log"
	"os"
)

// substituteGraves replaces each occurence of the sequence "%%GRAVE%%" with a
// single grave (backtick) rune.  In this source file, all templates are quoted in
// graves, but some templates contain graves, and a grave within a grave causes a
// syntax error.  The solution is to replace the graves in the template with
// "%%GRAVE%% and then pre-process the template before use.
func substituteGraves(s string) string {
	return strings.Replace(s, "%%GRAVE%%", "\x60", -1)
}

// createTemplateFromFile creates a template from a file.  The file is in the
// templates directory wherever the scaffolder is installed, and that is out of our
// control, so this should only be called when the "templatedir" command line
// argument is specified. 
func createTemplateFromFile(templateName string) *template.Template {
	log.SetPrefix("createTemplate() ")
	templateFile := templateRoot + templateName
	buf, err := ioutil.ReadFile(templateFile)
	if err != nil {
		log.Printf("cannot open template file %s - %s ",
			templateFile, err.Error())
		os.Exit(-1)
	}
	tp := string(buf)
	tp = substituteGraves(tp)
	return template.Must(template.New(templateName).Parse(tp))
}

func createTemplates(useBuiltIn bool) {

templateName := "controller.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
package {{.NameWithLowerFirst}}

{{.Imports}}

// Package {{.PluralNameWithLowerFirst}} provides the controller for the {{.PluralNameWithLowerFirst}} resource.  It provides a
// set of action functions that are triggered by HTTP requests and implement the
// Create, Read, Update and Delete (CRUD) operations on the {{.PluralNameWithLowerFirst}} resource:
//
//    GET {{.PluralNameWithLowerFirst}}/ - runs Index() to list all {{.PluralNameWithLowerFirst}}
//    GET {{.PluralNameWithLowerFirst}}/n - runs Show() to display the details of the {{.NameWithLowerFirst}} with ID n
//    GET {{.PluralNameWithLowerFirst}}/create - runs New() to display the page to create a {{.NameWithLowerFirst}} using any data in the form to pre-populate it
//    PUT {{.PluralNameWithLowerFirst}}/n - runs Create() to create a new {{.NameWithLowerFirst}} using the data in the supplied form
//    GET {{.PluralNameWithLowerFirst}}/n/edit - runs Edit() to display the page to edit the {{.NameWithLowerFirst}} with ID n, using any data in the form to pre-populate it
//    PUT {{.PluralNameWithLowerFirst}}/n - runs Update() to update the {{.NameWithLowerFirst}} with ID n using the data in the form
//    DELETE {{.PluralNameWithLowerFirst}}/n - runs Delete() to delete the {{.NameWithLowerFirst}} with id n

type Controller struct {
	services services.Services
	verbose bool
}

// MakeController is a factory that creates a {{.PluralNameWithLowerFirst}} controller
func MakeController(services services.Services, verbose bool) Controller {
	var controller Controller
	controller.SetServices(services)
	controller.SetVerbose(verbose)
	return controller
}

// Index fetches a list of all valid {{.PluralNameWithLowerFirst}} and displays the index page.
func (c Controller) Index(req *restful.Request, resp *restful.Response,
	form {{.NameWithLowerFirst}}Forms.ListForm) {

	log.SetPrefix("Index()")

	c.List{{.PluralNameWithUpperFirst}}(req, resp, form)
	return
}

// Show displays the details of the {{.NameWithLowerFirst}} with the ID given in the URI.
func (c Controller) Show(req *restful.Request, resp *restful.Response,
	form {{.NameWithLowerFirst}}Forms.SingleItemForm) {

	log.SetPrefix("Show()")

	repository := c.services.{{.NameWithUpperFirst}}Repository()

	// Get the details of the {{.NameWithLowerFirst}} with the given ID.
	{{.NameWithLowerFirst}}, err := repository.FindByID(form.{{.NameWithUpperFirst}}().ID())
	if err != nil {
		// no such {{.NameWithLowerFirst}}.  Display index page with error message
		em := "no such {{.NameWithLowerFirst}}"
		log.Printf("%s\n", em)
		c.ErrorHandler(req, resp, em)
		return
	}

	// The {{.NameWithLowerFirst}} in the form contains just an ID.  Replace it with the
	// complete {{.NameWithLowerFirst}} record that we just fetched.
	form.Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}})

	page := c.services.Template("{{.NameWithLowerFirst}}", "Show")
	if page == nil {
		em := fmt.Sprintf("internal error displaying Show page - no HTML template")
		log.Printf("%s\n", em)
		c.ErrorHandler(req, resp, em)
		return
	}

	err = page.Execute(resp.ResponseWriter, form)
	if err != nil {
		em := fmt.Sprintf("error displaying page - %s", err.Error())
		log.Printf("%s\n", em)
		c.ErrorHandler(req, resp, em)
		return
	}
	return
}

// New displays the page to create a new {{.NameWithLowerFirst}},
func (c Controller) New(req *restful.Request, resp *restful.Response,
	form {{.NameWithLowerFirst}}Forms.SingleItemForm) {

	log.SetPrefix("New()")

	// Display the page.
	page := c.services.Template("{{.NameWithLowerFirst}}", "Create")
	if page == nil {
		em := fmt.Sprintf("internal error displaying Create page - no HTML template")
		log.Printf("%s\n", em)
		c.ErrorHandler(req, resp, em)
		return
	}
	err := page.Execute(resp.ResponseWriter, form)
	if err != nil {
		log.Printf("error displaying new page - %s", err.Error())
		em := fmt.Sprintf("error displaying page - %s", err.Error())
		c.ErrorHandler(req, resp, em)
		return
	}
}

// Create creates a {{.NameWithLowerFirst}} using the data from the HTTP form displayed
// by a previous NEW request.
func (c Controller) Create(req *restful.Request, resp *restful.Response,
	form {{.NameWithLowerFirst}}Forms.SingleItemForm) {

	log.SetPrefix("Create()")

	if !(form.Valid()) {
		// validation errors.  Return to create screen with error messages in the form data
		page := c.services.Template("{{.NameWithLowerFirst}}", "Create")
		if page == nil {
			em := fmt.Sprintf("internal error displaying Create page - no HTML template")
			log.Printf("%s\n", em)
			c.ErrorHandler(req, resp, em)
			return
		}
		err := page.Execute(resp.ResponseWriter, &form)
		if err != nil {
			em := fmt.Sprintf("Internal error while preparing create form after failed validation - %s",
				err.Error())
			log.Printf("%s\n", em)
			c.ErrorHandler(req, resp, em)
			return
		}
		return
	}

	// Create a {{.NameWithLowerFirst}} in the database using the validated data in the form
	repository := c.services.{{.NameWithUpperFirst}}Repository()

	created{{.NameWithUpperFirst}}, err := repository.Create(form.{{.NameWithUpperFirst}}())
	if err != nil {
		// Failed to create {{.NameWithLowerFirst}}.  Display index page with error message.
		em := fmt.Sprintf("Could not create {{.NameWithLowerFirst}} %s - %s", form.{{.NameWithUpperFirst}}().DisplayName(), err.Error())
		c.ErrorHandler(req, resp, em)
		return
	}

	// Success! {{.NameWithUpperFirst}} created.  Display index page with confirmation notice
	notice := fmt.Sprintf("created {{.NameWithLowerFirst}} %s", created{{.NameWithUpperFirst}}.DisplayName())
	if c.verbose {
		log.Printf("%s\n", notice)
	}
	listForm := c.services.Make{{.NameWithUpperFirst}}ListForm()
	listForm.SetNotice(notice)
	c.List{{.PluralNameWithUpperFirst}}(req, resp, listForm)
	return
}

// Edit fetches the data for the {{.PluralNameWithLowerFirst}} record with the given ID and displays
// the edit page, populated with that data.
func (c Controller) Edit(req *restful.Request, resp *restful.Response,
	form {{.NameWithLowerFirst}}Forms.SingleItemForm) {

	log.SetPrefix("Edit() ")

	err := req.Request.ParseForm()
	if err != nil {
		// failed to parse form
		em := fmt.Sprintf("cannot parse form - %s", err.Error())
		log.Printf("%s\n", em)
		c.ErrorHandler(req, resp, em)
		return
	}
	// Get the ID of the {{.NameWithLowerFirst}}
	id := req.PathParameter("id")

	repository := c.services.{{.NameWithUpperFirst}}Repository()
	// Get the existing data for the {{.NameWithLowerFirst}}
	{{.NameWithLowerFirst}}, err := repository.FindByIDStr(id)
	if err != nil {
		// No such {{.NameWithLowerFirst}}.  Display index page with error message.
		em := err.Error()
		log.Printf("%s\n", em)
		c.ErrorHandler(req, resp, em)
		return
	}
	// Got the {{.NameWithLowerFirst}} with the given ID.  Put it into the form and validate it.
	// If the data is invalid, continue - the user may be trying to fix it.

	form.Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}})
	if !form.Validate() {
		em := fmt.Sprintf("invalid record in the {{.PluralNameWithLowerFirst}} database - %s",
			{{.NameWithLowerFirst}}.String())
		log.Printf("%s\n", em)
	}

	// Display the edit page
	page := c.services.Template("{{.NameWithLowerFirst}}", "Edit")
	if page == nil {
		em := fmt.Sprintf("internal error displaying Edit page - no HTML template")
		log.Printf("%s\n", em)
		c.ErrorHandler(req, resp, em)
		return
	}
	err = page.Execute(resp.ResponseWriter, form)
	if err != nil {
		// error while preparing edit page
		log.Printf("%s: error displaying edit page - %s", err.Error())
		em := fmt.Sprintf("error displaying page - %s", err.Error())
		c.ErrorHandler(req, resp, em)
	}
}

// Update responds to a PUT request.  For example:
// PUT /{{.PluralNameWithLowerFirst}}/1
// It's invoked by the form displayed by a previous Edit request.  If the ID in the URI is
// valid and the request parameters from the form specify valid {{.PluralNameWithLowerFirst}} data, it updates the
// record and displays the index page with a confirmation message, otherwise it displays
// the edit page again with the given data and some error messages.
func (c Controller) Update(req *restful.Request, resp *restful.Response,
	form {{.NameWithLowerFirst}}Forms.SingleItemForm) {

	log.SetPrefix("Update() ")
	
	if !form.Valid() {
		// The supplied data is invalid.  The validator has set error messages.  
		// Return to the edit screen.
		page := c.services.Template("{{.NameWithLowerFirst}}", "Edit")
		if page == nil {
			em := fmt.Sprintf("internal error displaying Edit page - no HTML template")
			log.Printf("%s\n", em)
			c.ErrorHandler(req, resp, em)
			return
		}
		err := page.Execute(resp.ResponseWriter, form)
		if err != nil {
			log.Printf("%s: error displaying edit page - %s", err.Error())
			em := fmt.Sprintf("error displaying page - %s", err.Error())
			c.ErrorHandler(req, resp, em)
			return
		}
		return
	}

	if form.{{.NameWithUpperFirst}}() == nil {
		em := fmt.Sprint("internal error - form should contain an updated {{.NameWithLowerFirst}} record")
		log.Printf("%s\n", em)
		c.ErrorHandler(req, resp, em)
		return
	}

	// Get the {{.NameWithLowerFirst}} specified in the form from the DB.
	// If that fails, the id in the form doesn't match any record.
	repository := c.services.{{.NameWithUpperFirst}}Repository()
	{{.NameWithLowerFirst}}, err := repository.FindByID(form.{{.NameWithUpperFirst}}().ID())
	if err != nil {
		// There is no {{.NameWithLowerFirst}} with this ID.  The ID is chosen by the user from a
		// supplied list and it should always be valid, so there's something screwy
		// going on.  Display the index page with an error message.
		em := fmt.Sprintf("error searching for {{.NameWithLowerFirst}} with id %s - %s",
			form.{{.NameWithUpperFirst}}().ID(), err.Error())
		log.Printf("%s\n", em)
		c.ErrorHandler(req, resp, em)
		return
	}

	// We have a matching {{.NameWithLowerFirst}} from the DB.
	if c.verbose {
		log.Printf("got {{.NameWithLowerFirst}} %v\n", {{.NameWithLowerFirst}})
	}

	// we have a record and valid new values.  Update.
	{{range .Fields}}
		{{$resourceNameLower}}.Set{{.NameWithUpperFirst}}(form.{{$resourceNameUpper}}().{{.NameWithUpperFirst}}())
	{{end}}
	if c.verbose {
		log.Printf("updating {{.NameWithLowerFirst}} to %v\n", {{.NameWithLowerFirst}})
	}
	_, err = repository.Update({{.NameWithLowerFirst}})
	if err != nil {
		// The commit failed.  Display the edit page with an error message
		em := fmt.Sprintf("Could not update {{.NameWithLowerFirst}} - %s", err.Error())
		log.Printf("%s\n", em)
		form.SetErrorMessage(em)

		page := c.services.Template("{{.NameWithLowerFirst}}", "Edit")
		if page == nil {
			em := fmt.Sprintf("internal error displaying Edit page - no HTML template")
			log.Printf("%s\n", em)
			c.ErrorHandler(req, resp, em)
			return
		}
		err = page.Execute(resp.ResponseWriter, form)
		if err != nil {
			// Error while recovering from another error.  This is looking like a habit!
			em := fmt.Sprintf("Internal error while preparing edit page after failing to update {{.NameWithLowerFirst}} in DB - %s", err.Error())
			log.Printf("%s\n", em)
			c.ErrorHandler(req, resp, em)
		} else {
			return
		}
	}

	// Success!  Display the index page with a confirmation notice
	notice := fmt.Sprintf("updated {{.NameWithLowerFirst}} %s", form.{{.NameWithUpperFirst}}().DisplayName())
	if c.verbose {
		log.Printf("%s:\n", notice)
	}
	listForm := c.services.Make{{.NameWithUpperFirst}}ListForm()
	listForm.SetNotice(notice)
	c.List{{.PluralNameWithUpperFirst}}(req, resp, listForm)
	return
}

// Delete responds to a DELETE request and deletes the record with the given ID,
// eg DELETE http://server:port/{{.PluralNameWithLowerFirst}}/1.
func (c Controller) Delete(req *restful.Request, resp *restful.Response,
	form {{.NameWithLowerFirst}}Forms.SingleItemForm) {

	log.SetPrefix("Delete()")

	repository := c.services.{{.NameWithUpperFirst}}Repository()
	// Attempt the delete
	_, err := repository.DeleteByID(form.{{.NameWithUpperFirst}}().ID())
	if err != nil {
		// failed - cannot delete {{.NameWithLowerFirst}}
		em := fmt.Sprintf("Cannot delete {{.NameWithLowerFirst}} with id %d - %s", 
			form.{{.NameWithUpperFirst}}().ID(), err.Error())
		log.Printf("%s\n", em)
		c.ErrorHandler(req, resp, em)
		return
	}
	// Success - {{.NameWithLowerFirst}} deleted.  Display the index view with a notification.
	listForm := c.services.Make{{.NameWithUpperFirst}}ListForm()
	notice := fmt.Sprintf("deleted {{.NameWithLowerFirst}} with id %d",
		form.{{.NameWithUpperFirst}}().ID())
	if c.verbose {
		log.Printf("%s:\n", notice)
	}
	listForm.SetNotice(notice)
	c.List{{.PluralNameWithUpperFirst}}(req, resp, listForm)
	return
}

// ErrorHandler displays the index page with an error message
func (c Controller) ErrorHandler(req *restful.Request, resp *restful.Response,
	errormessage string) {

	form := c.services.Make{{.NameWithUpperFirst}}ListForm()
	form.SetErrorMessage(errormessage)
	c.List{{.PluralNameWithUpperFirst}}(req, resp, form)
}

// SetServices sets the services.
func (c *Controller) SetServices(services services.Services) {
	c.services = services
}

// SetVerbose sets the verbosity level.
func (c *Controller) SetVerbose(verbose bool) {
	c.verbose = verbose
}

/*
 * The List{{.PluralNameWithUpperFirst}} helper method fetches a list of {{.PluralNameWithLowerFirst}} and displays the
 * index page.  It's used to fulfil an index request but the index page is
 * also used as the last page of a sequence of requests (for example new,
 * create, index).  If the sequence was successful, the form may contain a
 * confirmation note.  If the sequence failed, the form should contain an error
 * message.
 */
func (c Controller) List{{.PluralNameWithUpperFirst}}(req *restful.Request, resp *restful.Response,
	form {{.NameWithLowerFirst}}Forms.ListForm) {

	log.SetPrefix("Controller.List{{.PluralNameWithUpperFirst}}() ")

	repository := c.services.{{.NameWithUpperFirst}}Repository()

	{{.PluralNameWithLowerFirst}}List, err := repository.FindAll()
	if err != nil {
		em := fmt.Sprintf("error getting the list of {{.PluralNameWithLowerFirst}} - %s", err.Error())
		log.Printf("%s\n", em)
		form.SetErrorMessage(em)
	}
	if c.verbose{
		log.Printf("%d {{.PluralNameWithLowerFirst}}", len({{.PluralNameWithLowerFirst}}List))
	}
	if len({{.PluralNameWithLowerFirst}}List) <= 0 {
		form.SetNotice("there are no {{.PluralNameWithLowerFirst}} currently set up")
	}
	form.Set{{.PluralNameWithUpperFirst}}({{.PluralNameWithLowerFirst}}List)

	// Display the index page
	page := c.services.Template("{{.NameWithLowerFirst}}", "Index")
	if page == nil {
		log.Printf("no Index page for {{.NameWithLowerFirst}} controller")
		utilities.Dead(resp)
		return
	}
	err = page.Execute(resp.ResponseWriter, form)
	if err != nil {
		/*
		 * Error while displaying the index page.  We handle most internal
		 * errors by displaying the controller's index page.  That's just failed,
		 * so fall back to the static error page.
		 */
		log.Printf(err.Error())
		page = c.services.Template("html", "Error")
		if page == nil {
			log.Printf("no Error page")
			utilities.Dead(resp)
			return
		}
		err = page.Execute(resp.ResponseWriter, form)
		if err != nil {
			// Can't display the static error page either.  Bale out.
			em := fmt.Sprintf("fatal error - failed to display error page for error %s\n", err.Error())
			log.Printf(em)
			panic(em)
		}
		return
	}
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "controller.test.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
{{$resourceNamePluralUpper := .PluralNameWithUpperFirst}}
package {{.NameWithLowerFirst}}

{{.Imports}}

// Unit tests for the {{.NameWithLowerFirst}} controller.  Uses mock objects
// created by pegomock.

var panicValue string

{{/* This creates the expected values using the field names and the test 
     values, something like:
	 var expectedName1 string = "s1"
	 var expectedAge1 int64 = 2 
	 var expectedName2 string = "s3"
	 var expectedAge2 int64 = 4 */}}
{{range $index, $element := .Fields}}
	{{if eq .Type "string"}}
		var expected{{.NameWithUpperFirst}}1 {{.GoType}} = "{{index .TestValues 0}}"
	{{else}}
		var expected{{.NameWithUpperFirst}}1 {{.GoType}} = {{index .TestValues 0}}
	{{end}}
	{{if eq .Type "string"}}
		var expected{{.NameWithUpperFirst}}2 {{.GoType}} = "{{index .TestValues 1}}"
	{{else}}
		var expected{{.NameWithUpperFirst}}2 {{.GoType}} = {{index .TestValues 1}}
	{{end}}
{{end}}

// TestUnitIndexWithOne{{.NameWithUpperFirst}} checks that the Index method of the 
// {{.NameWithLowerFirst}} controller handles a list of {{.PluralNameWithLowerFirst}} from FindAll() containing one {{.NameWithLowerFirst}}.
func TestUnitIndexWithOne{{.NameWithUpperFirst}}(t *testing.T) {

	var expectedID1 uint64 = 42
	
	pegomock.RegisterMockTestingT(t)

	// Create a list containing one {{.NameWithLowerFirst}}.
	expected{{.NameWithUpperFirst}}1 := {{.NameWithLowerFirst}}.MakeInitialised{{$resourceNameUpper}}(expectedID1, {{range .Fields}}expected{{.NameWithUpperFirst}}1{{if not .LastItem}}, {{end}}{{end}})
	expected{{.NameWithUpperFirst}}List := make([]{{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}, 1)
	expected{{.NameWithUpperFirst}}List[0] = expected{{.NameWithUpperFirst}}1

	// Create the mocks and dummy objects.
	var url url.URL
	url.Opaque = "/{{.PluralNameWithLowerFirst}}" // url.RequestURI() will return "/{{.PluralNameWithLowerFirst}}"
	var httpRequest http.Request
	httpRequest.URL = &url
	httpRequest.Method = "GET"
	var request restful.Request
	request.Request = &httpRequest
	writer := mocks.NewMockResponseWriter()
	var response restful.Response
	response.ResponseWriter = writer
	mockTemplate := mocks.NewMockTemplate()
	mockRepository := mock{{.NameWithUpperFirst}}.NewMockRepository()
	
	innerPageMap := make(map[string]retrofitTemplate.Template)
	innerPageMap["Index"] = mockTemplate
	pageMap := make(map[string]map[string]retrofitTemplate.Template)
	pageMap["{{.NameWithLowerFirst}}"] = innerPageMap

	// Create a service that returns the mock repository and templates.
	var services services.ConcreteServices
	services.Set{{.NameWithUpperFirst}}Repository(mockRepository)
	services.SetTemplates(&pageMap)

	// Create the form
	form := {{.NameWithLowerFirst}}Forms.MakeListForm()

	// Expect the controller to call the {{.NameWithLowerFirst}} repository's FindAll method.  Return
	// the list containing one {{.NameWithLowerFirst}}.
	pegomock.When(mockRepository.FindAll()).ThenReturn(expected{{.NameWithUpperFirst}}List, nil)
	
	// The request supplies method "GET" and URI "/{{.PluralNameWithLowerFirst}}".  Expect
	// template.Execute to be called and return nil (no error).
	pegomock.When(mockTemplate.Execute(writer, form)).ThenReturn(nil)

	// Run the test.
	var controller Controller
	controller.SetServices(&services)
	controller.Index(&request, &response, form)

	// We expect that the form contains the expected {{.NameWithLowerFirst}} list -
	// one {{.NameWithLowerFirst}} object with contents as expected.
	if form.{{.PluralNameWithUpperFirst}}() == nil {
		t.Errorf("Expected a list, got nil")
	}

	if len(form.{{.PluralNameWithUpperFirst}}()) != 1 {
		t.Errorf("Expected a list of 1, got %d", len(form.{{.PluralNameWithUpperFirst}}()))
	}

	if form.{{.PluralNameWithUpperFirst}}()[0].ID() != expectedID1 {
		t.Errorf("Expected ID %d, got %d",
			expectedID1, form.{{.PluralNameWithUpperFirst}}()[0].ID())
	}
{{range .Fields}}
	if form.{{$resourceNamePluralUpper}}()[0].{{.NameWithUpperFirst}}() != expected{{.NameWithUpperFirst}}1 {
		t.Errorf("Expected {{.NameWithLowerFirst}} %v, got %v",
			expected{{.NameWithUpperFirst}}1, form.{{$resourceNamePluralUpper}}()[0].{{.NameWithUpperFirst}}())
	}
{{end}}
}

// TestUnitIndexWithErrorWhenFetching{{.PluralNameWithUpperFirst}} checks that the {{.NameWithLowerFirst}} controller's
// Index() method handles errors from FindAll() correctly.
func TestUnitIndexWithErrorWhenFetching{{.PluralNameWithUpperFirst}}(t *testing.T) {

	log.SetPrefix("TestUnitIndexWithErrorWhenFetching{{.PluralNameWithUpperFirst}} ")
	log.Printf("This test is expected to provoke error messages in the log")

	expectedErr := errors.New("Test Error Message")
	expectedErrorMessage := "error getting the list of {{.PluralNameWithLowerFirst}} - Test Error Message"

	// Create the mocks and dummy objects.
	pegomock.RegisterMockTestingT(t)
	var url url.URL
	url.Opaque = "/{{.PluralNameWithLowerFirst}}" // url.RequestURI() will return "/{{.PluralNameWithLowerFirst}}"
	var httpRequest http.Request
	httpRequest.URL = &url
	httpRequest.Method = "GET"
	var request restful.Request
	request.Request = &httpRequest
	writer := mocks.NewMockResponseWriter()
	var response restful.Response
	response.ResponseWriter = writer
	mockTemplate := mocks.NewMockTemplate()
	mockRepository := mock{{.NameWithUpperFirst}}.NewMockRepository()
	
	// Create the form
	form := {{.NameWithLowerFirst}}Forms.MakeListForm()


	// Expect the controller to call the {{.NameWithLowerFirst}} repository's FindAll method.  Return
	// the list containing one {{.NameWithLowerFirst}}.
	pegomock.When(mockRepository.FindAll()).ThenReturn(nil, expectedErr)
	
	// Expect the controller to call the tenmplate's Execute() method.  Return
	// nil (no error).
	pegomock.When(mockTemplate.Execute(writer, form)).ThenReturn(nil)
	
	innerPageMap := make(map[string]retrofitTemplate.Template)
	innerPageMap["Index"] = mockTemplate
	pageMap := make(map[string]map[string]retrofitTemplate.Template)
	pageMap["{{.NameWithLowerFirst}}"] = innerPageMap

	// Create a service that returns the mock repository and templates.
	var services services.ConcreteServices
	services.Set{{.NameWithUpperFirst}}Repository(mockRepository)
	services.SetTemplates(&pageMap)

	// Create the controller and run the test.
	controller := MakeController(&services, false)
	controller.Index(&request, &response, form)

	// Verify that the form contains the expected error message.
	if form.ErrorMessage() != expectedErrorMessage {
		t.Errorf("Expected error message to be %s actually %s", expectedErrorMessage, form.ErrorMessage())
	}
}


// TestUnitIndexWithManyFailures checks that the {{.PluralNameWithUpperFirst}} controller's
// Index() method handles a series of errors correctly.
//
// Panic handling based on http://stackoverflow.com/questions/31595791/how-to-test-panics
//
func TestUnitIndexWithManyFailures(t *testing.T) {

	log.SetPrefix("TestUnitIndexWithManyFailures ")
	log.Printf("This test is expected to provoke error messages in the log")
	
	em1 := "first error message"

	expectedFirstErrorMessage := errors.New(em1)

	em2 := "second error message"
	expectedSecondErrorMessage := errors.New(em2)

	em3 := "final error message"
	finalErrorMessage := errors.New(em3)

	// Create the mocks and dummy objects.
	pegomock.RegisterMockTestingT(t)
	var url url.URL
	url.Opaque = "/{{.PluralNameWithLowerFirst}}" // url.RequestURI() will return "/{{.PluralNameWithLowerFirst}}"
	var httpRequest http.Request
	httpRequest.URL = &url
	httpRequest.Method = "GET"
	var request restful.Request
	request.Request = &httpRequest
	mockResponseWriter := mocks.NewMockResponseWriter()
	var response restful.Response
	response.ResponseWriter = mockResponseWriter
	mockIndexTemplate := mocks.NewMockTemplate()
	mockErrorTemplate := mocks.NewMockTemplate()
	mockRepository := mock{{.NameWithUpperFirst}}.NewMockRepository()
	
	// Create a template map containing the mock templates
	pageMap := make(map[string]map[string]retrofitTemplate.Template)
	pageMap["html"] = make(map[string]retrofitTemplate.Template)
	pageMap["html"]["Error"] = mockErrorTemplate
	pageMap["{{.NameWithLowerFirst}}"] = make(map[string]retrofitTemplate.Template)
	pageMap["{{.NameWithLowerFirst}}"]["Index"] = mockIndexTemplate

	// Create a service that returns the mock repository and templates.
	var services services.ConcreteServices
	services.Set{{.NameWithUpperFirst}}Repository(mockRepository)
	services.SetTemplates(&pageMap)

	// Create the form
	form := {{.NameWithLowerFirst}}Forms.MakeListForm()

	// Expectations:
	// Index will run List{{.PluralNameWithUpperFirst}} which will call the {{.NameWithLowerFirst}}
	// repository's FindAll().  Make that return an error, then List{{.PluralNameWithUpperFirst}} 
	// will get the Index page from the template and call its Execute method.  Make 
	// that fail, and the controller will get the error page and call its Execute 
	// method.  Make that fail and the app will panic with a message "fatal error - 
	// failed to display error page for error ", followed by the error message from 
	// the last Execute call.

	pegomock.When(mockRepository.FindAll()).ThenReturn(nil, 
		expectedFirstErrorMessage)
	pegomock.When(mockIndexTemplate.Execute(mockResponseWriter, form)).
		ThenReturn(expectedSecondErrorMessage)
	pegomock.When(mockErrorTemplate.Execute(mockResponseWriter, form)).
		ThenReturn(finalErrorMessage)

	// Expect a panic, catch it and check the value.  (If there is no panic,
	// this raises an error.)

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("Expected the Index call to panic")
		} else {
			em := fmt.Sprintf("%s", r)
			// Verify that the panic value is as expected.
			if !strings.Contains(em, em3) {
				t.Errorf("Expected a panic with value containing \"%s\" actually \"%s\"",
					em3, em)
			}
		}
	}()

	// Run the test.
	controller := MakeController(&services, false)
	controller.Index(&request, &response, form)

	// Verify that the form has an error message containing the expected text.
	if strings.Contains(form.ErrorMessage(), em1) {
		t.Errorf("Expected error message to be \"%s\" actually \"%s\"",
			expectedFirstErrorMessage, form.ErrorMessage())
	}

	// Verify that the list of {{.PluralNameWithLowerFirst}} is nil
	if form.{{.PluralNameWithUpperFirst}}() != nil {
		t.Errorf("Expected the list of {{.PluralNameWithLowerFirst}} to be nil.  Actually contains %d entries",
			len(form.{{.PluralNameWithUpperFirst}}()))
	}

}

// TestUnitSuccessfulCreate checks that the {{.NameWithLowerFirst}} controller's Create method
// correctly handles a successful attempt to create a {{.NameWithLowerFirst}} in the database.
func TestUnitSuccessfulCreate(t *testing.T) {

	log.SetPrefix("TestUnitSuccessfulCreate ")

	expectedNoticeFragment := "created {{.NameWithLowerFirst}}"
{{range .Fields}}
{{if not .ExcludeFromDisplay}}
{{if eq .Type "int"}}
	expected{{.NameWithUpperFirst}}1_str := fmt.Sprintf("%d", expected{{.NameWithUpperFirst}}1)
{{end}}
{{if eq .Type "float"}}
	expected{{.NameWithUpperFirst}}1_str := fmt.Sprintf("%f", expected{{.NameWithUpperFirst}}1)
{{end}}
{{if eq .Type "bool"}}
	expected{{.NameWithUpperFirst}}1_str := fmt.Sprintf("%v", expected{{.NameWithUpperFirst}}1)
{{end}}
{{end}}
{{end}}
	pegomock.RegisterMockTestingT(t)

	// Create the mocks and dummy objects.
	var expectedID1 uint64 = 42
	expected{{.NameWithUpperFirst}}1 := {{.NameWithLowerFirst}}.MakeInitialised{{$resourceNameUpper}}(expectedID1, {{range .Fields}}expected{{.NameWithUpperFirst}}1{{if not .LastItem}}, {{end}}{{end}})
	singleItemForm := {{.NameWithLowerFirst}}Forms.MakeInitialisedSingleItemForm(expected{{.NameWithUpperFirst}}1)
	listForm := {{.NameWithLowerFirst}}Forms.MakeListForm()
	{{.PluralNameWithLowerFirst}} := make([]{{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}, 1)
	{{.PluralNameWithLowerFirst}}[0] = expected{{.NameWithUpperFirst}}1
	listForm.Set{{.PluralNameWithUpperFirst}}({{.PluralNameWithLowerFirst}})
	var url url.URL
	url.Opaque = "/{{.PluralNameWithLowerFirst}}/42" // url.RequestURI() will return "/{{.PluralNameWithLowerFirst}}/42"
	var httpRequest http.Request
	httpRequest.URL = &url
	httpRequest.Method = "POST"
	var request restful.Request
	request.Request = &httpRequest
	writer := mocks.NewMockResponseWriter()
	var response restful.Response
	response.ResponseWriter = writer
	mockIndexTemplate := mocks.NewMockTemplate()
	mockCreateTemplate := mocks.NewMockTemplate()
	mockRepository := mock{{.NameWithUpperFirst}}.NewMockRepository()
	mockServices := mocks.NewMockServices()

	// Create a template map containing the mock templates
	pageMap := make(map[string]map[string]retrofitTemplate.Template)
	pageMap["{{.NameWithLowerFirst}}"] = make(map[string]retrofitTemplate.Template)
	pageMap["{{.NameWithLowerFirst}}"]["Index"] = mockIndexTemplate
	pageMap["{{.NameWithLowerFirst}}"]["Create"] = mockCreateTemplate
	
	// Set expectations. The controller will display the Create template,
	// get some data, create a repository and use it to create a model object.
	// Then it will use the Index template to display the index page. 
	pegomock.When(mockServices.Template("{{.NameWithLowerFirst}}", "Create")).ThenReturn(mockCreateTemplate)
	pegomock.When(mockServices.{{.NameWithUpperFirst}}Repository()).ThenReturn(mockRepository)
	pegomock.When(mockRepository.Create(expected{{.NameWithUpperFirst}}1)).
		ThenReturn(expected{{.NameWithUpperFirst}}1, nil)
	pegomock.When(mockServices.Make{{.NameWithUpperFirst}}ListForm()).ThenReturn(listForm)
	pegomock.When(mockServices.Template("{{.NameWithLowerFirst}}", "Index")).ThenReturn(mockIndexTemplate)
	pegomock.When(mockRepository.FindAll()).ThenReturn({{.PluralNameWithLowerFirst}}, nil)
	pegomock.When(mockCreateTemplate.Execute(response.ResponseWriter, listForm)).
		ThenReturn(nil)

	// Run the test.
	controller := MakeController(mockServices, false)
	controller.Create(&request, &response, singleItemForm)

	// Verify that the form contains a notice with the expected contents.
	if !strings.Contains(listForm.Notice(), expectedNoticeFragment) {
		t.Errorf("Expected notice to contain \"%s\" actually \"%s\"",
			expectedNoticeFragment, listForm.Notice())
	}
{{range .Fields}}
{{if not .ExcludeFromDisplay}}
{{if eq .Type "string"}}
	if !strings.Contains(listForm.Notice(), expected{{.NameWithUpperFirst}}1) {
		t.Errorf("Expected notice to contain \"%s\" actually \"%s\"",
			expected{{.NameWithUpperFirst}}1, listForm.Notice())
	}
{{else}}
	if !strings.Contains(listForm.Notice(), expected{{.NameWithUpperFirst}}1_str) {
		t.Errorf("Expected notice to contain \"%s\" actually \"%s\"",
			expected{{.NameWithUpperFirst}}1_str, listForm.Notice())
	}
{{end}}
{{end}}
{{end}}
}

// TestUnitCreateFailsWithMissingFields checks that the {{.NameWithLowerFirst}} controller's 
// Create method correctly handles invalid data from the HTTP request.  Note: by 
// the time the code under test runs, number and boolean fields have already been 
// extracted from the HTML form and converted, so the only fields that can be made 
// invalid are mandatory string fields.  If there are none of those, the test will
// run successfully but it will do nothing useful.
// 
// The test uses pegomock to provide mocks.
func TestUnitCreateFailsWithMissingFields(t *testing.T) {

	log.SetPrefix("TestUnitCreateFailsWithMissingFields ")
{{range .Fields}}
	{{if and .Mandatory (eq .Type "string")}}
		expectedErrorMessage{{.NameWithUpperFirst}} := "you must specify the {{.NameWithLowerFirst}}"
	{{end}}
{{end}}
	pegomock.RegisterMockTestingT(t)
	
	var expectedID1 uint64 = 42
	// supply empty string for mandatory string fields, the given values for others.
	expected{{.NameWithUpperFirst}}1 := {{.NameWithLowerFirst}}.MakeInitialised{{$resourceNameUpper}}(expectedID1, {{range .Fields}}{{if and .Mandatory (eq .Type "string")}}"  "{{else}}expected{{.NameWithUpperFirst}}1{{end}}{{if not .LastItem}}, {{end}}{{end}})
	singleItemForm := {{.NameWithLowerFirst}}Forms.MakeInitialisedSingleItemForm(expected{{.NameWithUpperFirst}}1)

	// Create the mocks and dummy objects.
	
	var url url.URL
	url.Opaque = "/{{.PluralNameWithLowerFirst}}/42" // url.RequestURI() will return "/{{.PluralNameWithLowerFirst}}/42"
	var httpRequest http.Request
	httpRequest.URL = &url
	httpRequest.Method = "POST"
	var request restful.Request
	request.Request = &httpRequest
	writer := mocks.NewMockResponseWriter()
	var response restful.Response
	response.ResponseWriter = writer
	mockTemplate := mocks.NewMockTemplate()

	// Create a services layer that returns the mock create template.
	mockServices := mocks.NewMockServices()
	pegomock.When(mockServices.Template("{{.NameWithLowerFirst}}", "Create")).ThenReturn(mockTemplate)

	// Run the test.
	controller := MakeController(mockServices, false)

	controller.Create(&request, &response, singleItemForm)

	// If the {{.NameWithLowerFirst}} has mandatory string fields, verify that the 
	// form contains the expected error messages.
{{range .Fields}}
	{{if and .Mandatory (eq .Type "string")}}
		if singleItemForm.ErrorForField("{{.NameWithUpperFirst}}") != expectedErrorMessage{{.NameWithUpperFirst}} {
			t.Errorf("Expected error message to be %s actually %s",
				expectedErrorMessage{{.NameWithUpperFirst}}, singleItemForm.ErrorForField("{{.NameWithUpperFirst}}"))
		}
	{{end}}
{{end}}
}

// TestUnitCreateFailsWithDBError checks that the {{.NameWithLowerFirst}} handler's Create method
// correctly handles an error from the repository while attempting to create a
// {{.NameWithLowerFirst}} in the database.
func TestUnitCreateFailsWithDBError(t *testing.T) {

	log.SetPrefix("TestUnitCreateFailsWithDBError ")

	expectedErrorMessage := "some error"
	expectedErrorMessageLeader := "Could not create {{.NameWithLowerFirst}}"

	pegomock.RegisterMockTestingT(t)

	// Create the mocks and dummy objects.
	var expectedID1 uint64 = 42
	expected{{.NameWithUpperFirst}}1 := {{.NameWithLowerFirst}}.MakeInitialised{{$resourceNameUpper}}(expectedID1, {{range .Fields}}expected{{.NameWithUpperFirst}}1{{if not .LastItem}}, {{end}}{{end}})
	singleItemForm := {{.NameWithLowerFirst}}Forms.MakeInitialisedSingleItemForm(expected{{.NameWithUpperFirst}}1)
	listForm := {{.NameWithLowerFirst}}Forms.MakeListForm()
	var url url.URL
	url.Opaque = "/{{.PluralNameWithLowerFirst}}/42" // url.RequestURI() will return "/{{.PluralNameWithLowerFirst}}/42"
	var httpRequest http.Request
	httpRequest.URL = &url
	httpRequest.Method = "POST"
	var request restful.Request
	request.Request = &httpRequest
	writer := mocks.NewMockResponseWriter()
	var response restful.Response
	response.ResponseWriter = writer
	mockIndexTemplate := mocks.NewMockTemplate()
	mockCreateTemplate := mocks.NewMockTemplate()

	// Create a services layer that returns the mock create template.
	mockRepository := mock{{.NameWithUpperFirst}}.NewMockRepository()
	mockServices := mocks.NewMockServices()
	pegomock.When(mockServices.Template("{{.NameWithLowerFirst}}", "Create")).
		ThenReturn(mockCreateTemplate)
	pegomock.When(mockServices.{{.NameWithUpperFirst}}Repository()).ThenReturn(mockRepository)
	pegomock.When(mockRepository.Create(expected{{.NameWithUpperFirst}}1)).
		ThenReturn(nil, errors.New(expectedErrorMessage))
	pegomock.When(mockServices.Template("{{.NameWithLowerFirst}}", "Index")).
		ThenReturn(mockIndexTemplate)
	pegomock.When(mockServices.Make{{.NameWithUpperFirst}}ListForm()).ThenReturn(listForm)

	// Run the test.
	controller := MakeController(mockServices, false)

	controller.Create(&request, &response, singleItemForm)

	// Verify that the form contains the expected error message.
	if !strings.Contains(listForm.ErrorMessage(), expectedErrorMessageLeader) {
		t.Errorf("Expected error message to contain \"%s\" actually \"%s\"",
			expectedErrorMessageLeader, listForm.ErrorMessage())
	}

	if !strings.Contains(listForm.ErrorMessage(), expectedErrorMessage) {
		t.Errorf("Expected error message to contain \"%s\" actually \"%s\"",
			expectedErrorMessage, listForm.ErrorMessage())
	}
}

// Recover from any panic and record the error.
func catchPanic() {
	log.SetPrefix("catchPanic ")
	if p := recover(); p != nil {
		em := fmt.Sprintf("%v", p)
		panicValue = em
		log.Printf(em)
	}
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "form.concrete.list.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
package {{.NameWithLowerFirst}}

{{.Imports}}

// The ConcreteListForm satisfies the ListForm interface and holds view data 
// including a list of {{.PluralNameWithLowerFirst}}.  It's approximately equivalent
// to a Struts form bean.
type ConcreteListForm struct {
	{{.PluralNameWithLowerFirst}}       []{{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}
	notice       string
	errorMessage string
}

// Define the factory functions.

// MakeListForm creates and returns a new uninitialised ListForm object
func MakeListForm() ListForm {
	var concreteListForm ConcreteListForm
	return &concreteListForm
}

// {{.PluralNameWithUpperFirst}} returns the list of {{.NameWithUpperFirst}} objects from the form
func (clf *ConcreteListForm) {{.PluralNameWithUpperFirst}}() []{{.NameWithLowerFirst}}.{{.NameWithUpperFirst}} {
	return clf.{{.PluralNameWithLowerFirst}}
}

// Notice gets the notice.
func (clf *ConcreteListForm) Notice() string {
	return clf.notice
}

// ErrorMessage gets the general error message.
func (clf *ConcreteListForm) ErrorMessage() string {
	return clf.errorMessage
}

// Set{{.PluralNameWithUpperFirst}} sets the list of {{.NameWithUpperFirst}}s.
func (clf *ConcreteListForm) Set{{.PluralNameWithUpperFirst}}({{.PluralNameWithLowerFirst}} []{{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}) {
	clf.{{.PluralNameWithLowerFirst}} = {{.PluralNameWithLowerFirst}}
}

// SetNotice sets the notice.
func (clf *ConcreteListForm) SetNotice(notice string) {
	clf.notice = notice
}

// SetErrorMessage sets the error message.
func (clf *ConcreteListForm) SetErrorMessage(errorMessage string) {
	clf.errorMessage = errorMessage
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "form.concrete.single.item.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
package {{.NameWithLowerFirst}}

{{.Imports}}

// ConcreteSingleItemForm satisfies the {{.NameWithLowerFirst}} SingleItemForm interface.
type ConcreteSingleItemForm struct {
	{{.NameWithLowerFirst}} {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}
	errorMessage string
	notice       string
	fieldError   map[string]string
	isValid      bool
}

// Define the factory functions.

// MakeSingleItemForm creates and returns a new uninitialised form object
func MakeSingleItemForm() SingleItemForm {
	var concreteSingleItemForm ConcreteSingleItemForm
	return &concreteSingleItemForm
}

// MakeInitialisedSingleItemForm creates and returns a new form object
// containing the given {{.NameWithLowerFirst}}.
func MakeInitialisedSingleItemForm({{.NameWithLowerFirst}} {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}) SingleItemForm {
	form := MakeSingleItemForm()
	form.Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}})
	form.SetValid(form.Validate())
	return form
}

// Getters

// {{.NameWithUpperFirst}} gets the {{.NameWithLowerFirst}} embedded in the form.
func (form ConcreteSingleItemForm) {{.NameWithUpperFirst}}() {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}} {
	return form.{{.NameWithLowerFirst}}
}

// Notice gets the notice.
func (form ConcreteSingleItemForm) Notice() string {
	return form.notice
}

// ErrorMessage gets the general error message.
func (form ConcreteSingleItemForm) ErrorMessage() string {
	return form.errorMessage
}

// FieldErrors returns all the field errors as a map.
func (form ConcreteSingleItemForm) FieldErrors() map[string]string {
	return form.fieldError
}

// ErrorForField returns the error message about a field (may be an empty string).
func (form ConcreteSingleItemForm) ErrorForField(key string) string {
	if form.fieldError == nil {
		// The field error map has not been set up.
		return ""
	}
	return form.fieldError[key]
}

// Valid returns true if the contents of the form is valid
func (form ConcreteSingleItemForm) Valid() bool { 
	return form.isValid
}

// String returns a string version of the {{.NameWithUpperFirst}}Form.
func (form ConcreteSingleItemForm) String() string {
	return fmt.Sprintf("ConcreteSingleItemForm={{"{"}}{{.NameWithLowerFirst}}=%s, notice=%s,errorMessage=%s,fieldError=%s{{"}"}}",
		form.{{.NameWithLowerFirst}},
		form.notice,
		form.errorMessage,
		utilities.Map2String(form.fieldError))
}

// Setters

// Set{{.NameWithUpperFirst}} sets the {{.NameWithUpperFirst}} in the form.
func (form *ConcreteSingleItemForm) Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}} {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}) {
	form.{{.NameWithLowerFirst}} = {{.NameWithLowerFirst}}
}

// SetNotice sets the notice.
func (form *ConcreteSingleItemForm) SetNotice(notice string) {
	form.notice = notice
}

//SetErrorMessage sets the general error message.
func (form *ConcreteSingleItemForm) SetErrorMessage(errorMessage string) {
	form.errorMessage = errorMessage
}

// SetErrorMessageForField sets the error message for a named field
func (form *ConcreteSingleItemForm) SetErrorMessageForField(fieldname, errormessage string) {
	if form.fieldError == nil {
		form.fieldError = make(map[string]string)
	}
	form.fieldError[fieldname] = errormessage
}

// SetValid sets a warning that the data in the form is invalid
func (form *ConcreteSingleItemForm) SetValid(value bool) {
	form.isValid = value
}

// Validate validates the data in the {{.NameWithUpperFirst}} and sets the various error messages.
// It returns true if the data is valid, false if there are errors.
func (form *ConcreteSingleItemForm) Validate() bool {
	valid := true

	// Trim and test all mandatory string items. 
	{{range .Fields}}
		{{if and .Mandatory (eq .Type "string")}}
			if len(strings.TrimSpace(form.{{$resourceNameLower}}.{{.NameWithUpperFirst}}())) <= 0 {
					form.SetErrorMessageForField("{{.NameWithUpperFirst}}", "you must specify the {{.NameWithLowerFirst}}")
					valid = false
				}
		{{end}}
	{{end}}
	return valid
}`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "form.concrete.single.item.test.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
package {{.NameWithLowerFirst}}

{{.Imports}}

var expectedID1 uint64 = 42
var expectedID2 uint64 = 43
{{/* This creates the expected values using the field names and the test 
     values, something like:
	 var expectedName1 string = "s1"
	 var expectedAge1 int = 2 
	 var expectedName2 string = "s3"
	 var expectedAge2 int = 4 */}}
{{range $index, $element := .Fields}}
	{{if eq .Type "string"}}
		var expected{{.NameWithUpperFirst}}1 {{.GoType}} = "{{index .TestValues 0}}"
	{{else}}
		var expected{{.NameWithUpperFirst}}1 {{.GoType}} = {{index .TestValues 0}}
	{{end}}
	{{if eq .Type "string"}}
		var expected{{.NameWithUpperFirst}}2 {{.GoType}} = "{{index .TestValues 1}}"
	{{else}}
		var expected{{.NameWithUpperFirst}}2 {{.GoType}} = {{index .TestValues 1}}
	{{end}}
{{end}}

// Create a {{.NameWithLowerFirst}} and a ConcreteSingleItemForm containing it.  Retrieve the {{.NameWithLowerFirst}}.
func TestUnitCreate{{.NameWithUpperFirst}}FormAndRetrieve{{.NameWithUpperFirst}}(t *testing.T) {
	{{.NameWithLowerFirst}}Form := Create{{.NameWithUpperFirst}}Form(expectedID1, {{range .Fields}}expected{{.NameWithUpperFirst}}1{{if not .LastItem}}, {{end}}{{end}})
	if {{.NameWithLowerFirst}}Form.{{.NameWithUpperFirst}}().ID() != expectedID1 {
		t.Errorf("Expected ID to be %d actually %d", expectedID1, {{.NameWithLowerFirst}}Form.{{.NameWithUpperFirst}}().ID())
	}
	{{range .Fields}}
		if {{$resourceNameLower}}Form.{{$resourceNameUpper}}().{{.NameWithUpperFirst}}() != expected{{.NameWithUpperFirst}}1 {
			t.Errorf("Expected {{.NameWithLowerFirst}} to be %s actually %s", expected{{.NameWithUpperFirst}}1, {{$resourceNameLower}}Form.{{$resourceNameUpper}}().{{.NameWithUpperFirst}}())
		}
	{{end}}
}

{{$fields := .Fields}}
{{range .Fields}}
	{{if .Mandatory}}
		{{if eq .Type "string"}}
			{{$thisField := .NameWithLowerFirst}}
			{{$thisFieldUpper := .NameWithUpperFirst}}
			// Create a {{$resourceNameUpper}}Form containing a {{$resourceNameLower}} with no {{.NameWithLowerFirst}}, and validate it.
			func TestUnitCreate{{$resourceNameUpper}}FormNo{{.NameWithUpperFirst}}(t *testing.T) {
				expectedError := "you must specify the {{.NameWithLowerFirst}}"
				{{$resourceNameLower}}Form := Create{{$resourceNameUpper}}Form(expectedID2, {{range $fields}}{{if eq $thisField .NameWithLowerFirst}}""{{if not .LastItem}}, {{end}}{{else}}expected{{.NameWithUpperFirst}}2{{if .LastItem}}){{else}}, {{end}}{{end}}{{end}}
				if {{$resourceNameLower}}Form.Validate() {
					t.Errorf("Expected the validation to fail with missing {{$thisField}}")
				} else {
					if {{$resourceNameLower}}Form.ErrorForField("{{$thisFieldUpper}}") != expectedError {
						t.Errorf("Expected \"%s\", got \"%s\"", expectedError,
							{{$resourceNameLower}}Form.ErrorForField("{{$thisFieldUpper}}"))
					}
				}
				errors := {{$resourceNameLower}}Form.FieldErrors()
				if len(errors) != 1 {
					t.Errorf("Expected 1 error, got %d", len(errors))
				}
			}
		{{end}}
	{{end}}
{{end}}


func Create{{.NameWithUpperFirst}}Form(id uint64, {{range .Fields}}{{.NameWithLowerFirst}} {{.GoType}}{{if not .LastItem}}, {{end}}{{end}}) ConcreteSingleItemForm {
	var {{.NameWithLowerFirst}} {{.NameWithLowerFirst}}Model.Concrete{{.NameWithUpperFirst}}
	{{.NameWithLowerFirst}}.SetID(id)
	{{range .Fields}}
		{{$resourceNameLower}}.Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}})
	{{end}}
	var {{.NameWithLowerFirst}}Form ConcreteSingleItemForm
	{{.NameWithLowerFirst}}Form.Set{{.NameWithUpperFirst}}(&{{.NameWithLowerFirst}})
	return {{.NameWithLowerFirst}}Form
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "form.list.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
package {{.NameWithLowerFirst}}

{{.Imports}}

// The ListForm holds view data including a list of {{.PluralNameWithLowerFirst}}.  It's approximately equivalent 
// to a Struts form bean.
type ListForm interface {
	// {{.PluralNameWithUpperFirst}} returns the list of {{.NameWithUpperFirst}} objects from the form
	{{.PluralNameWithUpperFirst}}() []{{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}
	// Notice gets the notice.
	Notice() string
	// ErrorMessage gets the general error message.
	ErrorMessage() string
	// Set{{.PluralNameWithUpperFirst}} sets the list of {{.PluralNameWithLowerFirst}} in the form.
	Set{{.PluralNameWithUpperFirst}}([]{{.NameWithLowerFirst}}.{{.NameWithUpperFirst}})
	// SetNotice sets the notice.
	SetNotice(notice string)
	//SetErrorMessage sets the error message.
	SetErrorMessage(errorMessage string)
}`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "form.single.item.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
 package {{.NameWithLowerFirst}}

{{.Imports}}

// SingleItemForm holds view data about a {{.NameWithLowerFirst}}.  It's used as a data transfer object (DTO)
// in particular for use with views that handle a {{.NameWithUpperFirst}}.  (It's approximately equivalent to
// a Struts form bean.)  It contains a {{.NameWithUpperFirst}}; a validator function that validates the data
// in the {{.NameWithUpperFirst}} and sets the various error messages; a general error message (for errors not
// associated with an individual field of the {{.NameWithUpperFirst}}), a notice (for announcement that are
// not about errors) and a set of error messages about individual fields of the {{.NameWithUpperFirst}}.  It
// offers getters and setters for the various attributes that it supports.
type SingleItemForm interface {
	// {{.NameWithUpperFirst}} gets the {{.NameWithUpperFirst}} embedded in the form.
	{{.NameWithUpperFirst}}() {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}
	// Notice gets the notice.
	Notice() string
	// ErrorMessage gets the general error message.
	ErrorMessage() string
	// FieldErrors returns all the field errors as a map.
	FieldErrors() map[string]string
	// ErrorForField returns the error message about a field (may be an empty string).
	ErrorForField(key string) string
	// String returns a string version of the {{.NameWithUpperFirst}}Form.
	// Valid returns true if the contents of the form is valid
	Valid() bool
	String() string
	// Set{{.NameWithUpperFirst}} sets the {{.NameWithUpperFirst}} in the form.
	Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}} {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}})
	// SetNotice sets the notice.
	SetNotice(notice string)
	//SetErrorMessage sets the general error message.
	SetErrorMessage(errorMessage string)
	// SetErrorMessageForField sets the error message for a named field
	SetErrorMessageForField(fieldname, errormessage string)
	// SetValid sets a warning that the data in the form is invalid
	SetValid(value bool)
	// Validate validates the data in the {{.NameWithUpperFirst}} and sets the various error messages.
	// It returns true if the data is valid, false if there are errors.
	Validate() bool
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "gorp.concrete.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
package {{$resourceNameLower}}

{{.Imports}}

// Generated by the goblimey scaffold generator.  You are STRONGLY
// recommended not to alter this file, as it will be overwritten next time the 
// scaffolder is run.

// The Concrete{{$resourceNameUpper}} struct implements the {{$resourceNameUpper}} interface and holds a single row from
// the PEOPLE table, accessed via the GORP library.
//
// The fields must be public for GORP to work but the names must not clash with 
// those of the getters, so for a field "name" call the getter Name() and the 
// field NameField.
type Concrete{{$resourceNameUpper}} struct {
	IDField       uint64 %%GRAVE%%db: "id, primarykey, autoincrement"%%GRAVE%%
	{{range .Fields}}
	{{.NameWithUpperFirst}}Field {{.GoType}} %%GRAVE%%db: "{{.NameWithLowerFirst}}"%%GRAVE%%
	{{end}}
}

// Factory functions

// Make{{$resourceNameUpper}} creates and returns a new uninitialised {{$resourceNameUpper}} object
func Make{{$resourceNameUpper}}() {{$resourceNameLower}}.{{$resourceNameUpper}} {
	var Concrete{{$resourceNameUpper}} Concrete{{$resourceNameUpper}}
	return &Concrete{{$resourceNameUpper}}
}

// MakeInitialised{{$resourceNameUpper}} creates and returns a new {{$resourceNameUpper}} object initialised from
// the arguments
func MakeInitialised{{$resourceNameUpper}}(id uint64, {{range .Fields}}{{.NameWithLowerFirst}} {{.GoType}}{{if not .LastItem}}, {{end}}{{end}}) {{$resourceNameLower}}.{{$resourceNameUpper}} {
	{{$resourceNameLower}} := Make{{$resourceNameUpper}}()
	{{$resourceNameLower}}.SetID(id)
	{{range .Fields}}
		{{if eq .Type "string"}}
			{{$resourceNameLower}}.Set{{.NameWithUpperFirst}}(strings.TrimSpace({{.NameWithLowerFirst}}))
		{{else}}
			{{$resourceNameLower}}.Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}})
		{{end}}
	{{end}}
	return {{$resourceNameLower}}
}

// Clone creates and returns a new {{$resourceNameUpper}} object initialised from a source {{$resourceNameUpper}}.
func Clone({{$resourceNameLower}} {{$resourceNameLower}}.{{$resourceNameUpper}}) {{$resourceNameLower}}.{{$resourceNameUpper}} {
	return MakeInitialised{{$resourceNameUpper}}({{$resourceNameLower}}.ID(), {{range .Fields}}{{$resourceNameLower}}.{{.NameWithUpperFirst}}(){{if not .LastItem}}, {{end}}{{end}})
}

// Methods to implement the {{$resourceNameUpper}} interface.

// ID gets the id of the {{$resourceNameLower}}.
func (o Concrete{{$resourceNameUpper}}) ID() uint64 {
	return o.IDField
}
{{range .Fields}}
//{{.NameWithUpperFirst}} gets the {{.NameWithLowerFirst}} of the {{$resourceNameLower}}.
func (o Concrete{{$resourceNameUpper}}) {{.NameWithUpperFirst}}() {{.GoType}} {
	return o.{{.NameWithUpperFirst}}Field
}
{{end}}
// String gets the {{$resourceNameLower}} as a string.
func (o Concrete{{$resourceNameUpper}}) String() string {
	return fmt.Sprintf("Concrete{{$resourceNameUpper}}={id=%d, {{range .Fields}}{{.NameWithLowerFirst}}=%v{{if not .LastItem}}, {{end}}{{end}}{{"}"}}",
		o.IDField, {{range .Fields}}o.{{.NameWithUpperFirst}}Field{{if not .LastItem}}, {{end}}{{end}})		
}

// DisplayName returns a name for the object composed of the values of the id and 
// any fields not marked as excluded from the display name.
func (o Concrete{{$resourceNameUpper}}) DisplayName() string {
	return fmt.Sprintf("%d{{range .Fields}}{{if not .ExcludeFromDisplay}} %v{{end}}{{end}}",
		o.IDField{{range .Fields}}{{if not .ExcludeFromDisplay}}, o.{{.NameWithUpperFirst}}Field{{end}}{{end}})
}

// SetID sets the {{$resourceNameLower}}'s id to the given value
func (o *Concrete{{$resourceNameUpper}}) SetID(id uint64) {
	o.IDField = id
}
{{range .Fields}}
	{{if eq .Type "string"}}
		// Set{{.NameWithUpperFirst}} sets the {{.NameWithLowerFirst}} of the {{$resourceNameLower}}.
		func (o *Concrete{{$resourceNameUpper}}) Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}} {{.GoType}}) {
			o.{{.NameWithUpperFirst}}Field = strings.TrimSpace({{.NameWithLowerFirst}})
	{{else}}
		// Set{{.NameWithUpperFirst}} sets the {{.NameWithLowerFirst}} of the {{$resourceNameLower}}.
		func (o *Concrete{{$resourceNameUpper}}) Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}} {{.GoType}}) {
			o.{{.NameWithUpperFirst}}Field = {{.NameWithLowerFirst}}
	{{end}}
}
{{end}}

// Define the validation.
func (o *Concrete{{$resourceNameUpper}}) Validate() error {
	
	// Trim and test all mandatory string fields
	
	errorMessage := ""
	{{range .Fields}}
	    {{if and .Mandatory (eq .Type "string")}}
	        if len(strings.TrimSpace(o.{{.NameWithUpperFirst}}())) <= 0 {
				errorMessage += "you must specify the {{.NameWithLowerFirst}} "
			}
		{{end}}
	{{end}}
	if len(errorMessage) > 0 {
		return errors.New(errorMessage)
	}
	return nil
}`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "main.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
package main

{{.Imports}}

// resourceRE is the regular expression to extract the resource from the URI of
// the request to be.  For example in: "/people" and "/people/1/delete", the
// resource is "people".
var resourceRE = regexp.MustCompile(%%GRAVE%%^/([^/]+)%%GRAVE%%)

// The following regular expressions are for specific request URIs, to work out
// which controller method to call.  For example, a GET request with URI "/people"
// produces a call to the Index method of the people controller.
//
// The requests follow the REST model and therefore carry data such as IDs
// in the request URI rather than in HTTP parameters, for example:
//
//    GET /people/435
//
// rather than
//
//    GET/people&id=435
//
// Only form data is supplied through HTTP parameters

{{range .Resources}}
// The {{.NameWithLowerFirst}}DeleteRequestRE is the regular expression for the URI of a delete
// request containing a numeric ID - for example: "/{{.PluralNameWithLowerFirst}}/1/delete".
var {{.NameWithLowerFirst}}DeleteRequestRE = regexp.MustCompile(%%GRAVE%%^/{{.PluralNameWithLowerFirst}}/[0-9]+/delete$%%GRAVE%%)

// The {{.NameWithLowerFirst}}ShowRequestRE is the regular expression for the URI of a show
// request containing a numeric ID - for example: "/{{.PluralNameWithLowerFirst}}/1".
var {{.NameWithLowerFirst}}ShowRequestRE = regexp.MustCompile(%%GRAVE%%^/{{.PluralNameWithLowerFirst}}/[0-9]+$%%GRAVE%%)

// The {{.NameWithLowerFirst}}EditRequestRE is the regular expression for the URI of an edit
// request containing a numeric ID - for example: "/{{.PluralNameWithLowerFirst}}/1/edit".
var {{.NameWithLowerFirst}}EditRequestRE = regexp.MustCompile(%%GRAVE%%^/{{.PluralNameWithLowerFirst}}/[0-9]+/edit$%%GRAVE%%)

// The {{.NameWithLowerFirst}}UpdateRequestRE is the regular expression for the URI of an update
// request containing a numeric ID - for example: "/{{.PluralNameWithLowerFirst}}/1".  The URI
// is the same as for the show request, but we give it a different name for
// clarity.
var {{.NameWithLowerFirst}}UpdateRequestRE = {{.NameWithLowerFirst}}ShowRequestRE
{{end}}
var templateMap *map[string]map[string]retrofitTemplate.Template

// These values are set from the command line arguments.
var homeDir string // app server's home directory
var verbose bool   // verbose mode

func init() {
	const (
		defaultVerbose = false
		usage          = "enable verbose logging"
	)
	flag.BoolVar(&verbose, "verbose", defaultVerbose, usage)
	flag.BoolVar(&verbose, "v", defaultVerbose, usage+" (shorthand)")
	flag.StringVar(&homeDir, "homedir", "", "the application server's home directory (must contain the views directory)")
}

func main() {
	log.SetPrefix("main() ")
	// Find the home directory.  This is specified by the first command line
	// argument.  If that's not specified, the home is assumed to be the current
	//directory.
	// homeDir := flag.String("homedir", ".", "the home directory (should contain the views directory)")

	flag.Parse()
	log.Printf("args %s", strings.Join(os.Args, " "))
	if len(flag.Args()) >= 1 {
		homeDir = flag.Args()[0]
	}
	err := os.Chdir(homeDir)
	if err != nil {
		log.Printf("cannot change directory to homeDir %s - %s", homeDir,
			err.Error())
		os.Exit(-1)
	}

	// The home directory must contain a directory "views" containing the HTML and
	// the templates. If there is no views directory, give up.  Most likely, the
	// user has not moved to the right directory before running this.
	fileInfo, err := os.Stat("views")
	if err != nil {
		if os.IsNotExist(err) {
			// views does not exist
			em := "cannot find the views directory"
			log.Println(em)
			fmt.Fprintln(os.Stderr, em)

		} else if !fileInfo.IsDir() {
			// views exists but is not a directory
			em := "the file views must be a directory"
			log.Println(em)
			fmt.Fprintln(os.Stderr, em)

		} else {
			// some other error
			log.Println(err.Error())
			fmt.Fprintln(os.Stderr, err.Error())
		}

		os.Exit(-1)
	}

	templateMap = utilities.CreateTemplates()

	// Set up the restful web service.  Send all requests to marshall().

	log.Println("setting up routes")
	ws := new(restful.WebService)
	http.Handle("/stylesheets/", http.StripPrefix("/stylesheets/", http.FileServer(http.Dir("views/stylesheets"))))
	http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("views/html"))))
	// Handlers for static HTML pages.

	ws.Route(ws.GET("/").To(marshall))
	ws.Route(ws.GET("/error.html").To(marshall))
{{range .Resources}}
	// Tie all expected requests to the marshall.
	ws.Route(ws.GET("/{{.PluralNameWithLowerFirst}}").To(marshall))
	ws.Route(ws.GET("/{{.PluralNameWithLowerFirst}}/{id}/edit").To(marshall))
	ws.Route(ws.GET("/{{.PluralNameWithLowerFirst}}/{id}").To(marshall))
	ws.Route(ws.GET("/{{.PluralNameWithLowerFirst}}/create").To(marshall))
	ws.Route(ws.POST("/{{.PluralNameWithLowerFirst}}").Consumes("application/x-www-form-urlencoded").To(marshall))
	ws.Route(ws.POST("/{{.PluralNameWithLowerFirst}}/{id}").Consumes("application/x-www-form-urlencoded").To(marshall))
	ws.Route(ws.POST("/{{.PluralNameWithLowerFirst}}/{id}/delete").Consumes("application/x-www-form-urlencoded").To(marshall))
{{end}}
	restful.Add(ws)

	log.Println("starting the listener")
	err = http.ListenAndServe(":4000", nil)
	log.Printf("baling out - %s" + err.Error())
}

// marshall passes the request and response to the appropriate method of the
// appropriate  controller.
func marshall(request *restful.Request, response *restful.Response) {

	log.SetPrefix("main.marshall() ")

	defer catchPanic()
	
	// Create a service supplier
	var services services.ConcreteServices
	services.SetTemplates(templateMap)
{{range .Resources}}
	{{.NameWithUpperFirst}}Repository, err := {{.NameWithLowerFirst}}Repository.MakeRepository(verbose)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-1)
	}
	services.Set{{.NameWithUpperFirst}}Repository({{.NameWithUpperFirst}}Repository)
	defer {{.NameWithUpperFirst}}Repository.Close()
{{end}}

	// We get the HTTP request from the restful request via its public Request
	// attribute.  Getting the method from that requires another public attribute.
	// These operations and others cannot be defined using an interface, which is
	// why the request and response are passed to the controller as pointers to
	// concrete objects rather than as retro-fitted interfaces.

	uri := request.Request.URL.RequestURI()

	// The REST model uses HTTP requests such as PUT and DELETE.  The standard browsers do not support
	// these operations, so they are implemented using a POST request with a parameter "_method"
	// defining the operation.  (A post with a parameter "_method=PUT" simulates a PUT, and so on.)

	method := request.Request.Method
	if method == "POST" {
		// handle simulated PUT, DELETE etc via the _method parameter
		simMethod := request.Request.FormValue("_method")
		if simMethod == "PUT" || simMethod == "DELETE" {
			method = simMethod
		}
	}
	if verbose {
		log.Printf("uri %s method %s", uri, method)
	}
	
	// The home page "/" or "/index.html" is dealt with using the special resource 
	// "home".
	
	if uri == "/" || uri == "/index.html" {
		if verbose {
			log.Println("home page")
		}
		page := services.Template("home", "Index")
		if page == nil {
			log.Printf("no home Index page")
			utilities.Dead(response)
			return
		}
		// This template is just HTML, so it needs no data.
		err = page.Execute(response.ResponseWriter, nil)
		if err != nil {
			// Can't display the home index page.  Bale out.
			em := fmt.Sprintf("fatal error - failed to display error page for error %s\n", err.Error())
			log.Printf(em)
			panic(em)
		}
		return
	}
	
	// Extract the resource.  For uris "/people", /people/1 etc, the resource is
	// "people".  If the string matches the regular expression, result will have
	// at least two entries and the resource name will be in result[1]. 

	result := resourceRE.FindStringSubmatch(uri)

	if len(result) < 2 {
		em := fmt.Sprintf("illegal request uri %v", uri)
		log.Println(em)
		utilities.BadError(em, response)
		return
	}

	resource := result[1]

	switch resource {
{{range .Resources}}
	case "{{.PluralNameWithLowerFirst}}":

		if verbose {
			log.Printf("Sending request %s to {{.NameWithLowerFirst}} controller\n", uri)
		}
		
		var controller = {{.NameWithLowerFirst}}Controller.MakeController(&services, verbose)

		// Call the appropriate handler for the request

		switch method {

		case "GET":

			if uri == "/{{.PluralNameWithLowerFirst}}" {
				// "GET http://server:port/{{.PluralNameWithLowerFirst}}" - fetch all the valid {{.PluralNameWithLowerFirst}}
				// records and display them.
				form := services.Make{{.NameWithUpperFirst}}ListForm()
				controller.Index(request, response, form)

			} else if {{.NameWithLowerFirst}}EditRequestRE.MatchString(uri) {

				// "GET http://server:port/{{.PluralNameWithLowerFirst}}/1/edit" - fetch the {{.PluralNameWithLowerFirst}} record
				// given by the ID in the request and display the form to edit it.
				form := services.Make{{.NameWithUpperFirst}}Form()
				controller.Edit(request, response, form)


			} else if uri == "/{{.PluralNameWithLowerFirst}}/create" {

				// "GET http://server:port/{{.PluralNameWithLowerFirst}}/create" - display the form to
				// create a new single item {{.NameWithLowerFirst}} record.

				// Create an empty {{.NameWithLowerFirst}} to get started.
				{{.NameWithLowerFirst}} := services.Make{{.NameWithUpperFirst}}()
				form := services.MakeInitialised{{.NameWithUpperFirst}}Form({{.NameWithLowerFirst}})
				controller.New(request, response, form)


			} else if {{.NameWithLowerFirst}}ShowRequestRE.MatchString(uri) {

				// "GET http://server:port/{{.PluralNameWithLowerFirst}}/435" - fetch the {{.PluralNameWithLowerFirst}} record
				// with ID 435 and display it.
				
				// Get the ID from the HTML form data .  The data only contains the
				// ID so the resulting {{.NameWithUpperFirst}}Form may be marked 
				// as invalid, but we are only interested in the ID.
				form := makeValidated{{.NameWithUpperFirst}}FormFromRequest(request, &services)

				if form.{{.NameWithUpperFirst}}().ID() == 0 {
					// The ID in the HTML form data is junk.
					// This request is normally made from a link in a view.  The
					// link should always be correct, so this should never happen!
					em := fmt.Sprintf("illegal id")
					log.Println(em)
					controller.ErrorHandler(request, response, em)
				}
				
				controller.Show(request, response, form)

				
			} else {
				em := fmt.Sprintf("unexpected GET request - uri %v", uri)
				log.Println(em)
				controller.ErrorHandler(request, response, em)
			}

		case "PUT":
			if {{.NameWithLowerFirst}}UpdateRequestRE.MatchString(uri) {

				// POST http://server:port/{{.PluralNameWithLowerFirst}}/1" - update the single item {{.NameWithLowerFirst}} record with
				// the given ID from the URI using the form data in the body.
				form := makeValidated{{.NameWithUpperFirst}}FormFromRequest(request, &services)
				controller.Update(request, response, form)

			} else if uri == "/{{.PluralNameWithLowerFirst}}" {

				// POST http://server:port/{{.PluralNameWithLowerFirst}}" - create a new {{.PluralNameWithLowerFirst}} record from
				// the form data in the body.
				form := makeValidated{{.NameWithUpperFirst}}FormFromRequest(request, &services)
				controller.Create(request, response, form)

			} else {
				em := fmt.Sprintf("unexpected PUT request - uri %v", uri)
				log.Println(em)
				controller.ErrorHandler(request, response, em)
			}

		case "DELETE":
			if {{.NameWithLowerFirst}}DeleteRequestRE.MatchString(uri) {

				// "POST http://server:port/{{.PluralNameWithLowerFirst}}/1/delete" - delete the {{.PluralNameWithLowerFirst}}
				// record with the ID given in the request.
				
				// Get the ID from the HTML form data .  The data only contains the
				// ID so the resulting {{.NameWithUpperFirst}}Form may be marked 
				// as invalid, but we are only interested in the ID.
				form := makeValidated{{.NameWithUpperFirst}}FormFromRequest(request, &services)

				if form.{{.NameWithUpperFirst}}().ID() <= 0 {
					// The ID in the HTML form data is junk.
					// This request is normally made from a link in a view.  The
					// link should always be correct, so this should never happen!
					em := fmt.Sprintf("illegal id")
					log.Println(em)
					controller.ErrorHandler(request, response, em)
				}
				
				controller.Delete(request, response, form)

			}

		default:
			em := fmt.Sprintf("unexpected HTTP method %v", method)
			log.Println(em)
			controller.ErrorHandler(request, response, em)
		}
{{end}}
	default:
		em := fmt.Sprintf("unexpected resource %v in uri %v", resource, uri)
		log.Println(em)
		utilities.BadError(em, response)
	}
}

{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
{{range .Resources}}
// makeValidated{{.NameWithUpperFirst}}FormFromRequest gets the {{.NameWithLowerFirst}} data from the request, creates a
// {{.NameWithUpperFirst}} and returns it in a single item {{.NameWithLowerFirst}} form.
func makeValidated{{.NameWithUpperFirst}}FormFromRequest(req *restful.Request, services services.Services) {{.NameWithLowerFirst}}Forms.SingleItemForm {

	log.SetPrefix("makeValidated{{.NameWithUpperFirst}}FormFromRequest() ")

	{{.NameWithLowerFirst}} := services.Make{{.NameWithUpperFirst}}()
	{{.NameWithLowerFirst}}Form := services.MakeInitialised{{.NameWithUpperFirst}}Form({{.NameWithLowerFirst}})
	
	// The Validate method validates the {{.NameWithUpperFirst}}Form. For data
	// fields in the request that are destined for any object except a string
	// could also be invalid and we also have to check for that before we set
	// a field in the {{.NameWithUpperFirst}}Form.
	
	valid := true	// This will be set false on any error.
	
	err := req.Request.ParseForm()
	if err != nil {
		valid = false
		em := fmt.Sprintf("cannot parse form - %s", err.Error())
		log.Printf("%s\n", em)
		{{.NameWithLowerFirst}}Form.SetErrorMessage("Internal error while processing the last data input")
		// Cannot make any sense of the HTML form data - bale out.
		return {{.NameWithLowerFirst}}Form
	}
	
	// In some requests the ID is set.  If so, parse and copy it.
	var id uint64 = 0
	idStr := req.PathParameter("id")
	if idStr != "" {
		id, err = strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			valid = false
			em := fmt.Sprintf("invalid id %v in request - should be numeric", idStr)
			log.Printf("%s\n", em)
			{{.NameWithLowerFirst}}Form.SetErrorMessageForField("ID", "ID must be a whole number greater than 0")
		}
		{{.NameWithLowerFirst}}.SetID(id)
	}

	// Getting the value of a form field involves using the public Request
	// attribute of the restful request, and that operation cannot be
	// represented by an interface. This is why the request and response
	// are not represented as retro-fitted interfaces.
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
{{range .Fields}}
	{{if eq .GoType "string"}}
		{{.NameWithLowerFirst}} := req.Request.FormValue("{{.NameWithLowerFirst}}")
	{{else}}
		{{.NameWithLowerFirst}}Str := strings.TrimSpace(req.Request.FormValue("{{.NameWithLowerFirst}}"))
		{{if or (eq .GoType "int64") (eq .GoType "uint64")}}
			{{.NameWithLowerFirst}}, err := strconv.ParseInt({{.NameWithLowerFirst}}Str, 10, 64)
			if err != nil {
				valid = false
				log.Println(fmt.Sprintf("HTML for field {{.NameWithLowerFirst}} %s is not an integer", {{.NameWithLowerFirst}}Str))
				{{$resourceNameLower}}Form.SetErrorMessageForField("{{.NameWithUpperFirst}}", "must be a whole number")
			}
		{{else if eq .GoType "float64"}}
			{{.NameWithLowerFirst}}, err := strconv.ParseFloat({{.NameWithLowerFirst}}Str, 64)
			if err != nil {
				valid = false
				log.Println(fmt.Sprintf("HTML for field {{.NameWithLowerFirst}} %s is not a float value", {{.NameWithLowerFirst}}Str))
				{{$resourceNameLower}}Form.SetErrorMessageForField("{{.NameWithUpperFirst}}", "must be a number")
			}
		{{else if eq .GoType "bool"}}
			{{.NameWithLowerFirst}} := false
			if len({{.NameWithLowerFirst}}Str) > 0 {
				{{.NameWithLowerFirst}}, err = strconv.ParseBool({{.NameWithLowerFirst}}Str)
				if err != nil {
					valid = false
					log.Println(fmt.Sprintf("HTML for field {{.NameWithLowerFirst}} %s is not a boolean", {{.NameWithLowerFirst}}Str))
					{{$resourceNameLower}}Form.SetErrorMessageForField("{{.NameWithUpperFirst}}", "must be true or false")
				}
			}
		{{end}}
	{{end}}
	{{$resourceNameLower}}.Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}})
{{end}}
	if valid {
		// The HTML form data is valid so far - check the mandatory string fields.
		{{$resourceNameLower}}Form.SetValid({{.NameWithLowerFirst}}Form.Validate())
	} else {
		// Syntax errors in the HTML form data.  Validate the mandatory string
		// fields to set any remaining error messages, but set the form invalid 
		// anyway.
		{{$resourceNameLower}}Form.Validate()
		{{$resourceNameLower}}Form.SetValid(false)
	}
	return {{.NameWithLowerFirst}}Form
}
{{end}}

// Recover from any panic and log an error.
func catchPanic() {
	if p := recover(); p != nil {
		log.Printf("unrecoverable internal error %v\n", p)
	}
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "model.concrete.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
package {{$resourceNameLower}}

import (
	"errors"
	"fmt"
	"strings"
)

// Generated by the goblimey scaffold generator.  You are STRONGLY
// recommended not to alter this file, as it will be overwritten next time the 
// scaffolder is run.

// Concrete{{$resourceNameUpper}} represents a {{$resourceNameLower}} and satisfies the {{$resourceNameUpper}} interface.
type Concrete{{$resourceNameUpper}} struct {
	id       uint64
	{{range .Fields}}
		{{.NameWithLowerFirst}} {{.GoType}}
	{{end}}
}

// Define the factory functions.

// Make{{$resourceNameUpper}} creates and returns a new uninitialised {{$resourceNameUpper}} object
func Make{{$resourceNameUpper}}() {{$resourceNameUpper}} {
	var concrete{{$resourceNameUpper}} Concrete{{$resourceNameUpper}}
	return &concrete{{$resourceNameUpper}}
}

// MakeInitialised{{$resourceNameUpper}} creates and returns a new {{$resourceNameUpper}} object initialised from
// the arguments
func MakeInitialised{{$resourceNameUpper}}(id uint64, {{range .Fields}}{{.NameWithLowerFirst}} {{.GoType}}{{if not .LastItem}}, {{end}}{{end}}) {{$resourceNameUpper}} {
	{{$resourceNameLower}} := Make{{$resourceNameUpper}}()
	{{$resourceNameLower}}.SetID(id)
	{{range .Fields}}
		{{if eq .Type "string"}}
			{{$resourceNameLower}}.Set{{.NameWithUpperFirst}}(strings.TrimSpace({{.NameWithLowerFirst}}))
		{{else}}
			{{$resourceNameLower}}.Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}})
		{{end}}
	{{end}}
	return {{$resourceNameLower}}
}

// Clone creates and returns a new {{$resourceNameUpper}} object initialised from a source {{$resourceNameUpper}}.
func Clone({{$resourceNameLower}} {{$resourceNameUpper}}) {{$resourceNameUpper}} {
	return MakeInitialised{{$resourceNameUpper}}({{$resourceNameLower}}.ID(), {{range .Fields}}{{$resourceNameLower}}.{{.NameWithUpperFirst}}(){{if not .LastItem}}, {{end}}{{end}})
}

// Define the getters.

// ID() gets the id of the {{$resourceNameLower}}.
func (o Concrete{{$resourceNameUpper}}) ID() uint64 {
	return o.id
}
{{range .Fields}}
	//{{.NameWithUpperFirst}} gets the {{.NameWithLowerFirst}} of the {{$resourceNameLower}}.
	func (o Concrete{{$resourceNameUpper}}) {{.NameWithUpperFirst}}() {{.GoType}} {
		return o.{{.NameWithLowerFirst}}
	}
{{end}}
// String gets the {{$resourceNameLower}} as a string.
func (o Concrete{{$resourceNameUpper}}) String() string {
	return fmt.Sprintf("Concrete{{$resourceNameUpper}}={id=%d, {{range .Fields}}{{.NameWithLowerFirst}}=%v{{if not .LastItem}}, {{end}}{{end}}{{"}"}}",
		o.id, {{range .Fields}}o.{{.NameWithLowerFirst}}{{if not .LastItem}}, {{end}}{{end}})		
}
// DisplayName returns a name for the object composed of the values of the id and 
// the value of any field not marked as excluded.
func (o Concrete{{$resourceNameUpper}}) DisplayName() string {
	return fmt.Sprintf("%d{{range .Fields}}{{if not .ExcludeFromDisplay}} %v{{end}}{{end}}",
		o.id{{range .Fields}}{{if not .ExcludeFromDisplay}}, o.{{.NameWithLowerFirst}}{{end}}{{end}})
}

// Define the setters.

// SetID sets the id to the given value.
func (o *Concrete{{$resourceNameUpper}}) SetID(id uint64) {
	o.id = id
	}
	
{{range .Fields}}
	// Set{{.NameWithUpperFirst}} sets the {{.NameWithLowerFirst}} of the {{$resourceNameLower}}.
	func (o *Concrete{{$resourceNameUpper}}) Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}} {{.GoType}}) {
   	{{if eq .Type "string"}}
		o.{{.NameWithLowerFirst}} = strings.TrimSpace({{.NameWithLowerFirst}})
	{{else}}
		o.{{.NameWithLowerFirst}} = {{.NameWithLowerFirst}}
	{{end}}
	}
{{end}}

// Define the validation.
func (o *Concrete{{$resourceNameUpper}}) Validate() error {
	
	// Trim and test all mandatory string fields
	
	errorMessage := ""
	{{range .Fields}}
	    {{if and .Mandatory (eq .Type "string")}}
	        if len(strings.TrimSpace(o.{{.NameWithUpperFirst}}())) <= 0 {
				errorMessage += "you must specify the {{.NameWithLowerFirst}} "
			}
		{{end}}
	{{end}}
	if len(errorMessage) > 0 {
		return errors.New(errorMessage)
	}
	return nil
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "model.concrete.test.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
package {{.NameWithLowerFirst}}

import (
	"testing"
)

// Generated by the goblimey scaffold generator.  You are STRONGLY
// recommended not to alter this file, as it will be overwritten next time the 
// scaffolder is run.

// Unit tests for the Concrete{{$resourceNameUpper}} object.

{{/* This creates the expected values using the field name and the first test 
     value from each field, something like:
	 var expectedForename string = "s1"
	 var expectedSurname string = "s2"  */}}
{{range $index, $element := .Fields}}
	{{if eq .Type "string"}}
		var expected{{.NameWithUpperFirst}} {{.GoType}} = "{{index .TestValues 0}}"
	{{else}}
		var expected{{.NameWithUpperFirst}} {{.GoType}} = {{index .TestValues 0}}
	{{end}}
{{end}}
func TestUnitCreateConcrete{{$resourceNameUpper}}AndCheckContents(t *testing.T) {
	var expectedID uint64 = 42
	
	{{$resourceNameLower}} := MakeInitialised{{$resourceNameUpper}}(expectedID, {{range .Fields}}expected{{.NameWithUpperFirst}}{{if not .LastItem}}, {{end}}{{end}})
	if {{$resourceNameLower}}.ID() != expectedID {
		t.Errorf("expected ID to be %d actually %d", expectedID, {{$resourceNameLower}}.ID())
	}
	{{range .Fields}}
	if {{$resourceNameLower}}.{{.NameWithUpperFirst}}() != expected{{.NameWithUpperFirst}} {
		t.Errorf("expected {{.NameWithLowerFirst}} to be %s actually %s", expected{{.NameWithUpperFirst}}, {{$resourceNameLower}}.{{.NameWithUpperFirst}}())
	}
	{{end}}
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "model.interface.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
package {{$resourceNameLower}}

// Generated by the goblimey scaffold generator.  You are STRONGLY
// recommended not to alter this file, as it will be overwritten next time the 
// scaffolder is run.

// {{$resourceNameUpper}} represents a {{$resourceNameLower}}.
type {{$resourceNameUpper}} interface { 
	// ID() gets the id of the {{$resourceNameLower}}
	ID() uint64	
	{{range .Fields}}
		//{{.NameWithUpperFirst}} gets the {{.NameWithLowerFirst}} of the {{$resourceNameLower}}
		{{.NameWithUpperFirst}}() {{.GoType}} 
	{{end}}
	// String gets the {{$resourceNameLower}} as a string
	String() string
	// DisplayName gets a name composed of selected fields
	DisplayName() string
	// SetID sets the id to the given value
	SetID(id uint64)
	{{range .Fields}}
		// Set{{.NameWithUpperFirst}} sets the {{.NameWithLowerFirst}} of the {{$resourceNameLower}}
		Set{{.NameWithUpperFirst}}({{.NameWithLowerFirst}} {{.GoType}})
	{{end}}
	// Valdate checks the data in the {{.NameWithLowerFirst}}.
	Validate() error
}`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "repository.concrete.gorp.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
package gorp

{{.Imports}}

// Generated by the goblimey scaffold generator.  You are STRONGLY
// recommended not to alter this file, as it will be overwritten next time the 
// scaffolder is run.

// This package satisfies the {{.NameWithLowerFirst}} Repository interface and
// provides Create, Read, Update and Delete (CRUD) operations on the {{.PluralNameWithLowerFirst}} resource.
// In this case, the resource is a MySQL table accessed via GORP.

type GorpMysqlRepository struct {
	dbmap *gorp.DbMap
	verbose bool
}

// MakeRepository is a factory function that creates a GorpMysqlRepository and 
// returns it as a Repository.
func MakeRepository(verbose bool) ({{.NameWithLowerFirst}}Repo.Repository, error) {
	log.SetPrefix("{{.PluralNameWithLowerFirst}}.MakeRepository() ")

	db, err := sql.Open("{{.DB}}", "{{.DBLogin}}")
	if err != nil {
		log.Printf("failed to get DB handle - %s\n" + err.Error())
		return nil, errors.New("failed to get DB handle - " + err.Error())
	}
	// check that the handle works
	err = db.Ping()
	if err != nil {
		log.Printf("cannot connect to DB.  %s\n", err.Error())
		return nil, err
	}
	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	table := dbmap.AddTableWithName(gorp{{.NameWithUpperFirst}}.Concrete{{.NameWithUpperFirst}}{}, "{{.TableName}}").SetKeys(true, "IDField")
	if table == nil {
		em := "cannot add table {{.TableName}}"
		log.Println(em)
		return nil, errors.New(em)
	}

	table.ColMap("IDField").Rename("id")
	{{range .Fields}}
	table.ColMap("{{.NameWithUpperFirst}}Field").Rename("{{.NameWithLowerFirst}}")
	{{end}}
	// Create any missing tables.
	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		em := fmt.Sprintf("cannot create table - %s\n", err.Error())
		log.Printf("em")
		return nil, errors.New(em)
	}
	
	repository := GorpMysqlRepository{dbmap, verbose}
	return repository, nil
}

// SetVerbosity sets the verbosity level.
func (gmpd GorpMysqlRepository) SetVerbosity(verbose bool) {
	gmpd.verbose = verbose
}

// FindAll returns a list of all valid {{.NameWithUpperFirst}} records from the database in a slice.
// The result may be an empty slice.  If the database lookup fails, the error is
// returned instead.
func (gmpd GorpMysqlRepository) FindAll() ([]{{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}, error) {
	log.SetPrefix("FindAll() ")
	if gmpd.verbose {
		log.Println("")
	}

	transaction, err := gmpd.dbmap.Begin()
	if err != nil {
		em := fmt.Sprintf("cannot create transaction - %s", err.Error())
		log.Println(em)
		return nil, errors.New(em)
	}
	var {{.NameWithLowerFirst}}List []gorp{{.NameWithUpperFirst}}.Concrete{{.NameWithUpperFirst}}
	
	_, err = transaction.Select(&{{.NameWithLowerFirst}}List,
		"select id, {{range .Fields}}{{.NameWithLowerFirst}}{{if not .LastItem}}, {{end}}{{end}} from {{.TableName}}")
	if err != nil {
		transaction.Rollback()
		return nil, err
	}
	transaction.Commit()

	valid{{.PluralNameWithUpperFirst}} := make([]{{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}, len({{.NameWithLowerFirst}}List))

	// Validate and clone the {{.NameWithUpperFirst}} records

	next := 0 // Index of next valid{{.PluralNameWithUpperFirst}} entry
	for i, _ := range {{.NameWithLowerFirst}}List {
		// Check any mandatory string fields
		{{range .Fields}}
			{{if eq .Type "string" }}
				{{$resourceNameLower}}List[i].Set{{.NameWithUpperFirst}}(strings.TrimSpace({{$resourceNameLower}}List[i].{{.NameWithUpperFirst}}()))
			{{end}}
			{{if .Mandatory}}
				{{if eq .Type "string" }}
					if len({{$resourceNameLower}}List[i].{{.NameWithUpperFirst}}()) == 0 {
						continue
					}
				{{end}}
			{{end}}
		{{end}}
		
		// All mandatory string fields are set.  Clone the data.
		valid{{.PluralNameWithUpperFirst}}[next] = gorp{{.NameWithUpperFirst}}.Clone(&{{.NameWithLowerFirst}}List[i])
		next++
	}

	return valid{{.PluralNameWithUpperFirst}}, nil
}

// FindByID fetches the row from the {{.TableName}} table with the given uint64 id. It
// validates that data and, if it's valid, returns the {{.NameWithLowerFirst}}.  If the data is not
// valid the function returns an error message.
func (gmpd GorpMysqlRepository) FindByID(id uint64) ({{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}, error) {
	log.SetPrefix("FindByID() ")
	if gmpd.verbose{
		log.Printf("id=%d", id)
	}

	var {{.NameWithLowerFirst}} gorp{{.NameWithUpperFirst}}.Concrete{{.NameWithUpperFirst}}
	transaction, err := gmpd.dbmap.Begin()
	if err != nil {
		em := fmt.Sprintf("cannot create transaction - %s", err.Error())
		log.Println(em)
		return nil, errors.New(em)
	}

	err = transaction.SelectOne(&{{.NameWithLowerFirst}},
		"select id, {{range .Fields}}{{.NameWithLowerFirst}}{{if not .LastItem}}, {{end}}{{end}} from {{.TableName}} where id = ?", id)
	if err != nil {
		transaction.Rollback()
		log.Println(err.Error())
		return nil, err
	}
	transaction.Commit()
	if gmpd.verbose {
		log.Printf("found {{.NameWithLowerFirst}} %s", {{.NameWithLowerFirst}}.String())
	}
	
	if err != nil {
		return nil, err
	}
	{{range .Fields}}
		{{if eq .Type "string" }}
			if len(strings.TrimSpace({{$resourceNameLower}}.{{.NameWithUpperFirst}}())) < 1 {
				return nil, errors.New("invalid {{$resourceNameLower}} - {{.NameWithLowerFirst}} field must be set")
			}
		{{end}}
	{{end}}
	return &{{.NameWithLowerFirst}}, nil
}

// FindByIDStr fetches the row from the {{.TableName}} table with the given string id. It
// validates that data and, if it's valid, returns the {{.NameWithLowerFirst}}.  If the data is not valid
// the function returns an errormessage.  The ID in the database is numeric and the method
// checks that the given ID is also numeric before it makes the call.  This avoids hitting
// the DB when the id is obviously junk.
func (gmpd GorpMysqlRepository) FindByIDStr(idStr string) ({{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}, error) {
	log.SetPrefix("FindByIDStr() ")
	if gmpd.verbose {
		log.Printf("id=%s", idStr)
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		em := fmt.Sprintf("ID %s is not an unsigned integer", idStr)
		log.Println(em)
		return nil, fmt.Errorf("ID %s is not an unsigned integer", idStr)
	}
	return gmpd.FindByID(id)
}

// Create takes a {{.NameWithLowerFirst}}, creates a record in the {{.TableName}} table containing the same
// data with an auto-incremented ID and returns any error that the DB call returns.
// On a successful create, the method returns the created {{.NameWithLowerFirst}}, including
// the assigned ID.  This is all done within a transaction to ensure atomicity.
func (gmpd GorpMysqlRepository) Create({{.NameWithLowerFirst}} {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}) ({{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}, error) {
	log.SetPrefix("Create() ")
	if gmpd.verbose {
		log.Println("")
	}

	tx, err := gmpd.dbmap.Begin()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	{{.NameWithLowerFirst}}.SetID(0) // provokes the auto-increment
	err = tx.Insert({{.NameWithLowerFirst}})
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if gmpd.verbose {
		log.Printf("created {{.NameWithLowerFirst}} %s", {{.NameWithLowerFirst}}.String())
	}
	return {{.NameWithLowerFirst}}, nil
}

// Update takes a {{.NameWithLowerFirst}} record, updates the record in the {{.TableName}} table with the same ID
// and returns the updated {{.NameWithLowerFirst}} or any error that the DB call supplies to it.  The update
// is done within a transaction
func (gmpd GorpMysqlRepository) Update({{.NameWithLowerFirst}} {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}) (uint64, error) {
	log.SetPrefix("Update() ")

	tx, err := gmpd.dbmap.Begin()
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	rowsUpdated, err := tx.Update({{.NameWithLowerFirst}})
	if err != nil {
		tx.Rollback()
		log.Println(err.Error())
		return 0, err
	}
	if rowsUpdated != 1 {
		tx.Rollback()
		em := fmt.Sprintf("update failed - %d rows would have been updated, expected 1", rowsUpdated)
		log.Println(em)
		return 0, errors.New(em)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Println(err.Error())
		return 0, err
	}

	// Success!
	return 1, nil
}

// DeleteByID takes the given uint64 ID and deletes the record with that ID from the {{.TableName}} table.
// The function returns the row count and error that the database supplies to it.  On a successful
// delete, it should return 1, having deleted one row.
func (gmpd GorpMysqlRepository) DeleteByID(id uint64) (int64, error) {
	log.SetPrefix("DeleteByID() ")
	
	if gmpd.verbose {
		log.Printf("id=%d", id)
	}

	// Need a {{.NameWithUpperFirst}} record for the delete method, so fake one up.
	var {{.NameWithLowerFirst}} gorp{{.NameWithUpperFirst}}.Concrete{{.NameWithUpperFirst}}
	{{.NameWithLowerFirst}}.SetID(id)
	tx, err := gmpd.dbmap.Begin()
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	rowsDeleted, err := tx.Delete(&{{.NameWithLowerFirst}})
	if err != nil {
		tx.Rollback()
		log.Println(err.Error())
		return 0, err
	}
	if rowsDeleted != 1 {
		tx.Rollback()
		em := fmt.Sprintf("delete failed - %d rows would have been deleted, expected 1", rowsDeleted)
		log.Println(em)
		return 0, errors.New(em)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Println(err.Error())
		return 0, err
	}
	if err != nil {
		log.Println(err.Error())
	}
	return rowsDeleted, nil
}

// DeleteByIDStr takes the given String ID and deletes the record with that ID from the {{.TableName}} table.
// The ID in the database is numeric and the method checks that the given ID is also numeric before
// it makes the call.  If not, it returns an error.  If the ID looks sensible, the function attempts
// the delete and returns the row count and error that the database supplies to it.  On a successful
// delete, it should return 1, having deleted one row.
func (gmpd GorpMysqlRepository) DeleteByIDStr(idStr string) (int64, error) {
	log.SetPrefix("DeleteByIDStr() ")
	if gmpd.verbose {
		log.Printf("ID %s", idStr)
	}
	// Check the id.
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		em := fmt.Sprintf("ID %s is not an unsigned integer", idStr)
		log.Println(em)
		return 0, errors.New(em)
	}
	return gmpd.DeleteByID(id)
}

// Close closes the repository, reclaiming any redundant resources, in
// particular, any open database connection and transactions.  Anything that
// creates a repository MUST call this when it's finished, to avoid resource 
// leaks.
func (gmpd GorpMysqlRepository) Close() {
	log.SetPrefix("Close() ")
	if gmpd.verbose {	
		log.Printf("closing the {{.NameWithLowerFirst}} repository")
	}
	gmpd.dbmap.Db.Close()
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "repository.concrete.gorp.test.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
package gorp

{{.Imports}}

// This is an integration test for the Gorp mysql {{.NameWithLowerFirst}} repository.

{{/* This creates the expected values using the field names and the test 
     values, something like:
	 var expectedName1 string = "s1"
	 var expectedAge1 int = 2 
	 var expectedName2 string = "s3"
	 var expectedAge2 int = 4 */}}
{{range $index, $element := .Fields}}
	{{if eq .Type "string"}}
		var expected{{.NameWithUpperFirst}}1 {{.GoType}} = "{{index .TestValues 0}}"
	{{else}}
		var expected{{.NameWithUpperFirst}}1 {{.GoType}} = {{index .TestValues 0}}
	{{end}}
	{{if eq .Type "string"}}
		var expected{{.NameWithUpperFirst}}2 {{.GoType}} = "{{index .TestValues 1}}"
	{{else}}
		var expected{{.NameWithUpperFirst}}2 {{.GoType}} = {{index .TestValues 1}}
	{{end}}
{{end}}

// Create a {{.NameWithLowerFirst}} in the database, read it back, test the contents.
func TestIntCreate{{.NameWithUpperFirst}}StoreFetchBackAndCheckContents(t *testing.T) {
	log.SetPrefix("TestIntegrationegrationCreate{{.NameWithUpperFirst}}AndCheckContents")

	// Create a GORP {{.PluralNameWithLowerFirst}} repository
	repository, err := MakeRepository(false)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-1)
	}
	defer repository.Close()

	clearDown(repository, t)
	
	o := gorp{{.NameWithUpperFirst}}.MakeInitialised{{.NameWithUpperFirst}}(0, {{range .Fields}}expected{{.NameWithUpperFirst}}1{{if not .LastItem}}, {{end}}{{end}})
	{{.NameWithLowerFirst}}, err := repository.Create(o)
	if err != nil {
		t.Errorf(err.Error())
	}

	retrieved{{.NameWithUpperFirst}}, err := repository.FindByID({{.NameWithLowerFirst}}.ID())
	if err != nil {
		t.Errorf(err.Error())
	}

	if retrieved{{.NameWithUpperFirst}}.ID() != {{.NameWithLowerFirst}}.ID() {
		t.Errorf("expected ID to be %d actually %d", {{.NameWithLowerFirst}}.ID(),
			retrieved{{.NameWithUpperFirst}}.ID())
	}
	{{range .Fields}}
		if retrieved{{$resourceNameUpper}}.{{.NameWithUpperFirst}}() != expected{{.NameWithUpperFirst}}1 {
			t.Errorf("expected {{.NameWithLowerFirst}} to be %s actually %s", expected{{.NameWithUpperFirst}}1, {{$resourceNameLower}}.{{.NameWithUpperFirst}}())
	}
	{{end}}

	// Delete {{.NameWithLowerFirst}} and check response
	rows, err := repository.DeleteByID(retrieved{{.NameWithUpperFirst}}.ID())
	if err != nil {
		t.Errorf(err.Error())
	}
	if rows != 1 {
		t.Errorf("expected delete to return 1, actual %d", rows)
	}
	clearDown(repository, t)
}

// Create two {{.NameWithLowerFirst}} records in the DB, read them back and check the fields
func TestIntCreateTwo{{.PluralNameWithUpperFirst}}AndReadBack(t *testing.T) {
	log.SetPrefix("TestCreate{{.NameWithUpperFirst}}AndReadBack")

	// Create a GORP {{.PluralNameWithLowerFirst}} repository
	repository, err := MakeRepository(false)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-1)
	}
	defer repository.Close()

	clearDown(repository, t)

	//Create two {{.PluralNameWithLowerFirst}}

	o1 := gorp{{.NameWithUpperFirst}}.MakeInitialised{{.NameWithUpperFirst}}(0, {{range .Fields}}expected{{.NameWithUpperFirst}}1{{if not .LastItem}}, {{end}}{{end}})
	{{.NameWithLowerFirst}}1, err := repository.Create(o1)
	if err != nil {
		t.Errorf(err.Error())
	}

	o2 := gorp{{.NameWithUpperFirst}}.MakeInitialised{{.NameWithUpperFirst}}(0, {{range .Fields}}expected{{.NameWithUpperFirst}}2{{if not .LastItem}}, {{end}}{{end}})
	{{.NameWithLowerFirst}}2, err := repository.Create(o2)
	if err != nil {
		t.Errorf(err.Error())
	}

	// read all the {{.PluralNameWithLowerFirst}} in the DB - expect just the two we created
	{{.PluralNameWithLowerFirst}}, err := repository.FindAll()
	if err != nil {
		t.Errorf(err.Error())
	}

	if len({{.PluralNameWithLowerFirst}}) != 2 {
		t.Errorf("expected 2 rows, actual %d", len({{.PluralNameWithLowerFirst}}))
	}

	for _, {{.NameWithLowerFirst}} := range {{.PluralNameWithLowerFirst}} {
		
		matches := 1
		
		{{/* Check that the fields of each are consistent with the source object.
		     (Note: we don't know what in order the two objects will come back.) */}}

			{{$firstField := index .Fields 0}}
			{{$firstFieldName := $firstField.NameWithUpperFirst}}
			switch {{.NameWithLowerFirst}}.{{$firstFieldName}}() {
			case expected{{$firstFieldName}}1:
				{{range .Fields}}
					{{if ne $firstFieldName .NameWithUpperFirst}}
						if {{$resourceNameLower}}.{{.NameWithUpperFirst}}() == expected{{.NameWithUpperFirst}}1 {
							matches++
						} else {
							t.Errorf("expected {{.NameWithLowerFirst}} to be %s actually %s", 
								expected{{.NameWithUpperFirst}}1, {{$resourceNameLower}}.{{.NameWithUpperFirst}}())
						}
					{{end}}
				{{end}}
			case expected{{$firstFieldName}}2:
				{{range .Fields}}
					{{if ne $firstFieldName .NameWithUpperFirst}}
						if {{$resourceNameLower}}.{{.NameWithUpperFirst}}() == expected{{.NameWithUpperFirst}}2 {
							matches++
						} else {
							t.Errorf("expected {{.NameWithLowerFirst}} to be %s actually %s", 
								expected{{.NameWithUpperFirst}}2, {{$resourceNameLower}}.{{.NameWithUpperFirst}}())
						}
					{{end}}
				{{end}}
			default:
				t.Errorf("unexpected {{.NameWithLowerFirst}} with name %s - expected %s or %s", 
					{{$resourceNameLower}}.{{$firstFieldName}}(), expected{{$firstFieldName}}1, expected{{$firstFieldName}}2)
			}
			
			// We should have one match for each field
			if matches != {{len .Fields}} {
				t.Errorf("expected %d fields, actual %d", {{len .Fields}}, matches)
			}
	}

	
	// Find the first {{.NameWithLowerFirst}} by numeric ID and check the fields
	{{.NameWithLowerFirst}}1Returned, err := repository.FindByID({{.NameWithLowerFirst}}1.ID())
	if err != nil {
		t.Errorf(err.Error())
	}

	{{range .Fields}}
		if {{$resourceNameLower}}1Returned.{{.NameWithUpperFirst}}() != expected{{.NameWithUpperFirst}}1 {
			t.Errorf("expected {{.NameWithLowerFirst}} to be %s actually %s",
				expected{{.NameWithUpperFirst}}1, {{$resourceNameLower}}1Returned.{{.NameWithUpperFirst}}())
		}
	{{end}}

	// Find the second {{.NameWithLowerFirst}} by string ID and check the fields
	IDStr := strconv.FormatUint({{.NameWithLowerFirst}}2.ID(), 10)
	{{.NameWithLowerFirst}}2Returned, err := repository.FindByIDStr(IDStr)
	if err != nil {
		t.Errorf(err.Error())
	}
	
	{{range .Fields}}
		if {{$resourceNameLower}}2Returned.{{.NameWithUpperFirst}}() != expected{{.NameWithUpperFirst}}2 {
			t.Errorf("expected {{.NameWithLowerFirst}} to be %s actually %s",
				expected{{.NameWithUpperFirst}}2, {{$resourceNameLower}}2Returned.{{.NameWithUpperFirst}}())
		}
	{{end}}

	clearDown(repository, t)
}

// Create two {{.PluralNameWithUpperFirst}}, remove one, check that we get back just the other
func TestIntCreateTwo{{.PluralNameWithUpperFirst}}AndDeleteOneByIDStr(t *testing.T) {
	log.SetPrefix("TestIntegrationegrationCreateTwoPeopleAndDeleteOneByIDStr")

	// Create a GORP {{.PluralNameWithLowerFirst}} repository
	repository, err := MakeRepository(false)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-1)
	}
	defer repository.Close()

	clearDown(repository, t)

	// Create two {{.PluralNameWithLowerFirst}}
	o1 := gorp{{.NameWithUpperFirst}}.MakeInitialised{{.NameWithUpperFirst}}(0, {{range .Fields}}expected{{.NameWithUpperFirst}}1{{if not .LastItem}}, {{end}}{{end}})
	{{.NameWithLowerFirst}}1, err := repository.Create(o1)
	if err != nil {
		t.Errorf(err.Error())
	}

	o2 := gorp{{.NameWithUpperFirst}}.MakeInitialised{{.NameWithUpperFirst}}(0, {{range .Fields}}expected{{.NameWithUpperFirst}}2{{if not .LastItem}}, {{end}}{{end}})
	{{.NameWithLowerFirst}}2, err := repository.Create(o2)
	if err != nil {
		t.Errorf(err.Error())
	}

	var IDStr = fmt.Sprintf("%d", {{.NameWithLowerFirst}}1.ID())
	rows, err := repository.DeleteByIDStr(IDStr)
	if err != nil {
		t.Errorf(err.Error())
	}
	if rows != 1 {
		t.Errorf("expected one record to be deleted, actually %d", rows)
	}

	// We should have one record in the DB and it should match {{.NameWithLowerFirst}}2
	{{.PluralNameWithLowerFirst}}, err := repository.FindAll()
	if err != nil {
		t.Errorf(err.Error())
	}

	if len({{.PluralNameWithLowerFirst}}) != 1 {
		t.Errorf("expected one record, actual %d", len({{.PluralNameWithLowerFirst}}))
	}

	if {{.PluralNameWithLowerFirst}}[0].ID() != {{.NameWithLowerFirst}}2.ID() {
		t.Errorf("expected id to be %d actually %d",
			{{.NameWithLowerFirst}}2.ID(), {{.PluralNameWithLowerFirst}}[0].ID())
	}
	{{$name := .PluralNameWithLowerFirst}}
	{{range .Fields}}
		if {{$name}}[0].{{.NameWithUpperFirst}}() != expected{{.NameWithUpperFirst}}2 {
			t.Errorf("expected {{.NameWithLowerFirst}} to be %s actually %s",
				expected{{.NameWithUpperFirst}}2, {{$name}}[0].{{.NameWithUpperFirst}}())
		}
	{{end}}

	clearDown(repository, t)
}

// Create a {{.NameWithLowerFirst}} record, update the record, read it back and check the updated values.
func TestIntCreate{{.NameWithUpperFirst}}AndUpdate(t *testing.T) {
	log.SetPrefix("TestIntCreate{{.NameWithUpperFirst}}AndUpdate")

	// Create a GORP {{.PluralNameWithLowerFirst}} repository
	repository, err := MakeRepository(false)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-1)
	}
	defer repository.Close()

	clearDown(repository, t)

	// Create a {{.NameWithLowerFirst}} in the DB.
	o := gorp{{.NameWithUpperFirst}}.MakeInitialised{{.NameWithUpperFirst}}(0, {{range .Fields}}expected{{.NameWithUpperFirst}}1{{if not .LastItem}}, {{end}}{{end}})
	{{.NameWithLowerFirst}}, err := repository.Create(o)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Update the {{.NameWithLowerFirst}} in the DB.
	{{range .Fields}}
		{{$resourceNameLower}}.Set{{.NameWithUpperFirst}}(expected{{.NameWithUpperFirst}}2)
	{{end}}
	rows, err := repository.Update({{.NameWithLowerFirst}})
	if err != nil {
		t.Errorf(err.Error())
	}
	if rows != 1 {
		t.Errorf("expected 1 row to be updated, actually %d rows", rows)
	}

	// fetch the updated record back and check it.
	retrieved{{.NameWithUpperFirst}}, err := repository.FindByID({{.NameWithLowerFirst}}.ID())
	if err != nil {
		t.Errorf(err.Error())
	}

	{{range .Fields}}
		if retrieved{{$resourceNameUpper}}.{{.NameWithUpperFirst}}() != expected{{.NameWithUpperFirst}}2 {
			t.Errorf("expected {{.NameWithLowerFirst}} to be %s actually %s",
				expected{{.NameWithUpperFirst}}2, retrieved{{$resourceNameUpper}}.{{.NameWithUpperFirst}}())
		}
	{{end}}

	clearDown(repository, t)
}

// clearDown() - helper function to remove all {{.PluralNameWithLowerFirst}} from the DB
func clearDown(repository {{.NameWithLowerFirst}}.Repository, t *testing.T) {
	{{.PluralNameWithLowerFirst}}, err := repository.FindAll()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	for _, {{.NameWithLowerFirst}} := range {{.PluralNameWithLowerFirst}} {
		rows, err := repository.DeleteByID({{.NameWithLowerFirst}}.ID())
		if err != nil {
			t.Errorf(err.Error())
			continue
		}
		if rows != 1 {
			t.Errorf("while clearing down, expected 1 row, actual %d", rows)
		}
	}
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "repository.interface.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
package {{.NameWithLowerFirst}}

{{.Imports}}

// Generated by the goblimey scaffold generator.  You are STRONGLY
// recommended not to alter this file, as it will be overwritten next time the 
// scaffolder is run.

// Repository is the interface defining a repository (AKA a Data Access Object) for
// the {{.TableName}} table.
type Repository interface {

	// FindAll() returns a pointer to a slice of valid {{.PluralNameWithUpperFirst}} 
	// records.  Any invalid records are left out of the slice (so it may be empty).
	FindAll() ([]{{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}, error)

	// FindByid fetches the row from the {{.TableName}} table with the given uint64 
	// id and validates the data.  If the data is valid, the method creates a new
	// {{.NameWithUpperFirst}} record and returns a pointer to the version in memory.  
	// If the data is not valid the method returns an error message.
	FindByID(id uint64) ({{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}, error)

	// FindByid fetches the row from the {{.TableName}} table with the given string 
	// id and validates the data.  If it's valid the method creates a {{.NameWithUpperFirst}} 
	// object and returns a pointer to it.  If the data is not valid the function 
	// returns an error message.
	//
	// The ID in the database is always numeric so the method first checks that the 
	// given ID is numeric before making the DB call, returning an error if it's not.
	FindByIDStr(idStr string) ({{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}, error)

	// Create takes a {{.NameWithLowerFirst}} and creates a record in the {{.TableName}}
	// table containing the same data plus an auto-incremented ID.  It returns a 
	// pointer to the resulting {{.NameWithLowerFirst}} object, or any error that 
	// the DB call supplies to it.
	Create({{.NameWithLowerFirst}} {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}) ({{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}, error)

	// Update takes a {{.NameWithLowerFirst}} object, validates it and, if it's
	// valid, searches the {{.TableName}} table for a record with a matching ID and 
	// updates it.  It returns the number of rows affected or any error from the
	// DB update call.  On a successful update, it should return 1, having updated 
	// one row.
	Update({{.NameWithLowerFirst}} {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}) (uint64, error)

	// DeleteById takes the given uint64 ID and deletes the record with that ID 
	// from the {{.TableName}} table.  It return the count of rows affected or any
	// error from the DB delete call.  On a successful delete, it should return 1, 
	// having deleted one row.
	DeleteByID(id uint64) (int64, error)

	// DeleteByIdStr takes the given String ID and deletes the record with that ID 
	// from the {{.TableName}} table.  The IDs in the database are numeric aso the 
	// method checks that the given ID is also numeric before it makes the DB call
	// and returns an error if not.  If the ID looks sensible, the method attempts 
	// the delete and returns the number of rows affected or any error from the
	// DB delete call. On a successful delete, it should return 1, having deleted 
	// one row.

	DeleteByIDStr(idStr string) (int64, error)

	// Close closes the repository, reclaiming any redundant resources, in
	// particular, any open database connection and transactions.  Anything that
	// creates a repository MUST call this when it's finished, to avoid resource 
	// leaks.
	Close()
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "retrofit.template.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
package template

import (
	"io"
)

// The Template interface mimics some of the the html/Template functionality,
// allowing templates to be mocked.
type Template interface {
	// Execute executes the template
	Execute(wr io.Writer, data interface{}) error
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "services.concrete.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
package services

{{.Imports}}

type ConcreteServices struct {
	{{range .Resources}}
		{{.NameWithLowerFirst}}Repo  {{.NameWithLowerFirst}}Repo.Repository
	{{end}}
	templateMap *map[string]map[string]retrofitTemplate.Template
}

// Template returns an HTML template, given a resource and a CRUD operation (Index,
// Edit etc).
func (cs ConcreteServices) Template(resource string, operation string) retrofitTemplate.Template {
	return (*cs.templateMap)[resource][operation]
}

// SetTemplates sets all HTML templates from the given map.
func (cs *ConcreteServices) SetTemplates(templateMap *map[string]map[string]retrofitTemplate.Template) {
	cs.templateMap = templateMap
}

// SetTemplate sets the HTML template for the resource and operation.
func (cs *ConcreteServices) SetTemplate(resource string, operation string,
	template retrofitTemplate.Template) {

	if (*cs.templateMap)[resource] == nil {
		// New row.
		(*cs.templateMap)[resource] = make(map[string]retrofitTemplate.Template)
	}

	(*cs.templateMap)[resource][operation] = template
}

{{range .Resources}}
	{{$resourceNameLower := .NameWithLowerFirst}}
	{{$resourceNameUpper := .NameWithUpperFirst}}
	// {{.NameWithUpperFirst}}Repository gets the {{.NameWithLowerFirst}} repository.
	func (cs ConcreteServices) {{.NameWithUpperFirst}}Repository() {{.NameWithLowerFirst}}Repo.Repository {
		return cs.{{.NameWithLowerFirst}}Repo
	}
	
	// Set{{.NameWithUpperFirst}}Repository sets the {{.NameWithLowerFirst}} repository.
	func (cs *ConcreteServices) Set{{.NameWithUpperFirst}}Repository(repo {{.NameWithLowerFirst}}Repo.Repository) {
		cs.{{.NameWithLowerFirst}}Repo = repo
	}

	// Make{{.NameWithUpperFirst}} creates and returns a new uninitialised {{.NameWithLowerFirst}} object, made by the
	// GORP Make{{.NameWithUpperFirst}}.
	func (cs *ConcreteServices) Make{{.NameWithUpperFirst}}() {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}} {
		return gorp{{.NameWithUpperFirst}}.Make{{.NameWithUpperFirst}}()
	}
	
	// MakeInitialised{{.NameWithUpperFirst}} creates and returns a new {{.NameWithUpperFirst}} object initialised from
	// the arguments and created using the GORP MakeInitialised{{.NameWithUpperFirst}}.
	func (cs *ConcreteServices) MakeInitialised{{.NameWithUpperFirst}}(id uint64, {{range .Fields}}{{.NameWithLowerFirst}} {{.GoType}}{{if not .LastItem}}, {{end}}{{end}}) {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}} {
		return gorp{{.NameWithUpperFirst}}.MakeInitialised{{.NameWithUpperFirst}}(id, {{range .Fields}}{{.NameWithLowerFirst}}{{if not .LastItem}}, {{end}}{{end}})
	}
	
	// Clone{{.NameWithUpperFirst}} creates and returns a new {{.NameWithUpperFirst}} object initialised from a source {{.NameWithUpperFirst}}.
	// The copy is made using the GORP Clone.
	func (cs *ConcreteServices) Clone{{.NameWithUpperFirst}}(source {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}) {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}} {
		return gorp{{.NameWithUpperFirst}}.Clone(source)
	}
	
	// Make{{.NameWithUpperFirst}}Form creates and returns an uninitialised {{.NameWithLowerFirst}} form.
	func (cs *ConcreteServices) Make{{.NameWithUpperFirst}}Form() {{.NameWithLowerFirst}}Forms.SingleItemForm {
		return {{.NameWithLowerFirst}}Forms.MakeSingleItemForm()
	}
	
	// MakeInitialised{{.NameWithUpperFirst}}Form creates a GORP {{.NameWithLowerFirst}} form containing the given
	// {{.NameWithLowerFirst}} and returns it as a {{.NameWithUpperFirst}}Form.
	func (cs *ConcreteServices) MakeInitialised{{.NameWithUpperFirst}}Form({{.NameWithLowerFirst}} {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}) {{.NameWithLowerFirst}}Forms.SingleItemForm {
		return {{.NameWithLowerFirst}}Forms.MakeInitialisedSingleItemForm({{.NameWithLowerFirst}})
	}
	
	// MakeListForm creates and returns a new uninitialised {{.NameWithLowerFirst}} ListForm
	// object as a ListForm.
	func (cs *ConcreteServices) Make{{.NameWithUpperFirst}}ListForm() {{.NameWithLowerFirst}}Forms.ListForm {
		return {{.NameWithLowerFirst}}Forms.MakeListForm()
	}
{{end}}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "services.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
package services

{{.Imports}}

type Services interface {
	
	// Template gets the HTML template named by the resource and operation.
	Template(resource string, operation string) retrofitTemplate.Template

	// SetTemplate sets the HTML template for the resource and operation.
	SetTemplate(resource string, operation string, template retrofitTemplate.Template)

	// SetTemplates sets all HTML templates from the given map
	SetTemplates(templateMap *map[string]map[string]retrofitTemplate.Template)

{{range .Resources}}
	{{$resourceNameLower := .NameWithLowerFirst}}
	{{$resourceNameUpper := .NameWithUpperFirst}}
	// {{.NameWithUpperFirst}}Repository returns the {{.NameWithLowerFirst}} repository.
	{{.NameWithUpperFirst}}Repository() {{.NameWithLowerFirst}}Repo.Repository

	// Set{{.NameWithUpperFirst}}Repository sets the {{.NameWithLowerFirst}} repository.
	Set{{.NameWithUpperFirst}}Repository(repository {{.NameWithLowerFirst}}Repo.Repository)
	
	// Make{{.NameWithUpperFirst}} creates and returns a new uninitialised {{.NameWithUpperFirst}} object.
	Make{{.NameWithUpperFirst}}() {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}

	// MakeInitialised{{.NameWithUpperFirst}} creates and returns a new {{.NameWithUpperFirst}} object initialised from
	// the arguments.
	MakeInitialised{{.NameWithUpperFirst}}(id uint64, {{range .Fields}}{{.NameWithLowerFirst}} {{.GoType}}{{if not .LastItem}}, {{end}}{{end}}) {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}

	// Clone{{.NameWithUpperFirst}} creates and returns a new {{.NameWithUpperFirst}} object initialised from a source {{.NameWithUpperFirst}}.
	Clone{{.NameWithUpperFirst}}(source {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}) {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}

	// Make{{.NameWithUpperFirst}}Form creates and returns an uninitialised {{.NameWithLowerFirst}} form.
	Make{{.NameWithUpperFirst}}Form() {{.NameWithLowerFirst}}Forms.SingleItemForm

	// MakeInitialised{{.NameWithUpperFirst}}Form creates and returns a {{.NameWithLowerFirst}} form containing the
	// given {{.NameWithLowerFirst}} object.
	MakeInitialised{{.NameWithUpperFirst}}Form({{.NameWithLowerFirst}} {{.NameWithLowerFirst}}.{{.NameWithUpperFirst}}) {{.NameWithLowerFirst}}Forms.SingleItemForm

	// Make{{.NameWithUpperFirst}}ListForm creates a new uninitialised {{.NameWithLowerFirst}} ConcreteListForm object and
	// returns it as a ListForm.
	Make{{.NameWithUpperFirst}}ListForm() {{.NameWithLowerFirst}}Forms.ListForm
{{end}}
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "sh.build.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
#! /bin/sh

# Script to build the {{.NameAllUpper}} application server.
#
# The script is generated the first time you run the Goblimey scaffolder.  If you
# need to recreate it, run the scaffolder with the -overwrite option.
# 
# To buld the application, change directory to the one containing this file and 
# run it, for example:
#
#    cd $HOME/workspaces/{{.Name}}
#    . build.sh
#
# The script assumes that the scaffolder and the go tools are available via the
# PATH.

if test -f setenv.sh
then
    . setenv.sh
fi

scaffolder

goimports -w src/{{.SourceBase}}

gofmt -w src/{{.SourceBase}}

go install {{.SourceBase}}`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "sh.setenv.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
#!/bin/sh

# This script sets the environment for the {{.Name}} application server.
# It's generated the first time you run the Goblimey scaffolder.  If you
# need to recreate it, run the scaffolder with the -overwrite option.
#
# To run the application server, change directory to the one containing this file 
# and run the commands in it, then run the server.  For example:
#
#    cd $HOME/workspaces/films
#    . setenv.sh
#    {{.Name}}
#
# (Note the "." at the start of the second command.)
#
# The script adds to some environment variables such as the PATH, and it would be
# bad to do this over and over.  The variable {{.NameAllUpper}}_SETUP is used to 
# ensure that those variables are only updated once, so it's safe to run this 
# setup script within other scripts that are run repeatedly. 

if test -z ${{.NameAllUpper}}_SETUP
then
    # ensure that {{.NameAllUpper}}_SETUP exists
    {{.NameAllUpper}}_SETUP="notdone"
	export {{.NameAllUpper}}_SETUP
fi

if test ${{.NameAllUpper}}_SETUP != "done"
then
    {{.NameAllUpper}}_SETUP="done"
	export {{.NameAllUpper}}_SETUP
    # create GOPATH with the project directory or if it already exists, add the 
	# project directory to it
    if test -z $GOPATH
    then
       GOPATH={{.CurrentDir}}
       export GOPATH
    else
        GOPATH=$GOPATH:{{.CurrentDir}}
        export GOPATH
	fi
	# Add the bin directory to the PATH.
	PATH={{.CurrentDir}}/bin:$PATH
    export PATH
fi`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "sh.test.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
#! /bin/sh

# This script creates mock objects and runs the tests.  It's generated the first 
# time you run the Goblimey scaffolder.  If you need to recreate it, run the 
# scaffolder with the -overwrite option.
#
# it lives.  With no argument, run all tests.  With argument "unit" run just the
# unit tests.  With argument "int" run just the integration tests.
#
# The script must be run from the project root, which is where it is stored.

if test -f setenv.sh
then
    . setenv.sh
fi

testcmd='go test -test.v'
if test ! -z $1
then
	case $1 in
	unit )
		testcmd="$testcmd -run='^TestUnit'";;
	int )
		testcmd="$testcmd -run='^TestInt'";;
	* )
		echo "first argument must be unit or int" >&2
		exit -1
		;;
	esac
fi

startDir=%%GRAVE%%pwd%%GRAVE%%

# Build mocks
mkdir -p ${startDir}/src/{{.SourceBase}}/generated/crud/mocks/pegomock
dir='{{.SourceBase}}/generated/crud/mocks/pegomock'
echo ${dir}
cd ${startDir}/src/$dir
pegomock generate --package pegomock --output=mock_template.go {{.SourceBase}}/generated/crud/retrofit/template Template
pegomock generate --package pegomock --output=mock_services.go {{.SourceBase}}/generated/crud/services Services
pegomock generate --package pegomock --output=mock_response_writer.go net/http ResponseWriter
{{range .Resources}}
    mkdir -p {{.NameWithLowerFirst}}
    pegomock generate --package {{.NameWithLowerFirst}} --output={{.NameWithLowerFirst}}/mock_repository.go {{.SourceBase}}/generated/crud/repositories/{{.NameWithLowerFirst}} Repository
{{end}}

# Build

go build github.com/goblimey/animals

# Test

{{range .Resources}}
dir='{{.SourceBase}}/generated/crud/models/{{.NameWithLowerFirst}}'
echo ${dir}
cd ${startDir}/src/$dir
${testcmd}

dir='{{.SourceBase}}/generated/crud/models/{{.NameWithLowerFirst}}/gorp'
echo ${dir}
cd ${startDir}/src/$dir
${testcmd}

dir='{{.SourceBase}}/generated/crud/repositories/{{.NameWithLowerFirst}}/gorpmysql'
echo ${dir}
cd ${startDir}/src/$dir
${testcmd}

dir='{{.SourceBase}}/generated/crud/forms/{{.NameWithLowerFirst}}'
echo ${dir}
cd ${startDir}/src/$dir
${testcmd}

dir='{{.SourceBase}}/generated/crud/forms/{{.NameWithLowerFirst}}'
echo ${dir}
cd ${startDir}/src/$dir
${testcmd}

dir='{{.SourceBase}}/generated/crud/controllers/{{.NameWithLowerFirst}}'
echo ${dir}
cd ${startDir}/src/$dir
${testcmd}

{{end}}`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "sql.create.db.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
-- Command to create the {{.Name}} database.
-- Run these commands as the database admin user.

create database {{.Name}};

grant all on {{.Name}}.* to '{{.DBUser}}' identified by '{{.DBPassword}}';

quit`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "test.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
package {{$resourceNameLower}}

import (
	"testing"
)

// Generated by the goblimey scaffold generator.  You are STRONGLY
// recommended not to alter this file, as it will be overwritten next time the 
// scaffolder is run.

// Unit test for the Concrete{{$resourceNameUpper}} object.

func TestUnitCreateConcrete{{$resourceNameUpper}}Setters(t *testing.T) {
	var expectedID uint64 = 42
	/* 
	    This creates the expected values using the first test value from each field, 
	    something like:
		var expectedForename string = "s1"
		var expectedSurname string = "s3"
	*/
	{{range $index, $element := .Fields}}
	var expected{{.NameWithUpperFirst}} {{.Type}} = "{{index .TestValues 0}}"
	{{end}}
	{{$resourceNameLower}} := MakeInitialised{{$resourceNameUpper}}(expectedID, {{range .Fields}}expected{{.NameWithUpperFirst}}{{if not .LastItem}}, {{end}}{{end}})
	if {{$resourceNameLower}}.ID() != expectedID {
		t.Errorf("expected ID to be %d actually %d", expectedID, {{$resourceNameLower}}.ID())
	}
	{{range .Fields}}
	if {{$resourceNameLower}}.{{.NameWithUpperFirst}}() != expected{{.NameWithUpperFirst}} {
		t.Errorf("expected {{.NameWithLowerFirst}} to be %s actually %s", expected{{.NameWithUpperFirst}}, {{$resourceNameLower}}.{{.NameWithUpperFirst}}())
	}
	{{end}}
}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "utilities.go.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
package utilities

{{.Imports}}

func CreateTemplates() *map[string]map[string]retrofitTemplate.Template {

	templateMap := make(map[string]map[string]retrofitTemplate.Template)
	templateMap["html"] = make(map[string]retrofitTemplate.Template)
	templateMap["html"]["Index"] = template.Must(template.ParseFiles(
		"views/html/index.html",
	))
	templateMap["html"]["Error"] = template.Must(template.ParseFiles(
		"views/html/error.html",
	))
{{range .Resources}}
	templateMap["{{.NameWithLowerFirst}}"] = make(map[string]retrofitTemplate.Template)

	templateMap["{{.NameWithLowerFirst}}"]["Index"] = template.Must(template.ParseFiles(
		"views/generated/crud/templates/_base.ghtml",
		"views/generated/crud/templates/{{.NameWithLowerFirst}}/index.ghtml",
	))

	templateMap["{{.NameWithLowerFirst}}"]["Create"] = template.Must(template.ParseFiles(
		"views/generated/crud/templates/_base.ghtml",
		"views/generated/crud/templates/{{.NameWithLowerFirst}}/create.ghtml",
	))

	templateMap["{{.NameWithLowerFirst}}"]["Edit"] = template.Must(template.ParseFiles(
		"views/generated/crud/templates/_base.ghtml",
		"views/generated/crud/templates/{{.NameWithLowerFirst}}/edit.ghtml",
	))
	
	templateMap["{{.NameWithLowerFirst}}"]["Show"] = template.Must(template.ParseFiles(
		"views/generated/crud/templates/_base.ghtml",
		"views/generated/crud/templates/{{.NameWithLowerFirst}}/show.ghtml",
	))
	
{{end}}
	return &templateMap
}

// BadError handles difficult errors, for example, one that occurs before
// a controller is created.
func BadError(errorMessage string, response *restful.Response) {
	log.SetPrefix("BadError() ")
	log.Println()
	defer noPanic()
	fmt.Sprintf("foo", "1", "2")
	html := fmt.Sprintf("%s%s%s%s%s%s\n",
		"<html><head></head><body>",
		"<p><b><font color=\"red\">",
		errorMessage,
		"</font></b></p>",
		"</body></html>")

	_, err := fmt.Fprintln(response.ResponseWriter, html)
	if err != nil {
		log.Printf("error while attempting to display the error page of last resort - %s", err.Error())
		http.Error(response.ResponseWriter, err.Error(), http.StatusInternalServerError)
	}
	return
}

// Dead displays a hand-crafted error page.  It's the page of last resort.
func Dead(response *restful.Response) {
	log.SetPrefix("Dead() ")
	log.Println()
	defer noPanic()
	fmt.Sprintf("foo", "1", "2")
	html := fmt.Sprintf("%s%s%s%s%s%s\n",
		"<html><head></head><body>",
		"<p><b><font color=\"red\">",
		"This server is experiencing a Total Inability To Service Usual Processing (TITSUP).",
		"</font></b></p>",
		"<p>We will be restoring normality just as soon as we are sure what is normal anyway.</p>",
		"</body></html>")

	_, err := fmt.Fprintln(response.ResponseWriter, html)
	if err != nil {
		log.Printf("error while attempting to display the error page of last resort - %s", err.Error())
		http.Error(response.ResponseWriter, err.Error(), http.StatusInternalServerError)
	}
}

// Recover from any panic and log an error.
func noPanic() {
	if p := recover(); p != nil {
		log.Printf("unrecoverable internal error %v\n", p)
	}
}

// Trim removes leading and trailing white space from a string.
func Trim(str string) string {
	return strings.Trim(str, " \t\n")
}

// Map2String displays the contents of a map of strings with string values as a
// single string.The field named "foo" with value "bar" becomes 'foo="bar",'.
func Map2String(m map[string]string) string {
	// The result array has two entries for each map key plus leading and
	// trailing brackets.
	result := make([]string, 0, 2+len(m)*2)
	result = append(result, "[")
	for key, value := range m {
		result = append(result, key)
		result = append(result, "=\"")
		result = append(result, value)
		result = append(result, "\",")
	}
	result = append(result, "]")

	return strings.Join(result, "")
}`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "view.base.ghtml.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
<!DOCTYPE html>
<html lang="en">
    <head>
        <title>{{"{{"}} template "PageTitle" .{{"}}"}}</title>
        <link href='/stylesheets/scaffold.css' rel='stylesheet'/>
    </head>
    <body>
    	 <h2>{{.Name}}</h2>
    	 <h3>{{"{{"}}template "PageTitle" .}}</h3>
    		<p><font color='red'><b>{{"{{"}}.ErrorMessage{{"}}"}}</b></font></p>
			<p><font color='green'><b>{{"{{"}}.Notice{{"}}"}}</b></font></p>
        <section id="contents">
            {{"{{"}}template "content" .{{"}}"}}
        </section>
    </body>
</html>`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "view.error.html.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
<html>
<head>
	<title>Internal error</title>
</head>
<body>
	<h2>{{.NameWithUpperFirst}}</h2>
    <p>
        <font color='red'><b>Internal Error - please try again later</b></font>
    </p>
</body>
</html>`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "view.index.ghtml.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$projectNameLower := .NameWithLowerFirst}}
{{$projectNameUpper := .NameWithUpperFirst}}
<!DOCTYPE html>
<html lang="en">
    <head>
        <title>{{.NameWithUpperFirst}}</title>
        <link href='/stylesheets/scaffold.css' rel='stylesheet'/>
    </head>
    <body>
    		<h2>{{.NameWithUpperFirst}}</h2>
	{{range .Resources}}
		<p>
			<a id='{{.NameWithLowerFirst}}Link' href='/{{.PluralNameWithLowerFirst}}'>Manage {{.PluralNameWithLowerFirst}}</a>
		</p>
    	{{end}}
    </body>
</html>`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "view.resource.create.ghtml.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNamePluralLower := .PluralNameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
{{"{{"}} define "PageTitle"{{"}}"}}Create a {{.NameWithUpperFirst}} {{"{{end}}"}}
{{"{{"}} define "content" {{"}}"}}
	<p>Items marked "*" are mandatory</p>
    <form action='/{{.PluralNameWithLowerFirst}}' method='post'>
	    	<input id='methodParam' name='_method' value='PUT' type='hidden'/>
	    	<table>
			{{range .Fields}}
			    	<tr>
			    		<td>{{.NameWithUpperFirst}}:</td>
			    		<td>
					{{if eq .Type "bool"}}
						<input id='{{.NameWithLowerFirst}}' type="checkbox" name='{{.NameWithLowerFirst}}' value='true' {{"{{if "}}.{{$resourceNameUpper}}.{{.NameWithUpperFirst}}{{"}}checked{{end}}"}} /> 
					{{else}}
						<input id='{{.NameWithLowerFirst}}' type='text' name='{{.NameWithLowerFirst}}' value='{{"{{"}}.{{$resourceNameUpper}}.{{.NameWithUpperFirst}}{{"}}"}}'/>
					{{end}}
					</td>
					<td>{{if .Mandatory}}<td><font color='red'><b>*</font></td>{{end}}</td>
			    		</td>
					{{"{{if .FieldErrors."}}{{.NameWithUpperFirst}}{{"}}"}}
			    			<td><span id='{{.NameWithUpperFirst}}Error'><font color='red'>{{"{{.FieldErrors."}}{{.NameWithUpperFirst}}{{"}}"}}</font></span></td>
			    		{{"{{end}}"}}
					</td>
			    	</tr>
	    		{{end}}
	    </table>
	    <input id='CreateButton' type='submit' value='Create'/>
	</form>
	<p>
		<a id='homeLink' href='/'>Home</a>
		<a id='viewLink' href='/{{.PluralNameWithLowerFirst}}'>View All {{.PluralNameWithUpperFirst}}</a>
	</p>
{{"{{end}}"}}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "view.resource.edit.ghtml.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNamePluralLower := .PluralNameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
{{"{{"}} define "PageTitle"{{"}}"}}Edit {{.NameWithUpperFirst}} {{"{{."}}{{.NameWithUpperFirst}}.DisplayName{{"}}"}}{{"{{end}}"}}
{{"{{"}} define "content" {{"}}"}}
    <form id='updateForm' action='/{{.PluralNameWithLowerFirst}}/{{"{{"}}.{{.NameWithUpperFirst}}.ID{{"}}"}}' method='post'>
    	<input name='_method' value='PUT' type='hidden'/>
    	<table>
		{{range .Fields}}
		    	<tr>
		    		<td id='{{.NameWithUpperFirst}}Label'>{{.NameWithUpperFirst}}:</td>
		    		<td>
				{{if eq .Type "bool"}}
					<input id='{{.NameWithLowerFirst}}' type="checkbox" name='{{.NameWithLowerFirst}}' value='true' {{"{{if "}}.{{$resourceNameUpper}}.{{.NameWithUpperFirst}}{{"}}checked{{end}}"}} /> 
				{{else}}
					<input id='{{.NameWithUpperFirst}}Value' type="text" name='{{.NameWithLowerFirst}}' value='{{"{{"}}.{{$resourceNameUpper}}.{{.NameWithUpperFirst}}{{"}}"}}'/>
		    		{{end}}
				</td>
				<td>{{if .Mandatory}}<td><font color='red'><b>*</font></td>{{end}}</td>
					{{"{{"}}if .ErrorForField "{{.NameWithUpperFirst}}" {{"}}"}}
		    			<td><span id='{{.NameWithUpperFirst}}Error'><font color='red'>{{"{{"}}.ErrorForField "{{.NameWithUpperFirst}}"{{"}}"}}</font></span></td>
		    		{{"{{else}}"}}
		    			<td>&nbsp;</td>
		    		{{"{{end}}"}}
		    	</tr>
	    	{{end}}
	    </table>
	    <input id='UpdateButton' type='submit' value='Update'/>
	</form>
	<p>
		<form id='deleteForm' action='/{{.PluralNameWithLowerFirst}}/{{"{{"}}.{{.NameWithUpperFirst}}.ID{{"}}"}}/delete' method='post'>
			<input id='MethodParam' name='_method' value='DELETE' type='hidden'/>
			<input id='deleteButton' type='submit' value='Delete'/>
		</form>
    </p>
	<p>
		<a id='homeLink' href='/'>Home</a>
		<a id='ShowLink' href='/{{.PluralNameWithLowerFirst}}/{{"{{"}}.{{.NameWithUpperFirst}}.ID{{"}}"}}'>Show</a>
		<a id='ViewLink' href='/{{.PluralNameWithLowerFirst}}'>View All {{.PluralNameWithUpperFirst}}</a>
		<a id='CreateLink' href='/{{.PluralNameWithLowerFirst}}/create'>Create {{.NameWithUpperFirst}}</a>
	</p>
{{"{{end}}"}}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "view.resource.index.ghtml.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNamePluralLower := .PluralNameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
{{"{{"}} define "PageTitle"{{"}}"}}{{.PluralNameWithUpperFirst}}{{"{{end}}"}}
{{"{{"}} define "content" {{"}}"}}
    <table>
    {{"{{range ."}}{{.PluralNameWithUpperFirst}} {{"}}"}}
        <tr>		
        		<td>
	            <a id='LinkToShow {{"{{."}}DisplayName{{"}}"}}'  href='/{{$resourceNamePluralLower}}/{{"{{.ID}}"}}'>{{"{{."}}DisplayName{{"}}"}}</a>
            </td>
			<td>
	            <a id='LinkToEdit {{"{{."}}DisplayName{{"}}"}}' href='/{{$resourceNamePluralLower}}/{{"{{.ID}}"}}/edit'>Edit </a>
            </td>
            <td>
		        <form action='/{{.PluralNameWithLowerFirst}}/{{"{{.ID}}"}}/delete' method='post'>
			        <input name='_method' value='DELETE' type='hidden'/>
			        <input id='DeleteButton_{{"{{.ID}}"}}' type='submit' value='Delete'/>
		        </form>
            </td>  
        </tr>	
    {{"{{end}}"}}
    </table>
    <p>
		<a id='homeLink' href='/'>Home</a> 
		<a id='CreateLink' href='/{{.PluralNameWithLowerFirst}}/create'>Create {{.NameWithUpperFirst}}</a>
	</p>
{{"{{end}}"}}`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "view.resource.show.ghtml.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
{{$resourceNameLower := .NameWithLowerFirst}}
{{$resourceNameUpper := .NameWithUpperFirst}}
{{"{{define"}} "PageTitle"{{"}}"}}{{.NameWithUpperFirst}} {{"{{."}}{{.NameWithUpperFirst}}.DisplayName{{"}}"}}{{"{{end}}"}}
{{"{{define"}} "content"{{"}}"}}
    <p>
    	<b>id:</b> <span id='id'>{{"{{"}}.{{.NameWithUpperFirst}}.ID{{"}}"}}</span>
	</p>
	{{range .Fields}}
	    <p>
	    	<b>{{.NameWithLowerFirst}}:</b> <span id='{{.NameWithLowerFirst}}'>{{"{{"}}.{{$resourceNameUpper}}.{{.NameWithUpperFirst}}{{"}}"}}</span>
		</p>
	{{end}}
	<div id='DeleteButton' style='display: inline;'>
		<form id='DeleteForm' action='/{{.PluralNameWithLowerFirst}}/{{"{{"}}.{{.NameWithUpperFirst}}.ID{{"}}"}}/delete' method='post' style='display: inline;'>
			<input id='MethodParam' name='_method' value='DELETE' type='hidden'/>
			<input id='DeleteButton' type='submit' value='Delete'/>
		</form>
	</div>	
	<p>
		<a id='homeLink' href='/'>Home</a>
		<a id='EditLink' href='/{{.PluralNameWithLowerFirst}}/{{"{{"}}.{{.NameWithUpperFirst}}.ID{{"}}"}}/edit'>Edit</a>
		<a id='ViewLink' href='/{{.PluralNameWithLowerFirst}}'>View All {{.PluralNameWithUpperFirst}}</a>
	</p>
{{"{{end}}"}}
`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}

templateName = "view.stylesheets.scaffold.css.template"
	if useBuiltIn {
		if verbose {
			log.Printf("creating template %s from builtin template", templateName)
		}
		templateText := `
div.notice {
  color: green;
  font-weight: bold;
}

div.ErrorMessage {
  color: red;
  font-weight: bold;
}`
		templateText = substituteGraves(templateText)
		templateMap[templateName] =
			template.Must(template.New(templateName).Parse(templateText))
	} else {
		if verbose {
			log.Printf("creating template %s from file %s", templateName, templateRoot+templateName)
		}
		templateMap[templateName] = createTemplateFromFile(templateName)
	}
}
