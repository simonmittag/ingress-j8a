package server

import (
	"crypto/sha1"
	"fmt"
	"k8s.io/apimachinery/pkg/util/json"
	"sync"
	"time"
)

type Cache struct {
	Mementos []Memento
	lock     sync.Mutex
}

func NewCache() *Cache {
	return &Cache{
		Mementos: make([]Memento, 0),
		lock:     sync.Mutex{},
	}
}

func (c *Cache) update(data interface{}) {
	c.lock.Lock()

	var m *Memento
	if l := len(c.Mementos); l > 0 {
		m = c.Mementos[len(c.Mementos)-1].Clone()
	} else {
		m = NewMemento()
	}

	switch data.(type) {
	case []Route:
		m.Routes = data.([]Route)
		m.SetHash()
	}

	//TODO: does not do []Route currently individual route.
	c.Mementos = append(c.Mementos, *m)

	c.lock.Unlock()
}

type Memento struct {
	Routes      []Route
	Hash        string
	DateCreated time.Time
}

func NewMemento() *Memento {
	m := &Memento{
		Routes:      make([]Route, 0),
		Hash:        "",
		DateCreated: time.Now(),
	}
	m.SetHash()
	return m
}

func (m *Memento) SetHash() {
	data, _ := json.Marshal(m.Routes)
	m.Hash = fmt.Sprintf("%x", sha1.Sum(data))
}

func (m *Memento) Clone() *Memento {
	d := Memento{
		Routes:      m.Routes,
		DateCreated: time.Now(),
	}
	d.SetHash()
	return &d
}
