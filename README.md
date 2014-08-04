Google App Engine Billing System
=====================================================

*This app is not ready to be used.*

This app is designed to use google user service for login.  Blobstore to store PDF bills that can be sent out.  It uses appengine datastore to store users, companies and meta data about the bills.  Bills can be marked paid and reconciled.  

This is mainly just a basic attempt to use google app engine and its various services to build a tracking system we use at Pioneer Valley Books.

To get started:

* [Read about setting up GAE dev environment in go](https://developers.google.com/appengine/docs/go/gettingstarted/devenvironment)
* git clone git@github.com:nickdufresne/GAE-Billing-System.git
* ensure goapp bin is in your path
* goapp serve GAE-Billing-System
* navigate to [localhost:8080](http://localhost:8080/)
