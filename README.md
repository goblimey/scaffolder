# Scaffolder
Given a description of a database and its tables, the Goblimey Scaffolder creates that database
and generates a website which provides the Create, Read, Update and Delete (CRUD) operations
on it.
The scafolder is written in the Go programming language, as is the web server that it produces.
The server is designed according to the Model, View, Controller (MVC) architecture and is
implemented using RESTful requests.  (Those buzzwords are explained below.)
The result is presented as a complete prototype Go project with all source code included
along with unit and integration tests to check that the whole thing hangs together.

The idea of the scaffolder is taken from the Ruby-on-Rails scaffold generator.

Producing the right design for the database that sits behind a web site usually takes several attempts.
The scaffolder gives you a quick and easy way to create databases, put data into them and experiment with different versions.
Once you have a working version, you can extend the source code and produce your own production web server.
That's much easier than producing one from scratch.

Producing a complete piece of working source code also makes the scaffolder
a very useful aid to learning Go.
That means that it may be used by people who are new to Go and
possibly new to programming,
so this document doesn't assume extensive knowledge of Go or the Go tools.
That makes it a bit longer and more complicated than it would otherwise need to be,
but readers can just skip the parts that they understand.

I also assume that you might be using a computer which is running Microsoft Windows.
Many people (including me) think that Windows is not the best operating system
to use for software development, and most Go developers use a Linux machine or an Apple Mac.
However, if Windows is the only system you know, it's perfectly adequate.


For the Impatient
============
Get the dependencies and install the scaffolder

    go get github.com/go-sql-driver/mysql
    go get gopkg.in/gorp.v1
    go get github.com/emicklei/go-restful
    go get github.com/onsi/gomega
    go get golang.org/x/tools/cmd/goimports
    go get github.com/petergtz/pegomock/pegomock
    go get github.com/goblimey/scaffolder

Create a web server

* create an empty database
* create a Go project directory and cd to it
* create a specification of the tables and fields
* $ scaffolder
* $ ./install.sh
* start the web server
* in your web browser, navigate to <http://localhost:4000>
* create some data

Once you have downloaded the scaffolder, you can find an example table specification in the examples directory. 

By default, "go get" doesn't update any projects that you have already downloaded.
If you downloaded any of those projects a long time ago, 
you may wish to update it to the latest version using the -u flag, for example:

    go get -u github.com/petergtz/pegomock/pegomock


Creating Your Project
================

Go programmers can skip this section.

You need to have installed MySQL and the Go tools.  instructions are at the end of this document if you need them.

A Go project is really just a directory containing some Go source code.
However, as the How to Write Go Code document explains,
you should structure your project as if you are going to store it in a repository.
The Github is the most popular repository so I'm going to assume that you will store it there.

You cannot easily create a project on your computer and store it on the github later.
It's much easier to create a github project first.

If you don't have a Github account, create one at [github.com](https://github.com).
It's free.  For example, if your name is Alun Smithie, you could create an account called alunsmithie.
Your home page on the github would then be 
https://github.com/alunsmithie.
I'm going to assume that, so in the following examples, replace "alunsmithie" with your github account name.

On your github home page, use the "+" button at the top to create a project.
If you call your projects "animals", it will have a URL something like https://github.com/alunsmithie/animals.

When you create your project, the Github will encourage you to
specify what sort of licence you will issue when you publish the project,
and to set up a README.md file describing what the project is about.
(You are currently reading the README.md for the scaffolder project.)

Now create a clone of this project on your computer.  You can do all this in a command window.
(On Windows 7 use the Command Prompt option in the Start menu)

On Linux:

    $ mkdir -p $GOPATH/src/github.com/alunsmithie
    $ cd $GOPATH/src/github.com/alunsmithie
    $ git clone https://github.com/alunsmithie/animals
   
On Windows:

     mkdir %GOPATH%\src
     mkdir %GOPATH%\src\github.com
     mkdir %GOPATH%\src\github.com\alunsmithie
    cd %GOPATH%\src\github.com\alunsmithie
    git clone https://github.com/alunsmithie/animals

That creates a Go project directory "animals" containing a copy of any files you created on the github site and also some magic hidden files that make it a local git repository.


The JSON Specification
======================

