package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/niroopreddym/interceptors-grpc-go/pb"
	"github.com/niroopreddym/interceptors-grpc-go/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//LaptopServer implements the laptop server proto server interface
type LaptopServer struct {
	Store       store.LaptopStore
	ImageStore  *store.DiskImageStore
	RatingStore store.RatingStore
}

//NewLaptopServer provides the constructor for laptop server
func NewLaptopServer(store store.LaptopStore, imageStore *store.DiskImageStore, ratingStore store.RatingStore) *LaptopServer {
	return &LaptopServer{
		Store:       store,
		ImageStore:  imageStore,
		RatingStore: ratingStore,
	}
}

//CreateLaptop is the unary rpc implemetation to create a new laptop
func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Println("recieved a create-laptop request with id: ", laptop.Id)

	if len(laptop.Id) > 0 {
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop ID is not a valid UUID: "+err.Error())
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot generate a new laptop ID: "+err.Error())
		}

		laptop.Id = id.String()
	}

	//mock some heavy processing before finishing off the service request
	// time.Sleep(6 * time.Second)

	if ctx.Err() == context.Canceled {
		log.Println("ctx cancelled")
		return nil, status.Error(codes.Canceled, "ctx cancelled")
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Println("deadline exceeded")
		return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
	}

	err := server.Store.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, store.ErrAlreadyExists) {
			code = codes.AlreadyExists
		}

		return nil, status.Error(code, "cannot save laptop to the store: "+err.Error())
	}

	log.Printf("laptop save diwth id: %s", laptop.Id)
	res := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}

	return res, nil
}

//SearchLaptop searches and returns a single laptop from the datastore
func (server *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopService_SearchLaptopServer) error {
	filter := req.GetFilter()
	log.Printf("recieve a search-laptop request with filter: %v", filter)
	err := server.Store.Search(stream.Context(), filter, func(laptop *pb.Laptop) error {
		response := &pb.SearchLaptopResponse{
			Laptop: laptop,
		}

		err := stream.Send(response)
		if err != nil {
			return err
		}

		log.Printf("sent laptop with id : %s", laptop.GetId())
		return nil
	})

	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	log.Printf("Done Server Side code")
	return nil
}

const maxImageSize = 1 << 20

//UploadImage uploads the image to in memory datastore
func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot recieve message Info"))
	}

	laptopID := req.GetInfo().GetLaptopId()
	ImageType := req.GetInfo().GetImageType()
	log.Printf("recived upload image request for laptop %s with image type %s", laptopID, ImageType)

	laptop, err := server.Store.Find(laptopID)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot find laptop: %v", err))
	}

	if laptop == nil {
		return logError(status.Errorf(codes.InvalidArgument, "laptop: %s does not exist", laptopID))
	}

	imageData := bytes.Buffer{}
	imageSizeBuffered := 0

	log.Print("entering the read chunk loop")

	for {
		log.Print("waiting to recieve chunk data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}

		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot rcieve chunk data laptop: %v", err))
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		imageSizeBuffered += size

		log.Printf("receuved chunk data with size: %d", size)

		if imageSizeBuffered > maxImageSize {
			return logError(status.Errorf(codes.InvalidArgument, "imageSize too large: %d > %d", imageSizeBuffered, maxImageSize))
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
		}
	}

	imageID, err := server.ImageStore.Save(laptopID, ImageType, imageData)

	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot save image to the store: %v", err))
	}

	res := &pb.UploadImageResponse{
		Id:   imageID,
		Size: uint32(imageSizeBuffered),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}

	log.Printf("saved image with id: %s, size: %d", imageID, imageSizeBuffered)
	return nil
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}

	return err
}

//RateLaptop gets the stream of ratings
func (server *LaptopServer) RateLaptop(stream pb.LaptopService_RateLaptopServer) error {
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Printf("no more data")
			break
		}

		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot recueve stream request: %v", err))
		}

		laptopID := req.GetLaptopId()
		score := req.GetScore()

		log.Printf("recieved a rate-laptop request: id=%s, score=%.2f", laptopID, score)

		found, err := server.Store.Find(laptopID)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot find laptopt: %v", err))
		}

		if found == nil {
			return logError(status.Errorf(codes.NotFound, "laptop wit ID : %s not found", laptopID))
		}

		rating, err := server.RatingStore.Add(laptopID, score)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot add rating to the store: %v", err))
		}

		res := &pb.RateLaptopResponse{
			LaptopId:     laptopID,
			RatedCount:   rating.Count,
			AverageScore: rating.Sum / float64(rating.Count),
		}

		err = stream.Send(res)
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot send stream response: %v", err))
		}
	}

	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is canceled"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}
