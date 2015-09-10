package main

import (
	"log"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

var mongoUrl string = "localhost:27017"
var dbName string = "test"

type MongoDatastore struct {}

func (db *MongoDatastore) StoreTemplate(template Template) string {
    session, err := mgo.Dial(mongoUrl)
    check(err)
    defer session.Close()

    // Optional. Switch the session to a monotonic behavior.
    session.SetMode(mgo.Monotonic, true)

    objectId := bson.NewObjectId()
    template.Id = objectId
    c := session.DB(dbName).C("template")
    err = c.Insert(&template)
    if err != nil {
            log.Fatal(err)
    }

    return objectId.String()
}

func (db *MongoDatastore) RetrieveTemplate(id string) Template {
    session, err := mgo.Dial(mongoUrl)
    check(err)
    defer session.Close()

    objectId := bson.ObjectIdHex(id)
    // Optional. Switch the session to a monotonic behavior.
    session.SetMode(mgo.Monotonic, true)
    c := session.DB(dbName).C("message")
    
    result := Template{}
    err = c.Find(bson.M{"_id": objectId}).One(&result)
    if err != nil {
            log.Fatal(err)
    }

    return result
}

func (db *MongoDatastore) StoreContact(contact Contact) string {
    session, err := mgo.Dial(mongoUrl)
    check(err)
    defer session.Close()

    // Optional. Switch the session to a monotonic behavior.
    session.SetMode(mgo.Monotonic, true)

    objectId := bson.NewObjectId()
    contact.Id = objectId
    c := session.DB(dbName).C("contact")
    err = c.Insert(&contact)
    if err != nil {
            log.Fatal(err)
    }

    return objectId.String()
}

func (db *MongoDatastore) RetrieveContact(id string) Contact {
    session, err := mgo.Dial(mongoUrl)
    check(err)
    defer session.Close()

    // Optional. Switch the session to a monotonic behavior.
    session.SetMode(mgo.Monotonic, true)

    objectId := bson.ObjectIdHex(id)
    c := session.DB(dbName).C("contact")

    result := Contact{}
    err = c.Find(bson.M{"_id": objectId}).One(&result)
    if err != nil {
            log.Fatal(err)
    }

    return result
}

func (db *MongoDatastore) Status() bool {
    return true
}
