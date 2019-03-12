package bulletinboard

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/bl4ck5un/ChuRP/src/services"
	"github.com/bl4ck5un/ChuRP/src/utils/commitment"
	"github.com/bl4ck5un/ChuRP/src/utils/polyring"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	// "time"
	"github.com/golang/protobuf/proto"
	"github.com/ncw/gmp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// BulletinBoard Simulator Structure
type BulletinBoard struct {
	// Metadata Directory Path
	metadataPath string
	// Counter
	counter int
	// BulletinBoard IP Address
	bip string
	// IP
	ipList []string
	// Rand
	randState *rand.Rand
	// Reconstruction BulletinBoard
	reconstructionContent []*pb.Cmt1Msg
	// Proactivization BulletinBoard
	proCnt                 *int
	proactivizationContent []*pb.Cmt2Msg
	// Share Distribution BulletinBoard
	shaCnt *int

	// Mutexes
	mutex sync.Mutex

	nConn   []*grpc.ClientConn
	nClient []pb.NodeServiceClient

	// Metrics
	totMsgSize *int
}

func (bb *BulletinBoard) StartEpoch(ctx context.Context, in *pb.EmptyMsg) (*pb.AckMsg, error) {
	log.Print("[bulletinboard] start epoch")
	bb.ClientStartPhase1()
	return &pb.AckMsg{}, nil
}

func (bb *BulletinBoard) ReadPhase1(in *pb.EmptyMsg, stream pb.BulletinBoardService_ReadPhase1Server) error {
	log.Print("[bulletinboard] is being read in phase 1")
	for i := 0; i < bb.counter; i++ {
		if err := stream.Send(bb.reconstructionContent[i]); err != nil {
			log.Fatalf("bulletinboard failed to read phase1: %v", err)
			return err
		}
	}
	return nil
}

func (bb *BulletinBoard) WritePhase2(ctx context.Context, msg *pb.Cmt2Msg) (*pb.AckMsg, error) {
	*bb.totMsgSize = *bb.totMsgSize + proto.Size(msg)
	log.Print("[bulletinboard] is being written in phase 2")
	index := msg.GetIndex()
	bb.proactivizationContent[index-1] = msg
	bb.mutex.Lock()
	*bb.proCnt = *bb.proCnt + 1
	flag := (*bb.proCnt == bb.counter)
	bb.mutex.Unlock()
	if flag {
		*bb.proCnt = 0
		bb.ClientStartVerifPhase2()
	}
	return &pb.AckMsg{}, nil
}

func (bb *BulletinBoard) ReadPhase2(in *pb.EmptyMsg, stream pb.BulletinBoardService_ReadPhase2Server) error {
	log.Print("[bulletinboard] is beting read in phase 2")
	for i := 0; i < bb.counter; i++ {
		if err := stream.Send(bb.proactivizationContent[i]); err != nil {
			log.Fatalf("bulletinboard failed to read phase2: %v", err)
			return err
		}
	}
	return nil
}

func (bb *BulletinBoard) WritePhase3(ctx context.Context, msg *pb.Cmt1Msg) (*pb.AckMsg, error) {
	*bb.totMsgSize = *bb.totMsgSize + proto.Size(msg)
	log.Print("[bulletinboard] is being written in phase 3")
	index := msg.GetIndex()
	bb.reconstructionContent[index-1] = msg
	bb.mutex.Lock()
	*bb.shaCnt = *bb.shaCnt + 1
	flag := (*bb.shaCnt == bb.counter)
	bb.mutex.Unlock()
	if flag {
		*bb.shaCnt = 0
		bb.ClientStartVerifPhase3()
	}
	return &pb.AckMsg{}, nil
}

func (bb *BulletinBoard) ReadPhase3(in *pb.EmptyMsg, stream pb.BulletinBoardService_ReadPhase3Server) error {
	log.Print("[bulletinboard] is being read in phase 3")
	for i := 0; i < bb.counter; i++ {
		if err := stream.Send(bb.reconstructionContent[i]); err != nil {
			log.Fatalf("bulletinboard failed to read phase2: %v", err)
			return err
		}
	}
	return nil
}

func (bb *BulletinBoard) Connect() {
	for i := 0; i < bb.counter; i++ {
		nConn, err := grpc.Dial(bb.ipList[i], grpc.WithInsecure())
		if err != nil {
			log.Fatalf("bulletinboard did not connect: %v", err)
		}
		bb.nConn[i] = nConn
		bb.nClient[i] = pb.NewNodeServiceClient(nConn)
	}
}

