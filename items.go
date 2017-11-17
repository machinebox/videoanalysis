package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Items struct {
	sync.Mutex
	data map[string]Item
}

func (items *Items) Add(id string, it Item) {
	items.Lock()
	defer items.Unlock()
	items.data[id] = it
}

func (items *Items) List() []Item {
	items.Lock()
	defer items.Unlock()
	all := make([]Item, len(items.data))
	i := 0
	for k := range items.data {
		all[i] = items.data[k]
		i++
	}
	return all
}

func LoadItemsFromPath(dir string) *Items {
	items := Items{data: map[string]Item{}}
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		ll := strings.ToLower(info.Name())
		if !strings.HasSuffix(ll, ".mov") && !strings.HasSuffix(ll, ".mpg") {
			return nil
		}
		items.Add(info.Name(), Item{Name: info.Name(), Path: path})
		log.Printf("+ Add Video: %v -> %v \n", info.Name(), path)
		return nil
	})
	return &items
}

type Item struct {
	Path string `json:"path,omitempty"`
	Name string `json:"name,omitempty"`
}
