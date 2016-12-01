# scaffolder
Given a description of a DB table, builds a website that provides CRUD operations.  Based on the Ruby-on-Rails scaffold generator.

Fetching and Building the Scaffolder
======================================

First, get the dependencies:

```
go get github.com/go-sql-driver/mysql
go get gopkg.in/gorp.v1
go get github.com/emicklei/go-restful
go get github.com/petergtz/pegomock/pegomock
```

Note: by default, go get does not update anything that you have already downloaded.  If you downloaded any of those packages a long time ago, you may wish to update them to the latest version using the -u flag, for example:

go get -u github.com/petergtz/pegomock/pegomock

Then get the scaffolder:

go get github.com/goblimey/scaffolder

Building a CRUD server
======================

The scaffolder is driven by a text file containing JSON that specifies a database and a set of tables.
If the JSON in the file is illegal, the messages that you get are not very helpful.  
You will save yourself a lot of pain if you prepare the file 
using an editor that understands JSON and warns you if you get anything wrong.
Most Integrated development Environments (liteIDE, Eclipse, IntelliJ, VSCode) understand JSON,
as do many text editors including VIM.
For MS Windows there is also Notepad++.

This specification defines a MySQL database called "animals" containing tables "cats" and "mice":

```
{
    "name": "animals",
    "sourceBase": "github.com/goblimey/animals",
    "db": "mysql",
    "dblogin": "webuser:secret@tcp(localhost:3306)/animals",
    "dbuser": "webuser",
    "dbpassword": "secret",
    "orm": "gorp",
    "Resources": [
        {
            "name": "cat",
            "fields": [
                {
                    "name": "name",
                    "type": "string",
                    "mandatory": true,
                    "testValues": [
                        "a",
                        "b"
                    ]
                },
                {
                    "name": "breed",
                    "type": "string",
                    "mandatory": true
                },
                {
                    "name": "age",
                    "type": "int",
                    "mandatory": true,
		    "excludeFromDisplay": true
                },
				{
                    "name": "weight",
                    "type": "float",
                    "mandatory": true,
		    "excludeFromDisplay": true
                },
				{
                    "name": "chipped",
                    "type": "bool",
		    "excludeFromDisplay": true
                }
            ]
        },
        {
            "name": "mouse",
            "plural": "mice",
            "fields": [
                {
                    "name": "name",
                    "type": "string",
                    "mandatory": true
                },
                {
                    "name": "breed",
                    "type": "string",
		    "excludeFromDisplay": true
                }
            ]
        }
    ]
}
```

Most of that is fairly obvious.  I will describe the stuff that isn't.

Go source code is laid out in a hierarchy of directories.
It's assumed that you will commit the source code to a repository such as github.com,
so the hierarchy starts with the name of the project in the repository -
I have a github account "goblimey" and my project is called "animals" 
so my sourceBase in the json is "github.com/goblimey/animals".
When the scaffolder produces the Go source code for my project,
most of it is within the directory "github.com/goblimey/animals".

I assume that you want to use the stuff created by the scaffolder as a starting point
and create your own source code, so it creates stuff in the "github.com/goblimey/animals/generated" directory.
ANYTHING IN THERE WILL BE TRASHED NEXT TIME YOU RUN THE SCAFFOLDER
so you should create your own stuff outside of that directory.


