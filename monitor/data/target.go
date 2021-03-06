package data

import (
	"log"

	"github.com/maxcnunes/go-uptime-api/monitor/entities"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// DataTarget is the data configuration related to Target collection
type DataTarget struct {
	collection *mgo.Collection
	events     chan entities.Event
}

// FindOneByURL ...
func (d *DataTarget) FindOneByURL(url string) *entities.Target {
	var target entities.Target
	err := d.collection.Find(bson.M{"url": url}).One(&target)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil
		}

		log.Printf("got an error finding a doc %v\n", err)
	}

	return &target
}

// FindOneByID finds a single target by the id field
func (d *DataTarget) FindOneByID(id string) *entities.Target {
	_id := bson.ObjectIdHex(id)
	var target entities.Target

	err := d.collection.Find(bson.M{"_id": _id}).One(&target)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil
		}

		log.Printf("got an error finding a doc %v\n", err)
	}

	return &target
}

// Create a new target in the database
func (d *DataTarget) Create(url string, emails []string) *entities.Target {
	target := d.FindOneByURL(url)
	if target != nil {
		return target
	}

	log.Printf("Adding target %s", url)

	doc := entities.Target{ID: bson.NewObjectId(), URL: url, Status: 0, Emails: emails}
	if err := d.collection.Insert(doc); err != nil {
		log.Printf("Can't insert document: %v\n", err)
	}

	go func() { d.events <- entities.Event{Event: entities.Added, Target: &doc} }()

	return &doc
}

// Remove an existing target by the URL field
func (d *DataTarget) Remove(url string) {
	target := d.FindOneByURL(url)
	if target == nil {
		log.Printf("Can't find document with url: %s\n", url)
		return
	}

	log.Printf("Removing url %s", url)

	err := d.collection.Remove(bson.M{"url": url})
	if err != nil {
		log.Printf("Can't delete document: %v\n", err)
	}

	go func() { d.events <- entities.Event{Event: entities.Removed, Target: target} }()
}

// RemoveByID removes an existing target by the ID field
func (d *DataTarget) RemoveByID(id string) {
	target := d.FindOneByID(id)
	if target == nil {
		log.Printf("Can't find document with id: %s\n", id)
		return
	}

	log.Printf("Removing url %s", target.URL)

	err := d.collection.Remove(bson.M{"_id": target.ID})
	if err != nil {
		log.Printf("Can't delete document: %v\n", err)
	}

	go func() { d.events <- entities.Event{Event: entities.Removed, Target: target} }()
}

// UpdateStatusByURL updates a existing target by the URL field
func (d *DataTarget) UpdateStatusByURL(url string, status string) {
	target := d.FindOneByURL(url)
	if target != nil {
		log.Printf("Can't find document with url: %s\n", url)
		return
	}

	log.Printf("Updating url %s to status %s", url, status)
	err := d.collection.Update(bson.M{"url": url}, bson.M{"status": status})
	if err != nil {
		log.Printf("Can't update document: %v\n", err)
	}

	go func() { d.events <- entities.Event{Event: entities.Updated, Target: target} }()
}

// Update an existing target
func (d *DataTarget) Update(id string, data entities.Target) {
	target := d.FindOneByID(id)
	if target == nil {
		log.Printf("Can't find document with id: %s\n", id)
		return
	}

	attrs := bson.M{"url": data.URL, "emails": data.Emails}
	if data.Status != 0 {
		attrs["status"] = data.Status
	}

	log.Printf("Updating url %s to status %d", target.URL, target.Status)
	err := d.collection.UpdateId(target.ID, bson.M{"$set": attrs})
	if err != nil {
		log.Printf("Can't update document: %v\n", err)
	}

	go func() { d.events <- entities.Event{Event: entities.Updated, Target: target} }()
}

// GetAllURLS returns all existing target's URLs in the databse
func (d *DataTarget) GetAllURLS() []string {
	urls := []string{}

	targets := d.GetAll()

	for _, target := range targets {
		urls = append(urls, target.URL)
	}

	return urls
}

// GetAll returns all targets
func (d *DataTarget) GetAll() []entities.Target {
	targets := []entities.Target{}

	err := d.collection.Find(nil).All(&targets)
	if err != nil {
		log.Printf("got an error finding a doc %v\n", err)
	}

	return targets
}

// Start a new instance of data target
func (d *DataTarget) Start(db DB, events chan entities.Event) {
	d.collection = db.Session.DB(db.DBName).C("target")
	d.events = events
}
