package sample

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/niroopreddym/interceptors-grpc-go/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//Newkeyboard creates new kyboard object
func Newkeyboard() *pb.Keyboard {
	keyboard := &pb.Keyboard{
		Layout:  randomKeyboardLayout(),
		Backlit: randomBool(),
	}

	return keyboard
}

//NewCPU creates a new CPU
func NewCPU() *pb.CPU {
	brand := randomCPUBrand()
	name := randomCPUName(brand)
	numberOfCores := randomInt(2, 8)
	numberOfThreads := randomInt(numberOfCores, 12)
	minGhz := randomFloat(2.0, 3.5)
	maxGhz := randomFloat(minGhz, 5.0)

	cpu := &pb.CPU{
		Brand:           brand,
		Name:            name,
		NumberOfCores:   uint32(numberOfCores),
		NumberOfThreads: uint32(numberOfThreads),
		MinGhz:          minGhz,
		MaxGhz:          maxGhz,
	}

	return cpu
}

//NewGPU returns new GPU
func NewGPU() *pb.GPU {
	brand := randomGPUBrand()
	name := randomGPUName(brand)
	minGhz := randomFloat(1.0, 1.5)
	maxGhz := randomFloat(minGhz, 3.0)

	memory := pb.Memory{
		Value: uint64(randomInt(2, 6)),
		Unit:  pb.Unit_MEGABYTE,
	}

	gpu := &pb.GPU{
		Brand:  brand,
		Name:   name,
		MinGhz: minGhz,
		MaxGhz: maxGhz,
		Memory: &memory,
	}

	return gpu
}

//NewRAM returns new RAM
func NewRAM() *pb.Memory {
	RAM := pb.Memory{
		Value: uint64(randomInt(2, 6)),
		Unit:  pb.Unit_GIGABTE,
	}

	return &RAM
}

//NewSSD creates new SSD
func NewSSD() *pb.Storage {
	ssd := &pb.Storage{
		Driver: pb.Storage_SSD,
		Memory: &pb.Memory{
			Value: uint64(randomInt(128, 1024)),
			Unit:  pb.Unit_GIGABTE,
		},
	}

	return ssd
}

//NewHDD creates new HDD
func NewHDD() *pb.Storage {
	ssd := &pb.Storage{
		Driver: pb.Storage_HDD,
		Memory: &pb.Memory{
			Value: uint64(randomInt(1, 6)),
			Unit:  pb.Unit_TERABYTE,
		},
	}

	return ssd
}

//NewScreen returns new screen size
func NewScreen() *pb.Screen {
	screen := &pb.Screen{
		SizeInch:   randomFloat32(13, 17),
		Resolution: randomScreenResolution(),
		Panel:      randomScreenPanel(),
	}

	return screen
}

//NewLaptop creates a new laptop
func NewLaptop() *pb.Laptop {
	brand := randomLaptopBrand()
	name := randomlaptopName(brand)

	laptop := &pb.Laptop{
		Id:    randomUUID(),
		Brand: brand,
		Name:  name,
		Cpu:   NewCPU(),
		Ram:   NewRAM(),
		Gpus: []*pb.GPU{
			NewGPU(),
		},
		Storages: []*pb.Storage{
			NewSSD(),
			NewHDD(),
		},
		Screen:   NewScreen(),
		Keyboard: Newkeyboard(),
		Weight: &pb.Laptop_WeightKg{
			WeightKg: randomFloat(1.0, 3.0),
		},
		PriceUsd:    randomFloat(1500, 3000),
		ReleaseYear: uint32(randomInt(2015, 2021)),
		UpdatedAt:   (*timestamppb.Timestamp)(ptypes.TimestampNow()),
	}

	return laptop
}

//RandomLaptopScore returns a random score
func RandomLaptopScore() float64 {
	return float64(randomInt(1, 10))
}
