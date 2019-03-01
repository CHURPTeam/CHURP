package clock

import (
	// "os"
	// "fmt"
	"log"
	// "net"
	// "errors"
	"context"
	"strings"
	"io/ioutil"
	"google.golang.org/grpc"
	pb "../../services"
)

// Clock Simulator Structure
type Clock struct {
	// Metadata Directory Path
	metadataPath string
	// BulltinBoard IP
	bip string
	// BulletinBoard Service Client
	bConn *grpc.ClientConn
	bClient pb.BulletinBoardServiceClient
}

func (clock *Clock) Connect() {
	bConn, err := grpc.Dial(clock.bip, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("clock did not connect: %v", err)
	}
	clock.bConn = bConn
	clock.bClient = pb.NewBulletinBoardServiceClient(clock.bConn)
}

func (clock *Clock) Disconnect() {
	clock.bConn.Close()
}

func (clock *Clock) ClientStartEpoch() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
    log.Print("client start epoch")
	_, err := clock.bClient.StartEpoch(ctx, &pb.EmptyMsg{})
	if err != nil {
		log.Fatalf("clock start epoch failed: %v", err)
	}
}

func ReadIpList(metadataPath string) []string {
	ipData, err := ioutil.ReadFile(metadataPath + "/ip_list")
	if err != nil {
		log.Fatalf("clock failed to read iplist %v\n", err)
	}
	return strings.Split(string(ipData), "\n")
}

// New returns a network node structure
func New(metadataPath string) (Clock, error) {
	bip := ReadIpList(metadataPath)[0]
	return Clock{
		metadataPath:           metadataPath,
		bip:                    bip,
	}, nil
}