The scaffolder is driven by a text file in JavaScript Object Notation (JSON) format that specifies a database and a set of tables.  JSON is described [here](http://www.json.org).

When you are writing JSON, it's very easy to make a simple mistake such as missing out a comma.
The scaffolder uses an off-the-shelf JSON processor and the error messages it produces are not very helpful.
You will save yourself a lot of pain if you prepare the file 
using an editor that understands JSON and warns you about obvious errors.
Most Integrated development Environments (liteIDE, Eclipse, IntelliJ, VSCode etc) have editors that will do this.  Text editors such as Windows Notepad++ will do the same.

The scaffolder includes an example specification file so you can use that for a quick experiment.
Copy goprojects/scaffolder/examples/animals.scaffold.json into your project directory and rename it scaffold.json.

That specification defines a MySQL database called "animals" containing tables "cats" and "mice":

    {
        "name": "animals",
        "db": "mysql",
        "dbuser": "webuser",
        "dbpassword": "secret",
        "dbserver": "localhost",
        "orm": "gorp",
        "sourcebase": "github.com/alunsmithie/animals",
        "Resources": [
            {
                "name": "cat",
                "fields": [
                    {
                        "name": "name", "type": "string", "mandatory": true,
                        "testValues": ["a","b"]
                    },
                    {
                        "name": "breed", "type": "string", "mandatory": true
                    },
                    {
                        "name": "age", "type": "int", "mandatory": true,
                        "excludeFromDisplay": true
                    },
                   {
                        "name": "weight", "type": "float", "mandatory": true,
                        "excludeFromDisplay": true
                    },
		    {
                        "name": "chipped", "type": "bool",
                        "excludeFromDisplay": true
                    }
                ]
            },
            {
                "name": "mouse", "plural": "mice",
                "fields": [
                    { "name": "name", "type": "string", "mandatory": true },
                    { "name": "breed", "type": "string", "excludeFromDisplay": true }
                ]
            }
        ]
    }

As the JSON website explains, JSON is built on two structures:

* A collection of name/value pairs.
* An ordered list of values (AKA an array).

For example in the first line above "name": "animals" is a pair with name "name" and value "animals".  It defines the name of the resulting go project and the database table that it controls.
The value can be a strings of text, a number, a boolean value (true or false) and so on.

A JSON list is enclosed in brackets, for example 

    "testValues": ["a","b"]

That defines a list called "testValues" which contains strings of text.

The Resources section is a more complicated list, containing a sub-hierarchy of pairs.

As far as the scaffolder is concerned, each resource defines a database table, a model with an associated repository, a controller and a set of views.

In the example, the first few lines of the JSON define the database.  It's the one we created earlier, a MySQL database called "animals" accessed using the user name "webuser" and the password "secret".
"localhost" means that the MySQL server is running on this computer and listening on the default port. (You can specify the port  like so: "dbport": "1234".)

The ORM pair says which ORM to use.  At present
the only one supported is [GORP](https://github.com/coopernurse/gorp) version 1.
I plan to add support for other ORMs in the future.

The sourcebase says where the generated source code should be stored within the project. This should follow the Go package layout conventions. 
In this example the project will stored in the github repository http://github.com/alunsmithie/animals, so the sourcebase value is "github.com/alunsmithie/animals".  Given that, the scaffolder creates material in src/github.com/alunsmithie/animals within the go workspace directory.

Next comes the Resources section, which is a JSON list.
Each entry describes a resource with an associated database table,
This example describes the "cat" resource handling the "cats" table and the "mouse" resource handling the "mice" table.

Traditionally, database tables are named using the plural of the data that they contain.
If that's just the singular with an "s" added, you don't need to specify it.
The plural of "mouse" is "mice" so you the spec has to define that.

    "name": "mouse", "plural": "mice",

Each resource section defines a list of fields.  The cat resource has fields "name" and "breed" which contain strings of text,
"age" containing an integer (a whole number)
"weight" containing a floating point number (one with a fractional part)
and "chipped" containing a boolean value (true or false)
recording whether or not the cat has been microchipped.
All fields but the last are mandatory.
The mouse resource has just two fields.

The scaffolder generates a set of unit and integration test programs to check that the generated source code works properly.
A unit test takes a module of the source code and runs it in isolation, supplying it with test values and checking that the module produces the expected result.  An integration tests is similar, but tests that a set of modules work together properly.
Each field in the JSON can have a list of testValues to be used by the tests.
This is optional.  If you don't specify test values, they are generated automatically.
If you don't specify enough values, the rest are generated automatically.
Currently none of the the generated tests use more than two values,
so a list of two values is sufficient. 

The optional excludeFromDisplay value in the JSON 
controls a display label which is used to identify each database record on the generated web pages.
For example in the cats resource the fields "age", "weight" and "chipped" are excluded
from the display label, leaving the ID, name and breed fields.
So if Tommy the Siamese cat is described by record 42 in the cats table the display label for that record will be:

    42 Tommy Siamese   

The display used for various purposes.
For example, the index page for a resource shows a list of display labels, one for each record.

By default the display label contains the value of every field in the record.
That can be a lot more information than you want and often just a few fields are enough.

The display label is also used to create the title of the page that displays a single record
(the Show page) and the one that allows you to change the data in the record (the Edit page).
Within the HTML the label also used to create the IDs of the document fields to aid automated testing tools such as Selenium.

The scaffolder doesn't handle relations between tables, which is a serious omission.  I intend to add one to many and many to many relations in the next version.


Creating a Database
==================

The JSON in the previous section expects a database called "animals" that can be accessed by the MySQL user "webuser" using the password "secret".  You need to create that.

Run the MySQL client in a command window:

    mysql -u root -p
    {type the root password that you set when you installed mysql}
    
    mysql> create database animals;
    mysql> grant all on animals.* to 'webuser' identified by 'secret';
    mysql> quit

The user has all access rights, so it can create tables.
The web server generated by the scaffolder will connect to this database
and create the tables.

The tables will be created the first time you run the web server. 
Each table will have the fields specified in the JSON, plus an auto-incremented unique numeric ID.
The cats table will look like this:

    mysql> describe cats;
    +---------+---------------------+------+-----+---------+----------------+
    | Field   | Type                | Null | Key | Default | Extra          |
    +---------+---------------------+------+-----+---------+----------------+
    | id      | bigint(20) unsigned | NO   | PRI | NULL    | auto_increment |
    | name    | varchar(255)        | YES  |     | NULL    |                |
    | breed   | varchar(255)        | YES  |     | NULL    |                |
    | age     | bigint(20)          | YES  |     | NULL    |                |
    | weight  | double              | YES  |     | NULL    |                |
    | chipped | tinyint(1)          | YES  |     | NULL    |                |
    +---------+---------------------+------+-----+---------+----------------+

The id of each record will be set automatically to a unique value.

Building the Server
======================

When you run the scaffolder, by default it looks for a specification file "scaffold.json" in the current directory - something like the example above.  You can specify a different file if you want to.

By default the scaffolder generates the server in the current directory, which should be your github project directory (in the example, goprojects/src/github.com/alunsmithie/animals).
Alternatively you can run it from another directory and tell it where to find the project directory.

To get started, there is a version of the example JSON in the scaffolder source code.
In your Go workspace, look in scaffolder/examples.  
Copy the file animals.scaffold.json to your project directory and rename it scaffold.json.  Edit it to set your sourcebase ("github.com/alunsmithie/animals" or whatever).

In your command window, change directory to your project and run the scaffolder

    $ scaffolder

That creates the web server source code and some scripts.  The script install.sh  builds and installs the server on Linux:

    $ ./install.sh

install.bat does the same on Windows:

    install

There is also test.sh and test.bat.  These run the tests to ensure that all the bits work:

    $ ./test.sh

If all the tests pass, you can start the web server.  If goprojects/bin is in your path, you can run it like so:

     $ animals

or you can run it in verbose mode and see tracing messages in your command window:

     $ animals -v

The first time you run the server it will create the database tables.

The server runs on port 4000.  In your web browser, navigate to <http://localhost:4000>

That display the home page.  It has two links "Manage cats" and "Manage mice".
The first takes you to the index page for the cat resource.
The cats table is currently empty.  Use the Create button to create some.

Once you've done that, the index page lists the cats using the display label (the id, name and breed).
There are links and buttons to edit and delete the records, and a link back to the home page.

To add some mice, use the link to the home page and then the "Manage mice" link.

To stop the server, type ctrl/c in the command window.  (Hold down the ctrl key and type a single "c", you don't need to press the enter key.)

The scaffolder creates these files

* install.sh - a shell script to build the animals server
* install.bat batch script to do the same on Windows
* test.sh - a shell script to run the test suite
* test.bat same for Windows
* animals.go - the source code of the main module
* generated - the source code of the models, views, controllers, repositories and support software
* views - the templates used to create the html views.

You can edit the JSON and add some fields.  For example, you could add a field "favouritefood" to the cats table.  Run the scaffolder again and it will produce a new version of the server.  Run the install script to build and install it.

It's assumed that you may want to tweak things like the scripts, the main program, the home page  and so on.  If you run the scaffolder over this project again, by default only the stuff in the "generated" directories is overwritten.  

If you run the scaffolder with the overwrite option, it replaces everything.

The new server won't add the new field to the database tables.  If you don't mind losing the data you've already created, use the mysql client to drop the tables before you start the server.  It will then create new ones.  Alternatively, you can add the extra field to the table using the MySQL client.

If you change the JSON it's a good idea to run the tests again to make sure that nothing has been broken.  However, some of the integration tests write to the database and they will also trash any existing data if you run them.  If you want to avoid that, once you start changing the JSON and rebuilding the server, you can run just the unit tests:

    $ ./test.sh unit

To avoid all these problems, you can just create a new project that uses a different database.

To specify the JSON file:

    $ scaffolder ../specs/animals.json

To specify the workspace directory as well:

    $ scaffolder workspace=/home/simon/goprojects ../specs/animals.json

Run the scaffolder program like so to see all of the options:

    $ scaffolder -h
    Usage of scaffolder:
      -overwrite
          overwrite all files, not just the generated directory
     -projectdir string
          the project directory (default ".")
     -templatedir string
          the directory containing the scaffold templates (normally this is not specified and built in templates are used)
      -v enable verbose logging (shorthand)
      -verbose
          enable verbose logging



Keeping Track of Your Changes
=======

Your project directory is also a local git repository, so it can keep track of any changes you make to the files in it.  Git uses a two-stage commit process.
Once you have created some files, you first commit them to the local repository.
As you change the files, you can commit the new versions.
The local repository keeps track of the changes and you can roll a file back to an earlier version
if you make some unfortunate mistake. 
When you are happy with your project, you can "push" the current state of the repository to the github.com server, 
at which point your files become public and other people can get at them.

Once you've created your project, take a snapshot of progress by committing the files to your local repository:

    git add scaffold.json
    git add install.sh install.bat test.sh test.bat animals.go
    git add views/html/index.html views/html/error.html

and so on, followed by

   git commit -m "initial version"

Actually, since all of the other files are created from the scaffold.json, that may be the only one you need to commit.  

If create a project and then change any of the source files, you should add and commit them to the repository as well.  Don't change or commit anything in a directory called "generated", because it will be overwritten every time you rerun the scaffolder.

If one of the files that you have committed to the repository gets wrecked, you can restore it to the last commit:

    git checkout views/html/index.html

So just committing changes your local repository is useful even if you don't plan to publish the project on Github. 

Unfortunately, that requires you to open a Github account and create an empty project, just so you can make a local clone of it.  If that irks you, you could install your own git server and use that to keep track of your files instead.

If you do want to make your project public, once you have a working version, push the current state of the repository to the Github:

    git push

If you haven't configured git with your Github user name and password, it will ask you to supply them every time you do this.


Design Principles
=====================================

Given a description of a database, the scaffolder writes a Go program that creates the database
and provides a web server that allows you to create, read, update and delete records.  The web server builds HTML pages to order, depending on the data in the database.  Such a server is sometimes called a web application server, to distinguish it from a web server that simply feeds out static pages.

The web application server provides controlled access to data in the database. Each response is manufactured to order, based on the data. In a production system, it's usually impossible for a user to access the database directly.  They can only do it via the web server and they can only do what the web server allows.

The web server is designed using Model, View, Controller architecture, usually just called MVC. MVC benefits include:

* Isolation of business logic from the user interface
* Ease of keeping code DRY
* Clarifying where different types of code belong for easier maintenance

(DRY means Don't Repeat Yourself, in other words, avoid writing the same source code more than once.) 

A model represents the information (data) of the application and the rules to manipulate that data. 
In this case, models are used to manage the interaction with a corresponding database table. 
Each table in your database will correspond to one model in your application. 
There is also a corresponding Repository (AKA a Data Access Object (DAO)),
which is software used to fetch and store data.
(Not to be confused with the GIT repository.)

The Views represent the user interface of your application. 
They handle the presentation of the data. Views provide data to the web browser or other tool that is used to make requests from your application.
This web server generates an HTML response page for each request on the fly from a template, so its views are Go HTML templates.

The controller provides the “glue” between models and views.  The controller is responsible for processing the incoming requests from the web browser, interrogating the models for data, and passing that data on to the views for presentation.  There is one controller for each table.

Part of the job of designing a website is to define the HTTP requests that it accepts. 
A web browser and web server communicate with each other using HTTP requests and responses.  This happens behind the scenes, so it's not obvious.

For example, the page you are reading contains a link to http://localhost:4000. That's called a Uniform Resource Location (URL).  If you click on that link in your browser, the browser issue an HTTP request to a web server running on this computer and listening on port 4000.  The request looks like this:

    GET /

meaning get the server's home page.

The web server sends back an HTTP response.   It could contain an Excel spreadsheet, an Acrobat document, a sound file, a video or pretty much anything that can be stored on a computer.
In this case the response contains an HTML page, which the browser renders and displays. 

An HTTP request has a structure.  For example:

    GET /cats?operation=delete&id=42

The first part is the method - GET, POST etc.

The next part is the Universal Resource Identifier (URI). It's the part up to the parameters, in this case "/cats".  The whole URL identifies which server and port to send the request to, and the URI part says which resource to access on that server.  What it means to access a resource is up to the server.

The ? marks the start of the parameters, in this case we have a parameter called "operation" with value "delete" and another "id" with value 42.

When you design a web application server, you need to decide what requests it will accept and what response each request will provoke.

The REpresentational State Transfer (REST) model provides a coherent framework for requests and responses.  REST is described [here](https://en.wikipedia.org/wiki/Representational_state_transfer).

The REST model stresses the use of resource identifiers such as URIs to represent resources.
In the web server created by the scaffolder, each resource has an associated database table.
If we have a database table called "cats" then

    GET /cats
    GET /cats/

are both RESTful HTTP request to display a list of all cats in the table
(the "index" page for the cat resource).

    GET /cats/42

is a RESTful HTTP request to display the data of the cat with ID 42.

    DELETE /cats/42

is a RESTful HTTP request to delete that record.

REST requires that all GET requests are idempotent ("having equal effect").
This simply means that if browsers issue the same GET request many times and
the data in the database is not changed by some other agent, each request will produce the same response.
The upshot of that is that we use GET requests to read data from the database and other requests (PUT, DELETE etc) to change the data.

One advantage of the REST approach is that search engines respect this rule.  If your web site is public it will be crawled repeatedly by lots of search engines.  When a search engine crawls a site and attempts to find all the pages.  It does this by reading the home page, looking through it for HTTP requests to other pages, reading those pages and so on.  It will attempt to issue all GET requests but it will avoid issuing any other requests that it finds.  The crawler assumes that a GET requests won't change the data but any other request might.

REST requires that requests containing parameters are only used to submit form data. For example if this request deletes the record with ID 42:

    GET /cats?operation=delete&id=42

it doesn't follow the REST rules, firstly because it's using a GET request to change the data and secondly because it uses parameters but it's not carrying form data.

There's no REST policeman that will stop your web server using non-Restful requests.
REST is just a set of rules that make for a clean design if you follow them.

While you are using your browser to look at a web site, you can see the HTML requests that it's issuing.  For example, in Firefox, press the button at the top right-hand corner of the browser marked with three vertical lines.  In the resulting menu, choose  Developer and then Web Console.  The web console appears at the bottom of your browser.  If you hit a link on the page or press a button, that issues an HTTP request.  The console tab in the web console shows you the last request.  If you expand it, it shows you the contents.  Other browsers have a similar mechanism.


How the Server Handles a Request
==========================
When the web server starts, the main method in animals.go registers each request that it will handle and arranges that when any of them arrive, the marshal function is called to handle it, Given an HTTP request the marshal function figures out which controller to use and which of its method to call.

The main function starts the server, which runs until it's forcibly stopped.  It waits for HTTP requests and processes each one, producing an HTTP response.  The response is always an HTML web page.

When a request arrives, it's examined behind the scenes to check that it's one of the registered ones.  If not, it's rejected out of hand.  If it's registered, it's passed to the marshal function, which  examines it and decides which controller to use and which controller method to call.  

For example, the user wants to change the data for a cat in the database.  She views the cat index page, chooses one of the entries and clicks on its Edit link.  The browser sends this request to the server:

   GET /cats/42/edit

The marshal function is called to handle the request.

The marshal function creates a new instance of the cat controller and a cat repository.  (The repository is a bit of software that handles access to the database table.)  Creating a repository also allocates a database connection. There's a limit to how many connections you can have open at any one time, so at the end of this process, it's closed and recycled.  The marshal method also creates a service object which the controller will use to get at objects such as the repository.  It passes the service object to the controller in the constructor when it creates it.

To digress briefly, the idea of the service object is that the controllers don't create objects themselves, they get the service object to do it for them. It creates objects for them and conceals precise information about what those objects are.   For example, the cat controller calls a method in the service object to get a repository, and another to create a cat form.  These objects are defined using interfaces, so the controller only knows that it is being given an object that satisfies that interface - it doesn't know the precise type of the object. This allows us to write tests that run the controller in isolation, passing in carefully prepared dummy objects that supply just the right data for the tests.

The marshal method extracts the id from the request (in this case the id is 42) and creates a cat form object.  This contains a cat model object, which itself contains the id.  The marshal function puts the id into the model object and calls the controller method, passing the request, the response and the form as parameters.

The controller gets the ID from the form and looks up the record in the database.  It fills in the rest of the model object with those values - for a cat, the name, the breed, the age and so on. It gets the HTML template for the cat edit page from the service object and executes it, passing the form object to it.  That causes the server to create a response containing an HTML page and send it to the browser.  The page contains the values supplied in the form object.

The edit page contains an HTML form that allows the user to submit data.  The fields of the HTML form initially contain the data in the cat form object, so the user sees a form with boxes containing the current version of the cat's name, breed, age and so on.

The user changes the values in the form and presses the submit button.  That issues a new request:

    PUT /cats/42

The body of the request contains the form contents as parameters.  (Try updating a cat record and use the web console to see this request being issued. Open the response to see the data being sent.

That request is sent to to the marshal, which goes through a similar set of moves as before, but this time it gets the values coming in from the HTML form (name, breed age, etc) and uses them to fill in the model in the cat form object.  Then it calls the cat controller's Update method, passing the request, response and form.

The Update method validates the form.  For example, the name field is mandatory and must be filled in.  If any fields are invalid, Update adds an error message for that field to the form and invokes the edit page again.  The user sees an HTML form containing the values she filled in last time, with error messages against the invalid fields.

The user corrects the data and submits the HTML form again.  This time she gets everything right.  The Update method is called, validates the data and sees that it's correct.   It updates the record in the database.  Then it uses the repository to get a list of cats, creates another form containing a notification that it was successful and responds with the cat index page, showing the new list of cats plus the notification in green at the top.

The form object which is passed to a template to create a HTML page contains all the data that the page needs to display.  For the cat index page, that's a notification or error message to be displayed at the top and a list of cats.  For the cat edit page, it's a notification or error message to be shown at the top, the data for a single cat and a list of error messages, one for each field.  All the other cat pages can be satisfied by that form too, so we have just two forms, one containing a single cat object and another for the index page containing a list of cats.  We have two similar forms for the mice.

I stole the idea of using a common form object to carry data between the HTML pages and the controller from the Java Struts framework. 


Enhancing Your Server
================

The scaffolder generates a web server including complete source code, so you can see how it works.  You can use the result as a starting point to create your own server.

The easiest change is to alter the views.  For example, the HTML that appears at the beginning and the end of almost all of the pages is controlled by the base template views/_base.ghtml.  To add a standard banner to the top of those pages, edit that file.  The only other pages are the home page and the static error page in views/html.  You will need to edit those separately.  There also a CSS stylesheet in views/stylesheets. You can add your own styling controls there.

The templates for the individual views are in views/generated.  You don't want to change any of those files, because they will be trashed if you run the scaffolder again.  Instead, create a new directory in views.  Copy the source code of a generated view into there and tweak the copy.  For example, to change the cat edit view, create a directory views/cats and copy views/generated/crud/templates/cat/edit.ghtml to it. Make your changes to that file.

When a controller needs a view it gets it from a table.  The table is set up in utilities.go.  For example:
 this section sets up the cat edit view:

    templateMap["cat"]["Edit"] = template.Must(template.ParseFiles(
		"views/_base.ghtml",
		"views/generated/crud/templates/cat/edit.ghtml",
	))

Change that to:

    templateMap["cat"]["Edit"] = template.Must(template.ParseFiles(
		"views/_base.ghtml",
		"views/cat/edit.ghtml",
	))

Run the install script to rebuild the server and now it
will use that version of the cat edit template instead of the generated one.

More complex enhancements will require changes to the generated Go source code.  Before you can do that, you need to understand the existing design.  Read the previous section and work through the source code that handles a request, starting with the main program animals.go.

Once you have understood all that, you can add your own functionality to the web server.  For example, this is how to create an extra cat request:

You don't want to tweak the generated controller, because it will be trashed if you run the scaffolder again.  Instead, create a directory "controllers" at the top level (in the "animals" directory).  Copy the source code of the existing cat controller into there and tweak the copy.

Now you need to get your head around the Go package structure and the import mechanism.  At the top of the main module in animals.go, there is an import directive:

    catController "github.com/goblimey/animals/generated/crud/controllers/cat"

Change that to

    catController "github.com/goblimey/animals/controllers/cat"

So now the server will use your cat controller instead of the generated one.

Edit the main function in animals.go to register your new request and send it to the marshal method.

Edit the marshal method to intercept the new request and create and call your new controller method.

Run the install script to rebuild the server. 

If you want to write your own database repository, follow a similar line of attack.  The repositories are supplied to the controllers using the service object, so you need to tweak that to supply your version.  The repositories are added to the service object at the start of the main module's marshal method, for example this adds the cat repository:

    services.SetCatRepository(CatRepository)

Support for Testing
==================

The server source code is set up the way it is, with objects defined as interfaces, to support thorough testing.

Each module of source code in the generated project has an associated test program, stored in the same directory.  The test script runs through those directories and runs all the tests in them. 
If you add your own source code, write more tests.  Keep the test script up to date as you add directories to the source tree.

You can, of course, test everything by hand.  The problem comes when you change anything.  You need to run all the tests again to make sure that everything still works.  It's very easy to fix one thing and break another.  Recording a set of automated tests that run quickly and do a thorough test of the basic logic ensures a more consistent result.

The scaffolder generates unit tests and integration tests.  A unit test test a module of source code in isolation, without involving any of the other modules.  When the module under test depends on another module, the test supplies it with a fake version that satisfies the same interface.

An integration test checks that several modules work together.  For example, some of the integration tests check that a repository can manipulate data in the database, and they use the real database to do that.

Unit tests are very useful for checking the fine logic of a method.  For example, you can use a unit test to drive the method through a path of logic that it wouldn't normally follow, and make sure it works properly if that condition ever happens. Unit tests tend to run faster than integration tests, because they involve less baggage, but they are limited as to what they can check. You need both types of test (and more).

The Go tests framework allows you to distinguish between different types of test.  The names of all the generated unit tests start with TestUnit, and those of the integration tests all start with TestInt.  The tester can run them according to that pattern.  

This runs all the tests:

   $ ./test.sh

This runs just the unit tests:

    $ ./test.sh unit

and this runs just the integration tests:

  $ ./test int

On Windows that's:

    test int

NOTE that some of the integration tests use the real database.  It's safe to run them when you have just created a new server, but if you run them after you've played with it and created some data, that data will be trashed.

    $ ./test.sh unit

runs just the unit tests, and it's always safe to run them.

The generated unit tests make extensive use of a technique called mocking.  Mocking uses clever objects (mocks) that satisfy an interface and can be made to perform to order during a test.  The tests script uses [pegomock](https://github.com/petergtz/pegomock) to produce mocks.

The test program tells the mock what to do.  For example controller_test.go in the cat controller module tests the cat controller. 
The first test checks that the Index method produces the right response when everything works OK.

The Index method gets a repository from the service object and runs its FindAll method.  In the real world this gets a list of cats from the database and the controller creates an HTML page containing the list and sends it as the response.

To check that the controller does what's expected, the test program creates a mock repository and a list of cats called expectedCatList.  Then it does this:

    pegomock.When(mockRepository.FindAll()).ThenReturn(expectedCatList, nil)

which says to the mock repository "when your FindAll method is called, it should return the expectedCatList and nil (no error).

The test then creates a cat controller and calls its Index method:

    var controller Controller
    controller.SetServices(&services)
    controller.Index(&request, &response, form)

The services object, the repository, the templates, the request, the response and the form are all fakes, cooked up by the test program.   The controller's Index method runs, picks up the fake services object, gets from it the mock repository and calls its FindAll method.  That returns the cat list as instructed.

When the Index method returns, the test checks that the objects that it supplied are now in the state that it expects.  Behind the scenes, the mock checks that the FindAll method was actually called.  If all is not well, the test fails. 

Other tests make the mock return other values, and check that the results are as expected.


Installing MySQL
============

The free MySQL system provides a relational database server.
Download it from [here](http://dev.mysql.com/downloads/).
You need the community server and the tools.
The tools work from a command window, so
you might also find the workbench useful.
It provides a visual interface to manage your databases.

If you use the standard installation, the server will start up automatically and run in the background.

Once MySQL is installed, create a database.
To do this using the command-line tools, start a command window
and use the MySQL client.
In this example I create a database called "animals" that can be accessed by the MySQL user "webuser" using the password "secret":

    mysql -u root -p
    {type the root password that you set when you installed mysql}
    
    mysql> create database animals;
    mysql> grant all on animals.* to 'webuser' identified by 'secret';
    mysql> quit

The user has all access rights, so it can create tables.
The web server generated by the scaffolder can be configured to connect to this database.


Setting Up the Go Tools
=======================

The installation instructions for the Go tools are 
[here](https://golang.org/doc/install).

That document shows several ways to install Go.
It's a very good idea to install it from source code, but that's not the simplest option.
If you are running under Windows and you just want to get started quickly,
the MSI installer is simpler, but there may be a cost to going down that route later.
On every new release of Go, you will have to wait for the MSI installer to become available.
If you want to keep up-to-date with latest releases, you should follow the source code option.

Once you have installed Go, you need to read the document
[How to Write Go Code](https://golang.org/doc/code.html).

As that document explains, you can keep all your Go projects together in one directory (AKA folder),
and that needs to be defined by an environment variable GOPATH.
I'm going to assume that you will use a directory called goprojects in your home directory.

My user name is simon so under Linux my home directory is /home/simon.
Under Windows it's C:\Users\simon.

In the examples below, "$ command"
represents a command that you issue in a Linux command window.
The "$" represents the prompt that the command window gives you.
Type the command (without the dollar)
and press the enter key to run it.

Using your command window, create your project directory:

    $ mkdir goprojects

Set the GOPATH environment variable
and add the workspace's bin subdirectory to your PATH.

Under Linux:

    $ export GOPATH=/home/{your user name}/goprojects
    $ export PATH=$PATH:$GOPATH/bin

In my case, that would be:

    $ export GOPATH=/home/simon/goprojects
    $ export PATH=$PATH:$GOPATH/bin

That only sets those variables for that command window.
You should also add the same commands to the file .profile in your home directory so that they are run every time you log in and apply to all command windows.

 On Windows 7 go to the control panel in the Start menu. Choose System, then System Security.
In the left-hand menu, choose Advanced System Settings.
That produces a pop-up window with a button marked Environment Settings.
Press that and the Environment Variables window appears.
In the list of user variables at the top, there should be one called Path.
It's a list of directories separated by semicolons.

Use the New button to create an environment variable called GOPATH set to  "c:\Users\\{your user name}\goprojects".  In my case, that's "c:\Users\simon\goprojects".

Still in the Environment Variables window, find the user variable Path and edit it.
Run to the end of the line (it may be quite long) and add a semicolon followed by the GOPATH bin directory:

    ;c:\Users\{your user name}\goprojects\bin

and press OK to make the change.

Given that you have already created a variable GOPATH containing most of that text, you could instead add this to the Path:

    ;%GOPATH%\bin

Either way, don't forget the semicolon.

Those variables will be available to all command windows that you open from now on,
but NOT to any that you already have open.

Don't get confused between the PATH and the GOPATH variable.
PATH tells the system (Windows, Linux or whatever) where to find
executable programs.
GOPATH tells the Go tools where to find Go projects.

The Index method continues running, and eventually returns.  The test then checks the resulting data to make sure that the method has set it up as expected.  Behind the scenes (and automatically) the mock repository also checks that the FindAll method was called  and raises an error if it wasn't.  If all is well,the test is marked as passed and the test framework continues to the next one.

Some other people prefer to avoid using mock objects for testing, instead making the test program manufacture all the test objects itself.  There are advantages and disadvantages to both approaches, and entire books on the subject.
The most important thing is to test your software thoroughly as you are writing it.

