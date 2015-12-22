package main

import "./cqrs"
import "./uid"
import "errors"
import "fmt"

type Thing struct {
	cqrs.AggregateRoot
	Id string
}

func (t *Thing) HandleCreateEvent(e CreateEvent) {
	t.Id = e.Id
}

type ThingRepository struct {
	things []Thing
	bus    *cqrs.Bus
}

func NewThing(name string) Thing {
	id := uid.GetId()
	thing := Thing{}
	thing.AggregateRoot = cqrs.NewAggregateRoot(&thing)
	thing.Source(CreateEvent{id, name})
	return thing
}

func NewThingRepository(bus *cqrs.Bus) ThingRepository {
	return ThingRepository{make([]Thing, 0), bus}
}

func (t *ThingRepository) FindById(id string) (Thing, error) {
	for _, thing := range t.things {
		if thing.Id == id {
			return thing, nil
		}
	}
	return Thing{}, errors.New("Thing not found")
}

func (t *ThingRepository) Save(id string) {
	for _, thing := range t.things {
		if thing.Id == id {
			for _, change := range thing.Changes {
				// fmt.Println("PUBLI", change.Payload)
				t.bus.Publish(change.Payload)
			}
			thing.ClearChanges()
		}
	}
}

func (t *ThingRepository) Add(thing Thing) {
	t.things = append(t.things, thing)
}

func (t *Thing) ChangeName(NewName string) {
	t.Source(ChangeNameEvent{t.Id, NewName})
}

type ThingDetail struct {
	Id   string
	Name string
}

// Read model
type ThingList struct {
	things map[string]ThingDetail
}

func (t *ThingList) GetThing(id string) ThingDetail {
	return t.things[id]
}

func (t *ThingList) HandleChangeNameEvent(e ChangeNameEvent) {
	// t.things[e.Id].Name = e.NewName
	thing := t.things[e.Id]
	thing.Name = e.NewName
	t.things[e.Id] = thing
}

func (t *ThingList) HandleCreateEvent(e CreateEvent) {
	thing := ThingDetail{}
	thing.Id = e.Id
	thing.Name = e.Name
	t.things[e.Id] = thing
}

func NewThingList() ThingList {
	return ThingList{make(map[string]ThingDetail)}
}

func main() {
	bus := cqrs.NewBus()
	things := NewThingList()
	repo := NewThingRepository(&bus)
	bus.RegisterHandlers(&things)

	thing := NewThing("hola")
	repo.Add(thing)
	repo.Save(thing.Id)

	fmt.Println(things.GetThing(thing.Id))
	thing.ChangeName("adeu")
	repo.Save(thing.Id)

	fmt.Println(things.GetThing(thing.Id))
}
