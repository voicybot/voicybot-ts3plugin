package resolvers

import (
	"fmt"
	"sync"
)

var registered []Resolver = []Resolver{}
var mutex = sync.Mutex{}

func Register(instanceToRegister Resolver) {
	mutex.Lock()
	defer mutex.Unlock()

	// Is this instance already registered?
	for _, instance := range registered {
		if instance == instanceToRegister ||
			instance.Id() == instanceToRegister.Id() {
			panic(fmt.Errorf("This plugin was already registered: %s", instance.Id()))
		}
	}

	registered = append(registered, instanceToRegister)
}

func ById(id string) (instance Resolver) {
	mutex.Lock()
	defer mutex.Unlock()

	for _, foundInstance := range registered {
		if foundInstance.Id() == id {
			instance = foundInstance
			return
		}
	}
	return
}

func GetAll() (resolvers []Resolver) {
	mutex.Lock()
	defer mutex.Unlock()

	resolvers = make([]Resolver, len(registered))
	copy(resolvers, registered)

	return
}

func Unregister(instanceToUnregister Resolver) {
	mutex.Lock()
	defer mutex.Unlock()

	for index, instance := range registered {
		if instance == instanceToUnregister {
			registered = append(registered[0:index], registered[index+1:]...)
			return
		}
	}
}
