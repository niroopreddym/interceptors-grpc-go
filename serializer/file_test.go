package serializer

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/niroopreddym/interceptors-grpc-go/pb"
	"github.com/niroopreddym/interceptors-grpc-go/sample"
	"github.com/stretchr/testify/assert"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()

	binaryFile := "C:/Users/maneti.n/go/src/github.com/niroopreddym/interceptors-grpc-go/tmp/laptop.bin"
	jsonFile := "C:/Users/maneti.n/go/src/github.com/niroopreddym/interceptors-grpc-go/tmp/laptop.json"
	laptop1 := sample.NewLaptop()
	err := WriteProtobufToBinaryFile(laptop1, binaryFile)
	assert.Nil(t, err)

	laptop2 := &pb.Laptop{}
	err = ReadProtobufFromBinaryFile(binaryFile, laptop2)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(laptop1, laptop2))

	err = WriteProtobufToJSONFile(laptop1, jsonFile)
	assert.Nil(t, err)
}
