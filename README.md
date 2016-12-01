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
