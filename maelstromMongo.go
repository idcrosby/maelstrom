package main

import (
    "google.golang.org/cloud/compute/metadata"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

var mongoUrl string = "localhost:27017"
var dbName string = "test"

type MongoDatastore struct {}

func (db *MongoDatastore) Ping() bool {
    if gce {
        mongoUrl, _ = metadata.InstanceAttributeValue("mongoUrl")
        if Debug {
            InfoLog.Println("Mongo URL pulled from GCE Metadata: " + mongoUrl)
        }
    }
    session, err := mgo.Dial(mongoUrl)
    if err != nil {
        return false
    }
    session.Close()
    return true
}

func (db *MongoDatastore) StoreContact(contact Contact) Contact {
    session, err := mgo.Dial(mongoUrl)
    check(err)
    defer session.Close()

    // Optional. Switch the session to a monotonic behavior.
    // session.SetMode(mgo.Monotonic, true)

    contact.Id = bson.NewObjectId()
    c := session.DB(dbName).C("contact")
    err = c.Insert(&contact)
    if err != nil {
        ErrorLog.Println("Error storing Contact: " + contact.Name)
        return Contact{}
    }

    return contact
}

func (db *MongoDatastore) UpdateContact(contact Contact) Contact {
    session, err := mgo.Dial(mongoUrl)
    check(err)
    defer session.Close()

    // Optional. Switch the session to a monotonic behavior.
    // session.SetMode(mgo.Monotonic, true)

    c := session.DB(dbName).C("contact")
    err = c.UpdateId(contact.Id, &contact)
    if err != nil {
        return Contact{}
    }

    return contact
}

// func (db *MongoDatastore) RetrieveContact(id string) Contact {
//     session, err := mgo.Dial(mongoUrl)
//     check(err)
//     defer session.Close()

//     // Optional. Switch the session to a monotonic behavior.
//     session.SetMode(mgo.Monotonic, true)

//     objectId := bson.ObjectIdHex(id)
//     c := session.DB(dbName).C("contact")

//     result := Contact{}
//     err = c.Find(bson.M{"_id": objectId}).One(&result)
//     if err != nil {
//             log.Fatal(err)
//     }

//     return result
// }

// func (db *MongoDatastore) RetrieveContactsByTag(tag string) []Contact {
//     session, err := mgo.Dial(mongoUrl)
//     check(err)
//     defer session.Close()

//     // Optional. Switch the session to a monotonic behavior.
//     session.SetMode(mgo.Monotonic, true)

//     c := session.DB(dbName).C("contact")

//     result := []Contact{}
//     err = c.Find(bson.M{"tag": tag}).All(&result)
//     if err != nil {
//             log.Fatal(err)
//     }

//     return result
// }

func (db *MongoDatastore) RetrieveContactsBy(param string, value string) []Contact {
    session, err := mgo.Dial(mongoUrl)
    check(err)
    defer session.Close()

    // Optional. Switch the session to a monotonic behavior.
    // session.SetMode(mgo.Monotonic, true)

    c := session.DB(dbName).C("contact")

    result := []Contact{}
    err = c.Find(bson.M{param: value}).All(&result)
    if err != nil {
         if Debug {
            InfoLog.Printf("Cannot retrieve contact where %s = %s \n", param, value)
        }
        return []Contact{}
    }

    return result
}

func (db *MongoDatastore) DeleteContact(id string) bool {
    session, err := mgo.Dial(mongoUrl)
    check(err)
    defer session.Close()

    // Optional. Switch the session to a monotonic behavior.
    // session.SetMode(mgo.Monotonic, true)

    objectId := bson.ObjectIdHex(id)
    c := session.DB(dbName).C("contact")
    err = c.RemoveId(objectId)
    if err != nil {
         return false
    }

    return true
}

func (db *MongoDatastore) Status() bool {
    return true
}
