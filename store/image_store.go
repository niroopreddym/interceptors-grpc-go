package store

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

//ImageStore is an interface to store laptop images
type ImageStore interface {
	Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error)
}

//DiskImageStore is a struct to store image data
type DiskImageStore struct {
	mutex       sync.RWMutex
	imageFolder string
	images      map[string]*ImageInfo
}

//ImageInfo is a struct
type ImageInfo struct {
	LaptopID string
	Type     string
	Path     string
}

//NewDiskImageStore is the constructor for DiskImageStore
func NewDiskImageStore(imageFolder string) DiskImageStore {
	return DiskImageStore{
		imageFolder: imageFolder,
		images:      map[string]*ImageInfo{},
	}
}

//Save saves teh image to the idsk location
func (store *DiskImageStore) Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate image id : %w", err)
	}

	// imagePath := fmt.Sprintf("image path %s/%s%s", store.imageFolder, imageID, imageType)
	imagePath := filepath.Join(store.imageFolder, imageID.String()+imageType)
	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: %w", err)
	}

	log.Print("filepath is: ", imagePath)
	_, err = imageData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write image to file: %w", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.images[imageID.String()] = &ImageInfo{
		LaptopID: laptopID,
		Type:     imageType,
		Path:     imagePath,
	}

	return imageID.String(), nil
}