func (bb *BulletinBoard) Disconnect() {
	for i := 0; i < bb.counter; i++ {
		bb.nConn[i].Close()
	}
}

func (bb *BulletinBoard) Serve(aws bool) {
	port := bb.bip
	if aws {
		port = "0.0.0.0:12001"
	}
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("bulletinboard failed to listen %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterBulletinBoardServiceServer(s, bb)
	reflection.Register(s)
	log.Printf("bulletinboard serve on " + bb.bip)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("bulletinboard failed to serve %v", err)
	}
}

func (bb *BulletinBoard) ClientStartPhase1() {
	if bb.nConn[0] == nil {
		bb.Connect()
	}
	var wg sync.WaitGroup
	for i := 0; i < bb.counter; i++ {
		log.Print("[bulletinboard] start phase 1")
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			bb.nClient[i].StartPhase1(ctx, &pb.EmptyMsg{})
		}(i)
	}
	wg.Wait()
}

func (bb *BulletinBoard) ClientStartVerifPhase2() {
	var wg sync.WaitGroup
	for i := 0; i < bb.counter; i++ {
		log.Print("[bulletinboard] start verification in phase 2")
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			bb.nClient[i].StartVerifPhase2(ctx, &pb.EmptyMsg{})
		}(i)
	}
	wg.Wait()
}

func (bb *BulletinBoard) ClientStartVerifPhase3() {
	var wg sync.WaitGroup
	for i := 0; i < bb.counter; i++ {
		log.Print("[bulletinboard] start verification in phase 3")
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			bb.nClient[i].StartVerifPhase3(ctx, &pb.EmptyMsg{})
		}(i)
	}
	wg.Wait()
	*bb.proCnt = 0
	*bb.shaCnt = 0
	f, _ := os.OpenFile(bb.metadataPath+"/log0", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	fmt.Fprintf(f, "totMsgSize,%d\n", *bb.totMsgSize)
	*bb.totMsgSize = 0
}

func ReadIpList(metadataPath string) []string {
	ipData, err := ioutil.ReadFile(metadataPath + "/ip_list")
	if err != nil {
		log.Fatalf("bulletinboard failed to read iplist: %v", err)
	}
	return strings.Split(string(ipData), "\n")
}

// New returns a network node structure
func New(degree int, counter int, metadataPath string) (BulletinBoard, error) {
	f, _ := os.Create(metadataPath + "/log0")
	defer f.Close()
	if counter < 0 {
		return BulletinBoard{}, errors.New(fmt.Sprintf("counter must be non-negative, got %d", counter))
	}

	fixedRandState := rand.New(rand.NewSource(int64(3)))
	p := gmp.NewInt(0)
	p.SetString("57896044618658097711785492504343953926634992332820282019728792006155588075521", 10)
	dpc := commitment.DLPolyCommit{}
	dpc.SetupFix(counter)

	ipRaw := ReadIpList(metadataPath)[0 : counter+1]
	bip := ipRaw[0]
	ipList := ipRaw[1 : counter+1]

	proCnt := 0
	shaCnt := 0

	reconstructionContent := make([]*pb.Cmt1Msg, counter)
	poly, err := polyring.NewRand(degree, fixedRandState, p)
	if err != nil {
		log.Fatal("Error initializing random poly")
	}
	c := dpc.NewG1()
	dpc.Commit(c, poly)
	cBytes := c.CompressedBytes()
	for i := 0; i < counter; i++ {
		msg := &pb.Cmt1Msg{
			Index:   int32(i + 1),
			Polycmt: cBytes,
		}
		reconstructionContent[i] = msg
	}
	proactivizationContent := make([]*pb.Cmt2Msg, counter)

	nConn := make([]*grpc.ClientConn, counter)
	nClient := make([]pb.NodeServiceClient, counter)

	totMsgSize := 0

	return BulletinBoard{
		metadataPath:           metadataPath,
		counter:                counter,
		bip:                    bip,
		ipList:                 ipList,
		proCnt:                 &proCnt,
		shaCnt:                 &shaCnt,
		reconstructionContent:  reconstructionContent,
		proactivizationContent: proactivizationContent,
		nConn:                  nConn,
		nClient:                nClient,
		totMsgSize:             &totMsgSize,
	}, nil
}
