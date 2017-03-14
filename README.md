# Scaffolder
Given a description of a database and its tables, the Goblimey Scaffolder creates that database
and generates a Go web server that provides the Create, Read, Update and Delete (CRUD) operations
on it.
The server is designed according to the Model, View, Controller (MVC) architecture and is
implemented using RESTful requests.
The result is presented as a complete prototype Go project with all source code included
along with unit and integration tests to check that the whole thing hangs together.

The idea of the scaffolder is taken from the Ruby-on-Rails scaffold generator.

The scaffolder has all sorts of uses, the most obvious being a prototyping tool
for designing database tables.
Producing the right design for the database that sits behind a web site usually takes several attempts.
The scaffolder gives you a quick and easy way to experiment with different versions.
Once you have a working version, you can extend the source code and produce your own production web server.
That's much easier than producing one from scratch.

Software testers may also find the scaffolder useful.
Testers often have to create carefully-crafted database content
to drive a set of tests.
The scaffolder provides a bespoke tool for doing that.

Producing a complete piece of working source code also makes the scaffolder
a very useful aid to learning Go.
That means that it may be used by people who are new to Go and
possibly new to programming.
This document assumes a fair amount of specialist knowledge.
[These notes](http://goblimey.com/scaffolder/).
describes the scaffolder in at a gentler pace and 
covers basic issues such as installing Go and MySQL.

In this version
the scaffolder doesn't handle relations between tables. 
That is a serious omission
which I plan to fix in a future version.


For the Impatient
============

Get the dependencies and install the scaffolder:
 
    $ go get github.com/go-sql-driver/mysql
    $ go get gopkg.in/gorp.v1
    $ go get github.com/emicklei/go-restful
    $ go get github.com/onsi/gomega
    $ go get golang.org/x/tools/cmd/goimports
    $ go get github.com/petergtz/pegomock/pegomock
    $ go get github.com/goblimey/scaffolder

By default, "go get" doesn't update any projects that you have already downloaded.
If you downloaded any of those projects a long time ago, 
you may wish to update it to the latest version using the -u flag, for example:

    go get -u github.com/petergtz/pegomock/pegomock	
	
Once you have downloaded the scaffolder, you can find an example table specification in the examples directory. 
You can use this to create a simple web server like so:
 
* create an empty database
* create a Go project directory and cd to it
* Copy the example specification file to this directory and call it "scaffold.json"
* $ scaffolder
* $ ./install.sh
* $ animals        # start your web server
* in your web browser, navigate to <http://localhost:4000>
* create some cats and dogs


Creating Your Project
================

The How to Write Go Code document 
(which you can find [here](https://golang.org/doc/code.html))
suggests that you structure your project as if you are going to store it in a repository,
even if you don't ever store it there.

I'm going to assume that you will use the GitHub
to store your project.
If your GitHub account was called alunsmithie,
your home page on the github would be 
https://github.com/alunsmithie.
If your project is called 'animals'
then it would be stored in
https://github.com/alunsmithie/animals.

If you follow this stucture
but you don't want to put the result on the Github,
you can just create a directory to hold your project -
on Linux it's the directory:

    $GOPATH/src/github.com/alunsmithie/animals

on Windows it's:

    %GOPATH%\src\github.com\alunsmithie\animals

and you can skip the rest of this section.

If you are actually going to store
your project on the GitHub, rather than just structuring it so that you could,
it's much easier to create an empty project first
and add files later.
Use the '+' button at the top of your github home page
to create the project.
In this example,
it's called 'animals'

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

That creates a Go project directory "animals" 
which is also a local Git repository.
As you create files
you can add, commit and push them.


The JSON Specification
======================

The scaffolder is driven by a text file in JSON) format that specifies a database and a set of tables.

When you are writing JSON, it's very easy to make a simple mistake such as missing out a comma.
The resulting error messages may not be very helpful.
You will save yourself a lot of pain if you prepare the file 
using an editor that understands JSON and warns you about obvious errors.
Most Integrated development Environments (liteIDE, Eclipse, IntelliJ, VSCode etc) have editors that will do this.  Text editors such as Windows Notepad++ will do the same.

The scaffolder includes an example specification file so you can use that for a quick experiment.
Copy goprojects/scaffolder/examples/animals.scaffold.json into your project directory and rename it scaffold.json.

The example specification defines a MySQL database called "animals" containing tables "cats" and "mice":

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

In the example, the first few lines of the JSON define the project and its database.
In this case the project is the one we created earlier - animals .
The resulting server controls a MySQL database called "animals".

The sourcebase defines the location of the project.
In this example the sourcebase is "github.com/alunsmithie/animals",
so the project is stored in
src/github.com/alunsmithie/animals within your workspace.
You created that directory in the previous section.

When the scaffolder creates files, it creates them within this directory.

The database definition specifies the user name and password
("webuser" and "secret" in this example)
and the name of the database server and the port that it is listening on.
In this case the server machine is "localhost" (this computer) 
and the MySQL server is listening on its default port.
(If not you can specify the port  like so: "dbport": "1234".)

Go supports a number of Object-Realational Mapping (ORM) tools
to manage the connection with a datbase.
The ORM value says which one to use.  
At present
the only one supported is [GORP](https://github.com/coopernurse/gorp) version 1.
I plan to add support for other ORMs in the future.

The Resources section defines a list of resources.
Each resource definition produces a database table, a model, a repository, 
a controller and a set of views.
This example describes the "cat" resource and the "mouse" resource supported by the table with the same name as its resource.

Traditionally, database tables are named using the plural of the data that they contain.
By default the scaffolder just takes the name of the resource and adds an "s"
so the table for the cat resource is called "cats".
If that won't do, you can specify the table name like so:

    "name": "mouse", "plural": "mice",

Each resource section defines a list of fields.  The cat resource has fields "name" and "breed" which contain strings,
"age" containing an integer
"weight" containing a floating point number
and "chipped" containing a boolean value,
recording whether or not the cat has been microchipped.
The "chipped" field is optional by default.
The rest are marked as mandatory.

The mouse resource has just two fields,
"name" which is mandatory and "breed" which is optional.
Both contain strings.

Given this JSON spec, 
the scaffolder generates a set of unit and integration test programs to check that the generated source code works properly.
A unit test takes a module of the source code and runs it in isolation, supplying it with test values and checking that the module produces the expected result.  An integration tests is similar, but checks that a set of modules work together properly.
Each field in the JSON can have an optiona; list of testValues to be used by the tests.
If you don't specify an test values, they are all generated automatically. 
If you don't specify enough, the rest are generated automatically.
If you specify too many, the extra ones are ignored.
Currently none of the the generated tests use more than two values,
so a list of two values is always sufficient. 

The optional excludeFromDisplay value in the JSON 
controls the contents of the display label.
This identifies each database record in the generated web pages
and it's used in all sorts of ways.
For example,
the index page shows a list of all records in the table.
It uses the display label to represent each record.
By default the display label contains the values of all the fields,
so if no fields were excluded,
a record in the index page for cats would look something something like this:

    1 Tommy Siamese 2 5 true

If there are a lot of fields the label can become unwieldy,
Excluding some of them from the label
makes it more manageable.
In the cats resource in the example,
the fields "age", "weight" and "chipped" are excluded,
so the display label will be something like:

    1 Tommy Siamese

If you view the HTML for the index page,
you can see that it's a series of links,
one to show each record and one to edit each record.
Each link has a unique ID,
made up using the display label:

    <td>
	    <a id='LinkToShow 1 Tommy Siamese'  href='/cats/1'>1 Tommy Siamese</a>
    </td>
    <td>
	    <a id='LinkToEdit 1 Tommy Siamese' href='/cats/1/edit'>Edit </a>
    </td>

Giving each of the the objects on the page a unique ID
makes it easier to test the solution using 
web testing tools such as a Selenium.
	
If you press the edit button
and then view the HTML for that page,
you can see that
the title and the h3 heading are also made from the display label:

    <!DOCTYPE html>
    <html lang="en">
        <head>
            <title>Edit Cat 1 Tommy Siamese</title>
            <link href='/stylesheets/scaffold.css' rel='stylesheet'/>
        </head>
        <body>
            <h2>Animals</h2>
    	       <h3>Edit Cat 1 Tommy Siamese</h3>


Creating a Database
==================

The JSON in the previous section expects a database called "animals" which can be accessed by the MySQL user "webuser" using the password "secret".  
Before you run the generated web server for the first time
you need to create an empty database and give the user access rights:

Run the MySQL client in a command window:

    mysql -u root -p
    {type the root password that you set when you installed mysql}
    
    mysql> create database animals;
    mysql> grant all on animals.* to 'webuser' identified by 'secret';
    mysql> quit

The web server will connect to this database
and create the tables if they don't already exist.
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

When you create a record,
its ID field will be set automatically to a unique value.

Building the Server
======================

When you run the scaffolder, by default it looks for a specification file "scaffold.json" in the current directory - something like the example above.  You can specify a different file if you want to.

By default the scaffolder generates the server in the current directory, which should be your github project directory (in the example, goprojects/src/github.com/alunsmithie/animals).
Alternatively you can run it from another directory and tell it where to find the project directory.

In your command window, change directory to your project and run the scaffolder:

    $ cd $HOME/goprojects/src/github.com/alunsmithie/animals
    $ scaffolder

That creates the web server source code and some scripts.

To use a different specification file:

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
The generated script install.sh builds and installs the server on Linux:

    $ ./install.sh

install.bat does the same on Windows:

    install

There is also test.sh and test.bat.  
These run the tests to ensure that all the generated parts work properly:

    $ ./test.sh

If all the tests pass, you can start the web server.  



Running the Server
==================

If your Go bin directory goprojects/bin is in your path, 
you can run your server like so:

     $ animals

or you can run it in verbose mode and see tracing messages in your command window:

     $ animals -v

The first time you run the server it will create the database tables.
(Assuming that you have created an empty database
and permitted the web server's user to create tables
as described earlier.)

The server runs on port 4000.  In your web browser, navigate to <http://localhost:4000>

That display the home page.  It has two links "Manage cats" and "Manage mice".
The first takes you to the index page for the cat resource.
The cats table is currently empty.  Use the Create button to create some.

Once you've done that, 
the index page lists the cats with links and buttons 
to edit and delete the records, and a link back to the home page.

To add some mice, use the link to the home page and then the "Manage Mice" link.

To stop the server, type ctrl/c in the command window.  (Hold down the ctrl key and type a single "c", you don't need to press the enter key.)


Changing the JSON
==================

The scaffolder creates these files

* install.sh - a shell script to build the animals server
* install.bat batch script to do the same on Windows
* test.sh - a shell script to run the test suite
* test.bat same for Windows
* animals.go - the source code of the main module
* generated - the source code of the models, views, controllers, repositories and support software
* views - the templates used to create the html views.

You can edit the JSON and add some fields.  For example, you could add a field "favouritefood" to the cats table.  Run the scaffolder again and it will produce a new version of the server.  Run the install script to build and install it.

It's assumed that you may want to tweak things like the build scripts, the main program, the home page  and so on.  If you run the scaffolder over this project again, by default only the stuff in the "generated" directories is overwritten.  

If you run the scaffolder with the overwrite option, it replaces everything:

    $scaffolder --overwrite

The server only creates the database tables
if they are missing,
so if you change the JSON and add some fields,
they won't be added to the database tables.
You can add the extra fields to the tables using the MySQL client  
or you can simply drop the tables
and then restart the server.
It will create any missing tables using the new specification,
but they will be empty.
If you have created a lot of test data 
you might want to use the first option
of adding the fields by hand,
or maybe create a new project connected to a
different database.

If you change the JSON it's a good idea to run the tests again to make sure that nothing has been broken.  However, some of the integration tests write to the database and they will also trash any existing data if you run them.  If you want to avoid that, you can run just the unit tests:

    $ ./test.sh unit



MVC
=====================================

Given a description of a database, the scaffolder writes a Go program that creates the database
and provides a web server that allows you to create, read, update and delete records.  The web server builds HTML pages to order, depending on the data in the database.  Such a server is sometimes called a web application server, to distinguish it from a web server that simply feeds out static pages.

The web application server provides controlled access to data in the database. Each response is manufactured to order, based on the data. In a production system, it's usually impossible for a user to access the database directly.  They can only do it via the web server and they can only do what the web server allows.

The web server generated by the scaffolder
is designed using Model, View, Controller architecture(MVC). 
The benfits of MVC include:

* Isolation of business logic from the user interface
* Ease of keeping code DRY
* Clarifying where different types of code belong for easier maintenance

(DRY means Don't Repeat Yourself - don't write the same source code more than once.) 

A model represents the information in the database
and the rules to manipulate that data. 
In this case, models are used to manage the interaction with a corresponding database table. 
Each table in your database will correspond to one model in your application. 
There is also a corresponding Repository (AKA a Data Access Object (DAO)),
which is software used to fetch and store data.
(Not to be confused with the GIT repository.)

The Views represent the user interface of your application. 
They handle the presentation of the data. Views provide data to the web browser or other tool that is used to make requests from your application.
This web server generates an HTML response page for each request on the fly from a template, so its views are Go HTML templates.

The controller provides the “glue” between models and views.  The controller is responsible for processing the incoming requests from the web browser, interrogating the models for data, and passing that data on to the views for presentation.  There is one controller for each table.


Restful Requests
================

The web server created by the scaffolder handles requests
that conform to the
REpresentational State Transfer (REST) model.
REST is described [here](https://en.wikipedia.org/wiki/Representational_state_transfer).

The REST model stresses the use of resource identifiers such as URIs to represent resources.
In the generated web server, each resource has an associated database table
with the same name.
If we have a database table called "cats" then we have web resource called "cats".

    GET /cats
    GET /cats/

are both RESTful HTTP request to display a list of all cat resources
(all records in the cats table)
This is the "index" page for the cat resource.

    GET /cats/42

is a RESTful HTTP request to display the data for the cat with ID 42.

    DELETE /cats/42

is a RESTful HTTP request to delete that record.

REST requires that all GET requests are idempotent ("having equal effect").
This simply means that if some browsers issue the same GET request many times and
the data in the database is not changed by some other agent, 
each request will produce the same response.

The upshot of that is that we use GET requests to read data 
from the database and other requests (PUT, DELETE etc) to change the data.

One advantage of the REST approach is that search engines respect this rule.  If your web site is public it will be crawled repeatedly by lots of search engines.  When a search engine crawls a site it attempts to visit all the pages.  It scans the home page and looks through it for HTTP requests to other pages.  Then it scans those pages and so on.  If it finds a GET request,
it will attempt to issue it,
but it will avoid issuing any other requests that it finds.  The crawler assumes that a GET requests won't change your data but any other request might.

REST requires that requests containing parameters are only used to submit form data. For example if this request deletes the record with ID 42:

    GET /cats?operation=delete&id=42

it doesn't follow the REST rules, firstly because it's using a GET request to change the data and secondly because it uses parameters but it's not carrying form data.

There's no REST policeman that will stop your web server using non-Restful requests.
REST is just a set of rules that # Scaffolder
Given a description of a database and its tables, the Goblimey Scaffolder creates that database
and generates a Go web server that provides the Create, Read, Update and Delete (CRUD) operations
on it.
The server is designed according to the Model, View, Controller (MVC) architecture and is
implemented using RESTful requests.
The result is presented as a complete prototype Go project with all source code included
along with unit and integration tests to check that the whole thing hangs together.

The idea of the scaffolder is taken from the Ruby-on-Rails scaffold generator.

The scaffolder has all sorts of uses, the most obvious being a prototyping tool
for designing database tables.
Producing the right design for the database that sits behind a web site usually takes several attempts.
The scaffolder gives you a quick and easy way to experiment with different versions.
Once you have a working version, you can extend the source code and produce your own production web server.
That's much easier than producing one from scratch.

Software testers may also find the scaffolder useful.
Testers often have to create carefully-crafted database content
to drive a set of tests.
The scaffolder provides a bespoke tool for doing that.

Producing a complete piece of working source code also makes the scaffolder
a very useful aid to learning Go.
That means that it may be used by people who are new to Go and
possibly new to programming.
This document assumes a fair amount of specialist knowledge.
[These notes](http://goblimey.com/scaffolder/).
describes the scaffolder in at a gentler pace and 
covers basic issues such as installing Go and MySQL.

In this version
the scaffolder doesn't handle relations between tables. 
That is a serious omission
which I plan to fix in a future version.


For the Impatient
============

Get the dependencies and install the scaffolder:
 
    $ go get github.com/go-sql-driver/mysql
    $ go get gopkg.in/gorp.v1
    $ go get github.com/emicklei/go-restful
    $ go get github.com/onsi/gomega
    $ go get golang.org/x/tools/cmd/goimports
    $ go get github.com/petergtz/pegomock/pegomock
    $ go get github.com/goblimey/scaffolder

By default, "go get" doesn't update any projects that you have already downloaded.
If you downloaded any of those projects a long time ago, 
you may wish to update it to the latest version using the -u flag, for example:

    go get -u github.com/petergtz/pegomock/pegomock	
	
Once you have downloaded the scaffolder, you can find an example table specification in the examples directory. 
You can use this to create a simple web server like so:
 
* create an empty database
* create a Go project directory and cd to it
* Copy the example specification file to this directory and call it "scaffold.json"
* $ scaffolder
* $ ./install.sh
* $ animals        # start your web server
* in your web browser, navigate to <http://localhost:4000>
* create some cats and dogs


Creating Your Project
================

The How to Write Go Code document 
(which you can find [here](https://golang.org/doc/code.html))
suggests that you structure your project as if you are going to store it in a repository,
even if you don't ever store it there.

I'm going to assume that you will use the GitHub
to store your project.
If your GitHub account was called alunsmithie,
your home page on the github would be 
https://github.com/alunsmithie.
If your project is called 'animals'
then it would be stored in
https://github.com/alunsmithie/animals.

If you follow this stucture
but you don't want to put the result on the Github,
you can just create a directory to hold your project -
on Linux it's the directory:

    $GOPATH/src/github.com/alunsmithie/animals

on Windows it's:

    %GOPATH%\src\github.com\alunsmithie\animals

and you can skip the rest of this section.

If you are actually going to store
your project on the GitHub, rather than just structuring it so that you could,
it's much easier to create an empty project first
and add files later.
Use the '+' button at the top of your github home page
to create the project.
In this example,
it's called 'animals'

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

That creates a Go project directory "animals" 
which is also a local Git repository.
As you create files
you can add, commit and push them.


The JSON Specification
======================

The scaffolder is driven by a text file in JSON) format that specifies a database and a set of tables.

When you are writing JSON, it's very easy to make a simple mistake such as missing out a comma.
The resulting error messages may not be very helpful.
You will save yourself a lot of pain if you prepare the file 
using an editor that understands JSON and warns you about obvious errors.
Most Integrated development Environments (liteIDE, Eclipse, IntelliJ, VSCode etc) have editors that will do this.  Text editors such as Windows Notepad++ will do the same.

The scaffolder includes an example specification file so you can use that for a quick experiment.
Copy goprojects/scaffolder/examples/animals.scaffold.json into your project directory and rename it scaffold.json.

The example specification defines a MySQL database called "animals" containing tables "cats" and "mice":

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

In the example, the first few lines of the JSON define the project and its database.
In this case the project is the one we created earlier - animals .
The resulting server controls a MySQL database called "animals".

The sourcebase defines the location of the project.
In this example the sourcebase is "github.com/alunsmithie/animals",
so the project is stored in
src/github.com/alunsmithie/animals within your workspace.
You created that directory in the previous section.

When the scaffolder creates files, it creates them within this directory.

The database definition specifies the user name and password
("webuser" and "secret" in this example)
and the name of the database server and the port that it is listening on.
In this case the server machine is "localhost" (this computer) 
and the MySQL server is listening on its default port.
(If not you can specify the port  like so: "dbport": "1234".)

Go supports a number of Object-Realational Mapping (ORM) tools
to manage the connection with a datbase.
The ORM value says which one to use.  
At present
the only one supported is [GORP](https://github.com/coopernurse/gorp) version 1.
I plan to add support for other ORMs in the future.

The Resources section defines a list of resources.
Each resource definition produces a database table, a model, a repository, 
a controller and a set of views.
This example describes the "cat" resource and the "mouse" resource supported by the table with the same name as its resource.

Traditionally, database tables are named using the plural of the data that they contain.
By default the scaffolder just takes the name of the resource and adds an "s"
so the table for the cat resource is called "cats".
If that won't do, you can specify the table name like so:

    "name": "mouse", "plural": "mice",

Each resource section defines a list of fields.  The cat resource has fields "name" and "breed" which contain strings,
"age" containing an integer
"weight" containing a floating point number
and "chipped" containing a boolean value,
recording whether or not the cat has been microchipped.
The "chipped" field is optional by default.
The rest are marked as mandatory.

The mouse resource has just two fields,
"name" which is mandatory and "breed" which is optional.
Both contain strings.

Given this JSON spec, 
the scaffolder generates a set of unit and integration test programs to check that the generated source code works properly.
A unit test takes a module of the source code and runs it in isolation, supplying it with test values and checking that the module produces the expected result.  An integration tests is similar, but checks that a set of modules work together properly.
Each field in the JSON can have an optiona; list of testValues to be used by the tests.
If you don't specify an test values, they are all generated automatically. 
If you don't specify enough, the rest are generated automatically.
If you specify too many, the extra ones are ignored.
Currently none of the the generated tests use more than two values,
so a list of two values is always sufficient. 

The optional excludeFromDisplay value in the JSON 
controls the contents of the display label.
This identifies each database record in the generated web pages
and it's used in all sorts of ways.
For example,
the index page shows a list of all records in the table.
It uses the display label to represent each record.
By default the display label contains the values of all the fields,
so if no fields were excluded,
a record in the index page for cats would look something something like this:

    1 Tommy Siamese 2 5 true

If there are a lot of fields the label can become unwieldy,
Excluding some of them from the label
makes it more manageable.
In the cats resource in the example,
the fields "age", "weight" and "chipped" are excluded,
so the display label will be something like:

    1 Tommy Siamese

If you view the HTML for the index page,
you can see that it's a series of links,
one to show each record and one to edit each record.
Each link has a unique ID,
made up using the display label:

    <td>
	    <a id='LinkToShow 1 Tommy Siamese'  href='/cats/1'>1 Tommy Siamese</a>
    </td>
    <td>
	    <a id='LinkToEdit 1 Tommy Siamese' href='/cats/1/edit'>Edit </a>
    </td>

Giving each of the the objects on the page a unique ID
makes it easier to test the solution using 
web testing tools such as a Selenium.
	
If you press the edit button
and then view the HTML for that page,
you can see that
the title and the h3 heading are also made from the display label:

    <!DOCTYPE html>
    <html lang="en">
        <head>
            <title>Edit Cat 1 Tommy Siamese</title>
            <link href='/stylesheets/scaffold.css' rel='stylesheet'/>
        </head>
        <body>
            <h2>Animals</h2>
    	       <h3>Edit Cat 1 Tommy Siamese</h3>


Creating a Database
==================

The JSON in the previous section expects a database called "animals" which can be accessed by the MySQL user "webuser" using the password "secret".  
Before you run the generated web server for the first time
you need to create an empty database and give the user access rights:

Run the MySQL client in a command window:

    mysql -u root -p
    {type the root password that you set when you installed mysql}
    
    mysql> create database animals;
    mysql> grant all on animals.* to 'webuser' identified by 'secret';
    mysql> quit

The web server will connect to this database
and create the tables if they don't already exist.
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

When you create a record,
its ID field will be set automatically to a unique value.

Building the Server
======================

When you run the scaffolder, by default it looks for a specification file "scaffold.json" in the current directory - something like the example above.  You can specify a different file if you want to.

By default the scaffolder generates the server in the current directory, which should be your github project directory (in the example, goprojects/src/github.com/alunsmithie/animals).
Alternatively you can run it from another directory and tell it where to find the project directory.

In your command window, change directory to your project and run the scaffolder:

    $ cd $HOME/goprojects/src/github.com/alunsmithie/animals
    $ scaffolder

That creates the web server source code and some scripts.

To use a different specification file:

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
The generated script install.sh builds and installs the server on Linux:

    $ ./install.sh

install.bat does the same on Windows:

    install

There is also test.sh and test.bat.  
These run the tests to ensure that all the generated parts work properly:

    $ ./test.sh

If all the tests pass, you can start the web server.  



Running the Server
==================

If your Go bin directory goprojects/bin is in your path, 
you can run your server like so:

     $ animals

or you can run it in verbose mode and see tracing messages in your command window:

     $ animals -v

The first time you run the server it will create the database tables.
(Assuming that you have created an empty database
and permitted the web server's user to create tables
as described earlier.)

The server runs on port 4000.  In your web browser, navigate to <http://localhost:4000>

That display the home page.  It has two links "Manage cats" and "Manage mice".
The first takes you to the index page for the cat resource.
The cats table is currently empty.  Use the Create button to create some.

Once you've done that, 
the index page lists the cats with links and buttons 
to edit and delete the records, and a link back to the home page.

To add some mice, use the link to the home page and then the "Manage Mice" link.

To stop the server, type ctrl/c in the command window.  (Hold down the ctrl key and type a single "c", you don't need to press the enter key.)


Changing the JSON
==================

The scaffolder creates these files

* install.sh - a shell script to build the animals server
* install.bat batch script to do the same on Windows
* test.sh - a shell script to run the test suite
* test.bat same for Windows
* animals.go - the source code of the main module
* generated - the source code of the models, views, controllers, repositories and support software
* views - the templates used to create the html views.

You can edit the JSON and add some fields.  For example, you could add a field "favouritefood" to the cats table.  Run the scaffolder again and it will produce a new version of the server.  Run the install script to build and install it.

It's assumed that you may want to tweak things like the build scripts, the main program, the home page  and so on.  If you run the scaffolder over this project again, by default only the stuff in the "generated" directories is overwritten.  

If you run the scaffolder with the overwrite option, it replaces everything:

    $scaffolder --overwrite

The server only creates the database tables
if they are missing,
so if you change the JSON and add some fields,
they won't be added to the database tables.
You can add the extra fields to the tables using the MySQL client  
or you can simply drop the tables
and then restart the server.
It will create any missing tables using the new specification,
but they will be empty.
If you have created a lot of test data 
you might want to use the first option
of adding the fields by hand,
or maybe create a new project connected to a
different database.

If you change the JSON it's a good idea to run the tests again to make sure that nothing has been broken.  However, some of the integration tests write to the database and they will also trash any existing data if you run them.  If you want to avoid that, you can run just the unit tests:

    $ ./test.sh unit

