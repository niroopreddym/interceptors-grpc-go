package service_test

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/niroopreddym/interceptors-grpc-go/pb"
	"github.com/niroopreddym/interceptors-grpc-go/sample"
	"github.com/niroopreddym/interceptors-grpc-go/serializer"
	"github.com/niroopreddym/interceptors-grpc-go/service"
	"github.com/niroopreddym/interceptors-grpc-go/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestClientCreatelaptop(t *testing.T) {
	t.Parallel()
	laptopStore := store.NewInMemoryLaptopStore()
	serverAddress := startTestLaptopServer(t, laptopStore, nil, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	laptop := sample.NewLaptop()
	expectedID := laptop.Id
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	res, err := laptopClient.CreateLaptop(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, expectedID, res.Id)

	//check the laptop saved to the store
	other, err := laptopStore.Find(laptop.Id)
	assert.NoError(t, err)
	assert.NotNil(t, other)

	//check the laptop is same as send laptop
	assertSameLaptop(t, laptop, other)

}

func TestClientSearchlaptop(t *testing.T) {
	t.Parallel()

	filter := &pb.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz:   2.2,
		MinRam: &pb.Memory{
			Value: 8,
			Unit:  pb.Unit_GIGABTE,
		},
	}
	store := store.NewInMemoryLaptopStore()
	expectedIDs := make(map[string]bool)

	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()
		switch i {
		case 0:
			laptop.PriceUsd = 2500
		case 1:
			laptop.Cpu.NumberOfCores = 2
		case 2:
			laptop.Cpu.MinGhz = 2.0
		case 3:
			laptop.Ram = &pb.Memory{Value: 4, Unit: pb.Unit_GIGABTE}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberOfCores = 6
			laptop.Cpu.MinGhz = 2.5
			laptop.Ram = &pb.Memory{
				Value: 12,
				Unit:  pb.Unit_GIGABTE,
			}
			expectedIDs[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 2000
			laptop.Cpu.NumberOfCores = 12
			laptop.Cpu.MinGhz = 2.5
			laptop.Ram = &pb.Memory{
				Value: 64,
				Unit:  pb.Unit_GIGABTE,
			}
			expectedIDs[laptop.Id] = true
		}

		err := store.Save(laptop)
		assert.Nil(t, err)
	}

	serverAddress := startTestLaptopServer(t, store, nil, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}

	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	assert.NoError(t, err)
	found := 0

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		assert.NoError(t, err)
		assert.Contains(t, expectedIDs, res.GetLaptop().GetId())
		found += 1
	}

	assert.Equal(t, len(expectedIDs), found)
}

func TestClientRateLaptop(t *testing.T) {
	t.Parallel()

	laptopStore := store.NewInMemoryLaptopStore()
	ratingStore := store.NewInMemoryRatingStore()

	laptop := sample.NewLaptop()
	err := laptopStore.Save(laptop)
	assert.NoError(t, err)

	serverAddress := startTestLaptopServer(t, laptopStore, nil, ratingStore)
	laptopClient := newTestLaptopClient(t, serverAddress)

	stream, err := laptopClient.RateLaptop(context.Background())
	require.NoError(t, err)

	scores := []float64{8, 7.5, 10}
	averages := []float64{8, 7.75, 8.5}

	n := len(scores)
	for i := 0; i < n; i++ {
		req := &pb.RateLaptopRequest{
			LaptopId: laptop.GetId(),
			Score:    scores[i],
		}

		err := stream.Send(req)
		require.NoError(t, err)
	}

	err = stream.CloseSend()
	require.NoError(t, err)

	for idx := 0; ; idx++ {
		res, err := stream.Recv()
		if err == io.EOF {
			require.Equal(t, n, idx)
			return
		}

		require.NoError(t, err)
		require.Equal(t, laptop.GetId(), res.GetLaptopId())
		require.Equal(t, uint32(idx+1), res.GetRatedCount())
		require.Equal(t, averages[idx], res.GetAverageScore())
	}
}

func startTestLaptopServer(t *testing.T, store store.LaptopStore, imageStore *store.DiskImageStore, ratingStore store.RatingStore) string {
	laptopServer := service.NewLaptopServer(store, imageStore, ratingStore)

	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	listener, err := net.Listen("tcp", ":0") //radnom port :0
	assert.NoError(t, err)

	go grpcServer.Serve(listener) //non blocking code

	return listener.Addr().String()
}

func newTestLaptopClient(t *testing.T, serverAddress string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	assert.NoError(t, err)
	return pb.NewLaptopServiceClient(conn)
}

func assertSameLaptop(t *testing.T, laptop1 *pb.Laptop, laptop2 *pb.Laptop) {
	json1, err := serializer.ProtobufToJSON(laptop1)
	assert.NoError(t, err)

	json2, err := serializer.ProtobufToJSON(laptop2)
	assert.NoError(t, err)

	assert.Equal(t, json1, json2)
}
