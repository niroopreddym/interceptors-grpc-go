package store

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/jinzhu/copier"
	"github.com/niroopreddym/interceptors-grpc-go/pb"
)

//ErrAlreadyExists returns if the laptop with same id already exists in the store
var ErrAlreadyExists = errors.New("error already exists")

//LaptopStore proides an interface to save the laptop data
type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	Find(id string) (*pb.Laptop, error)
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

//InMemoryLaptopStore proides an interface to save the laptop data to in memory store
type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

//NewInMemoryLaptopStore returns a new InMemoryLaptopStore
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

//Save saves the laptop to the store
func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	//deep copy
	other, err := deepCopy(laptop)
	if err != nil {
		return err
	}

	store.data[other.Id] = other
	return nil
}

//Find finds a laptop by ID
func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}

	//deep copy
	return deepCopy(laptop)
}

//Search searches the in memory data store
func (store *InMemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, laptop := range store.data {
		// time.Sleep(1 * time.Second)
		log.Print("checking laptop id: ", laptop.GetId())

		//check the context
		if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
			log.Print("Context is cancelled")
			return errors.New("context is cancelled")
		}

		if isQualified(filter, laptop) {
			// deep copy
			other, err := deepCopy(laptop)
			if err != nil {
				return err
			}

			err = found(other)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}

	if laptop.GetCpu().GetNumberOfCores() < uint32(filter.GetMinCpuCores()) {
		return false
	}

	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}

	if toBit(laptop.GetRam()) < toBit(filter.GetMinRam()) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	value := memory.GetValue()
	switch memory.GetUnit() {
	case pb.Unit_BIT:
		return value
	case pb.Unit_BYTE:
		return value << 3 // 8 = 2 ^ 3
	case pb.Unit_KILOBYTE:
		return value << 13 // 1024 x 8 = 2 ^ 13
	case pb.Unit_MEGABYTE:
		return value << 23 // 1024 x 1024 x 8 = 2 ^ 23
	case pb.Unit_GIGABTE:
		return value << 33
	case pb.Unit_TERABYTE:
		return value << 43
	default:
		return 0
	}
}

func deepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy latop data : %w", err)
	}

	return other, nil
}
