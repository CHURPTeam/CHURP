package nodes

import (
	pb "../../services"
	"../../utils/commitment"
	"../../utils/interpolation"
	"../../utils/polypoint"
	"../../utils/polyring"
	"context"
	"errors"
	"fmt"
	"github.com/Nik-U/pbc"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"sync"
	"github.com/ncw/gmp"
	"google.golang.org/grpc/reflection"
	"github.com/golang/protobuf/proto"
)

// Network Node Structure
type Node struct {
	// Metadata Path
	metadataPath string
	// Basic Sharing Information
	// [+] Label of Node
	label int
	// [+] Number of Nodes
	counter int
	// [+] Polynomial Degree
	degree int
	// [+] Prime Defining Group Z_p
	p *gmp.Int

	// IP Information
	// [+] Bulletinboard IP Address
	bip string
	// [+] Node IP Address List
	ipList []string

	// Utilities
	// [+] Rand Source
	randState *rand.Rand
	// [+] Commitment
	dc  *commitment.DLCommit
	dpc *commitment.DLPolyCommit

	// Sharing State
	// [+] Polynomial State
	secretShares []*polypoint.PolyPoint

	// Reconstruction Phase
	// [+] Polynomial Reconstruction State
	recShares []*polypoint.PolyPoint
	// [+] Counter of Polynomial Reconstruction State
	recCnt *int
	// [+] Mutex for everything
	mutex sync.Mutex
	// [+] Reconstructed Polynomial
	recPoly *polyring.Polynomial

	// Proactivization Phase
	// [+] Lagrange Coefficient
	lambda []*gmp.Int
	// [+] Zero Shares
	zeroShares []*gmp.Int
	// [+] Counter of Messages Received
	zeroCnt *int
	// [+] Zero Share
	zeroShare *gmp.Int
	// [+] Proactivization Polynomial
	proPoly *polyring.Polynomial
	// [+] Commitment & Witness in Phase 2
	zeroShareCmt *pbc.Element
    zeroPolyCmt  *pbc.Element
    zeroPolyWit  *pbc.Element

	// Share Distribution Phase
	// [+] New Poynomials
	newPoly *polyring.Polynomial
	// [+] Counter for New Secret Shares
	shareCnt     *int

	// Commitment and Witness from BulletinBoard
	oldPolyCmt []*pbc.Element
    zerosumShareCmt []*pbc.Element
    zerosumPolyCmt  []*pbc.Element
    zerosumPolyWit  []*pbc.Element
	midPolyCmt []*pbc.Element
	newPolyCmt []*pbc.Element

	// Metrics
	totMsgSize *int
	s1 *time.Time
	e1 *time.Time
	s2 *time.Time
	e2 *time.Time
	s3 *time.Time
	e3 *time.Time

	// gRPC Clients and Server
	bConn   *grpc.ClientConn
	nConn   []*grpc.ClientConn
	bClient pb.BulletinBoardServiceClient
	nClient []pb.NodeServiceClient

	// Initialize Flag
	iniflag *bool
}

// Start Phase 1
// Call ClientSharePhase1 to share secret shares with other nodes
func (node *Node) StartPhase1(ctx context.Context, in *pb.EmptyMsg) (*pb.AckMsg, error) {
	log.Printf("[node %d] start phase 1", node.label)
	*node.s1 = time.Now()
	node.ClientSharePhase1()
	return &pb.AckMsg{}, nil
}

// Share Phase 1
// The server function which takes the sent message of secret shares and store it locally. Then it starts ClientReadPhase1 to read the commitments of polynomials on bulletinboard.
func (node *Node) SharePhase1(ctx context.Context, msg *pb.PointMsg) (*pb.AckMsg, error) {
	*node.totMsgSize = *node.totMsgSize + proto.Size(msg)
	index := msg.GetIndex()
	log.Printf("[node %d] receives point message from [node %d] in phase 1", node.label, index)
	x := msg.GetX()
	y := gmp.NewInt(0)
	y.SetBytes(msg.Y)
	witness := node.dpc.NewG1()
	witness.SetCompressedBytes(msg.Witness)
	p := polypoint.PolyPoint{
		X:       x,
		Y:       y,
		PolyWit: witness,
	}
	node.mutex.Lock()
	node.recShares[*node.recCnt] = &p
	*node.recCnt = *node.recCnt + 1
	flag := (*node.recCnt == node.counter)
	node.mutex.Unlock()
	if flag {
		*node.recCnt = 0
		node.ClientReadPhase1()
	}
	return &pb.AckMsg{}, nil
}

// Share Phase 2
// The server function which takes the sent message of zero shares and sum them up to get the final share and generate the proactivization polynomial according to the zero share. It then calls ClientWritePhase2 to write the commitment of zeroshare, zeropolynomial and the witness at zero on the bulletinboard.
func (node *Node) SharePhase2(ctx context.Context, msg *pb.ZeroMsg) (*pb.AckMsg, error) {
	*node.totMsgSize = *node.totMsgSize + proto.Size(msg)
	index := msg.GetIndex()
	log.Printf("[node %d] receive zero message from [node %d] in phase 2", node.label, index)
	inter := gmp.NewInt(0)
	inter.SetBytes(msg.GetShare())
	node.mutex.Lock()
	node.zeroShare.Add(node.zeroShare, inter)
	*node.zeroCnt = *node.zeroCnt + 1
	flag := (*node.zeroCnt == node.counter)
	node.mutex.Unlock()
	if flag {
		*node.zeroCnt = 0
		node.zeroShare.Mod(node.zeroShare, node.p)
		node.dc.Commit(node.zeroShareCmt, node.zeroShare)
		poly, _ := polyring.NewRand(node.degree, node.randState, node.p)
		poly.SetCoefficient(0, 0)
		node.dpc.Commit(node.zeroPolyCmt, poly)
		node.dpc.CreateWitness(node.zeroPolyWit, poly, gmp.NewInt(0))

		poly.SetCoefficientBig(0, node.zeroShare)
		node.proPoly.ResetTo(poly.DeepCopy())

		node.ClientWritePhase2()
	}
	return &pb.AckMsg{}, nil
}

// After the bulletinboard has received the writing of all nodes, it will start a client call to this function telling the nodes to read the commitment on it.
func (node *Node) StartVerifPhase2(ctx context.Context, in *pb.EmptyMsg) (*pb.AckMsg, error) {
	log.Printf("[node %d] start verification in phase 2", node.label)
	node.ClientReadPhase2()
	return &pb.AckMsg{}, nil
}

// Share Phase 3
// The server function which takes the sent message in share distribution phase and store it locally as the new secret shares. It then calls ClientWritePhase3 to write the commitment of the new polynomial on the bulletinboard.
func (node *Node) SharePhase3(ctx context.Context, msg *pb.PointMsg) (*pb.AckMsg, error) {
	*node.totMsgSize = *node.totMsgSize + proto.Size(msg)
	index := msg.GetIndex()
	log.Printf("[node %d] receive point message from [node %d] in phase3", node.label, index)
	Y := msg.GetY()
	witness := msg.GetWitness()
	node.secretShares[index-1].Y.SetBytes(Y)
	node.secretShares[index-1].PolyWit.SetCompressedBytes(witness)
	node.mutex.Lock()
	*node.shareCnt = *node.shareCnt + 1
	flag := (*node.shareCnt == node.counter)
	node.mutex.Unlock()
	if flag {
		*node.shareCnt = 0
		node.ClientWritePhase3()
	}
	return &pb.AckMsg{}, nil
}

// After the bulletinboard receives all the writings in phase 3. It will start a client call to this function notifying the nodes to read the new commitment and verify its own shares.
func (node *Node) StartVerifPhase3(ctx context.Context, in *pb.EmptyMsg) (*pb.AckMsg, error) {
	log.Printf("[node %d] start verification in phase 3", node.label)
	node.ClientReadPhase3()
	return &pb.AckMsg{}, nil
}

func (node *Node) Connect() {
	bConn, err := grpc.Dial(node.bip, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("node did not connect to bulletinboard: %v", err)
	}
	node.bConn = bConn
	node.bClient = pb.NewBulletinBoardServiceClient(node.bConn)
	for i := 0; i < node.counter; i++ {
		if i != node.label-1 {
			nConn, err := grpc.Dial(node.ipList[i], grpc.WithInsecure())
			if err != nil {
				log.Fatalf("node did not connect to node: %v", err)
			}
			node.nConn[i] = nConn
			node.nClient[i] = pb.NewNodeServiceClient(nConn)
		}
	}
}

func (node *Node) Disconnect() {
	node.bConn.Close()
	for i := 0; i < node.counter; i++ {
		if i != node.label-1 {
			node.nConn[i].Close()
		}
	}
}

func (node *Node) Serve(aws bool) {
	port := node.ipList[node.label-1]
	if aws {
		port = "0.0.0.0:12001"
	}
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("node failed to listen %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterNodeServiceServer(s, node)
	reflection.Register(s)
	log.Printf("node %d serve on %s", node.label, port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("node failed to serve %v", err)
	}
}

// The function that starts client calls to all other nodes to send the secret shares.
func (node *Node) ClientSharePhase1() {
	if *node.iniflag {
		node.Connect()
		*node.iniflag = false
	}
	p := polypoint.PolyPoint{
		X:       node.secretShares[node.label-1].X,
		Y:       node.secretShares[node.label-1].Y,
		PolyWit: node.secretShares[node.label-1].PolyWit,
	}
	// Race Condition Here
	node.mutex.Lock()
	node.recShares[*node.recCnt] = &p
	*node.recCnt = *node.recCnt + 1
	flag := (*node.recCnt == node.counter)
	node.mutex.Unlock()
	if flag {
		*node.recCnt = 0
		node.ClientReadPhase1()
	}
	var wg sync.WaitGroup
	for i := 0; i < node.counter; i++ {
		if i != node.label-1 {
			log.Printf("[node %d] send point message to [node %d] in phase 1", node.label, i+1)
            x := node.secretShares[i].X
            y := node.secretShares[i].Y.Bytes()
            witness := node.secretShares[i].PolyWit.CompressedBytes()
            msg := &pb.PointMsg{
                X:       x,
                Y:       y,
                Witness: witness,
            }
			wg.Add(1)
			go func(i int, msg *pb.PointMsg) {
				defer wg.Done()
			    ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				node.nClient[i].SharePhase1(ctx, msg)
			} (i, msg)
		}
	}
	wg.Wait()
}

// Read from the bulletinboard and does the interpolation and verifiication.
func (node *Node) ClientReadPhase1() {
	if *node.iniflag {
		node.Connect()
		*node.iniflag = false
	}
	log.Printf("[node %d] read bulletinboard in phase 1", node.label)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := node.bClient.ReadPhase1(ctx, &pb.EmptyMsg{})
	if err != nil {
		log.Fatalf("client failed to read phase1: %v", err)
	}
	for i := 0; i < node.counter; i++ {
		msg, err := stream.Recv()
		*node.totMsgSize = *node.totMsgSize + proto.Size(msg)
		if err != nil {
			log.Fatalf("client failed to receive in read phase1: %v", err)
		}
		index := msg.GetIndex()
		polycmt := msg.GetPolycmt()
		node.oldPolyCmt[index-1].SetCompressedBytes(polycmt)
	}
	x := make([]*gmp.Int, 0)
	y := make([]*gmp.Int, 0)
	polyCmt := node.dpc.NewG1()
	polyCmt.Set(node.oldPolyCmt[node.label-1])
	for i := 0; i <= node.degree; i++ {
		point := node.recShares[i]
		x = append(x, gmp.NewInt(int64(point.X)))
		y = append(y, point.Y)
		if !node.dpc.VerifyEval(polyCmt, gmp.NewInt(int64(point.X)), point.Y, point.PolyWit) {
			panic("Reconstruction Verification failed")
		}
	}
	poly, err := interpolation.LagrangeInterpolate(node.degree, x, y, node.p)
	if err != nil {
		for i:=0; i<len(x); i++ {
			log.Print(x[i])
			log.Print(y[i])
		}
		log.Print(err)
		panic("Interpolation failed")
	}
	node.recPoly.ResetTo(poly)
	*node.e1 = time.Now()
	*node.s2 = time.Now()
	node.ClientSharePhase2()
}

// The function that really does the work of generating and sending zero shares.
func (node *Node) ClientSharePhase2() {
	// Generate Random Numbers
	for i := 0; i < node.counter-1; i++ {
		node.zeroShares[i].Rand(node.randState, gmp.NewInt(10))
		inter := gmp.NewInt(0)
		inter.Mul(node.zeroShares[i], node.lambda[i])
		node.zeroShares[node.counter-1].Sub(node.zeroShares[node.counter-1], inter)
	}
	node.zeroShares[node.counter-1].Mod(node.zeroShares[node.counter-1], node.p)
	inter := gmp.NewInt(0)
	inter.ModInverse(node.lambda[node.counter-1], node.p)
	node.zeroShares[node.counter-1].Mul(node.zeroShares[node.counter-1], inter)
	node.zeroShares[node.counter-1].Mod(node.zeroShares[node.counter-1], node.p)
	node.mutex.Lock()
	node.zeroShare.Add(node.zeroShare, node.zeroShares[node.label-1])
	*node.zeroCnt = *node.zeroCnt + 1
	flag := (*node.zeroCnt == node.counter)
	node.mutex.Unlock()
	if flag {
		*node.zeroCnt = 0
		node.zeroShare.Mod(node.zeroShare, node.p)
		node.dc.Commit(node.zeroShareCmt, node.zeroShare)
		poly, _ := polyring.NewRand(node.degree, node.randState, node.p)
		poly.SetCoefficient(0, 0)
		node.dpc.Commit(node.zeroPolyCmt, poly)
		node.dpc.CreateWitness(node.zeroPolyWit, poly, gmp.NewInt(0))

		poly.SetCoefficientBig(0, node.zeroShare)
		node.proPoly.ResetTo(poly.DeepCopy())
		node.ClientWritePhase2()
	}
    var wg sync.WaitGroup
	for i := 0; i < node.counter; i++ {
		if i != node.label-1 {
			log.Printf("[node %d] send message to [node %d] in phase 2", node.label, i+1)
			msg := &pb.ZeroMsg{
				Index: int32(node.label),
				Share: node.zeroShares[i].Bytes(),
			}
			wg.Add(1)
			go func(i int, msg *pb.ZeroMsg) {
				defer wg.Done()
	            ctx, cancel := context.WithCancel(context.Background())
		        defer cancel()
				node.nClient[i].SharePhase2(ctx, msg)
			} (i, msg)
		}
	}
	wg.Wait()
}

func (node *Node) ClientWritePhase2() {
	log.Printf("[node %d] write bulletinboard in phase 2", node.label)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	msg := &pb.Cmt2Msg{
		Index:       int32(node.label),
		Sharecmt:    node.zeroShareCmt.CompressedBytes(),
		Polycmt:     node.zeroPolyCmt.CompressedBytes(),
		Zerowitness: node.zeroPolyWit.CompressedBytes(),
	}
	node.bClient.WritePhase2(ctx, msg)
}

// Read from bulletinboard and does the verification in phase 2.
func (node *Node) ClientReadPhase2() {
	log.Printf("[node %d] read bulletinboard in phase 2", node.label)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := node.bClient.ReadPhase2(ctx, &pb.EmptyMsg{})
	if err != nil {
		log.Fatalf("client failed to read phase2: %v", err)
	}
	for i := 0; i < node.counter; i++ {
		msg, err := stream.Recv()
		*node.totMsgSize = *node.totMsgSize + proto.Size(msg)
		if err != nil {
			log.Fatalf("client failed to receive in read phase1: %v", err)
		}
		index := msg.GetIndex()
		sharecmt := msg.GetSharecmt()
		polycmt := msg.GetPolycmt()
		zerowitness := msg.GetZerowitness()
		node.zerosumShareCmt[index-1].SetCompressedBytes(sharecmt)
		inter := node.dpc.NewG1()
		inter.SetString(node.zerosumShareCmt[index-1].String(), 10)
		node.zerosumPolyCmt[index-1].SetCompressedBytes(polycmt)
		node.midPolyCmt[index-1].Mul(inter, node.zerosumPolyCmt[index-1])
		node.zerosumPolyWit[index-1].SetCompressedBytes(zerowitness)
	}
	exponentSum := node.dc.NewG1()
	exponentSum.Set1()
	for i := 0; i < node.counter; i++ {
		lambda := big.NewInt(0)
		lambda.SetString(node.lambda[i].String(), 10)
		tmp := node.dc.NewG1()
		tmp.PowBig(node.zerosumShareCmt[i], lambda)
		// log.Printf("label: %d #share %d\nlambda %s\nzeroshareCmt %s\ntmp %s", node.label, i+1, lambda.String(), node.zerosumShareCmt[i].String(), tmp.String())
		exponentSum.Mul(exponentSum, tmp)
	}
	// log.Printf("%d exponentSum: %s", node.label, exponentSum.String())
	if !exponentSum.Is1() {
		panic("Proactivization Verification 1 failed")
	}
	flag := true
	for i := 0; i < node.counter; i++ {
		if !node.dpc.VerifyEval(node.zerosumPolyCmt[i], gmp.NewInt(0), gmp.NewInt(0), node.zerosumPolyWit[i]) {
			flag = false
		}
	}
	if !flag {
		panic("Proactivization Verification 2 failed")
	}
	*node.e2 = time.Now()
	*node.s3 = time.Now()
	node.ClientSharePhase3()
}

// The function that does the real work of sending new secret shares to all nodes.
func (node *Node) ClientSharePhase3() {
	node.newPoly.Add(*node.recPoly, *node.proPoly)
	var wg sync.WaitGroup
	for i := 0; i < node.counter; i++ {
		eval := gmp.NewInt(0)
		node.newPoly.EvalMod(gmp.NewInt(int64(i+1)), node.p, eval)
		witness := node.dpc.NewG1()
		node.dpc.CreateWitness(witness, *node.newPoly, gmp.NewInt(int64(i+1)))
		if i != node.label-1 {
			log.Printf("[node %d] send point message to [node %d] in phase 3", node.label, i+1)
			msg := &pb.PointMsg{
				Index:   int32(node.label),
				X:       int32(i+1),
				Y:       eval.Bytes(),
				Witness: witness.CompressedBytes(),
			}
			wg.Add(1)
			go func(i int, msg *pb.PointMsg) {
	            ctx, cancel := context.WithCancel(context.Background())
		        defer cancel()
				defer wg.Done()
				node.nClient[i].SharePhase3(ctx, msg)
			} (i, msg)
		} else {
			node.secretShares[i].Y.Set(eval)
			node.secretShares[i].PolyWit.Set(witness)
			node.mutex.Lock()
			*node.shareCnt = *node.shareCnt + 1
			flag := (*node.shareCnt == node.counter)
			node.mutex.Unlock()
			if flag {
				*node.shareCnt = 0
				node.ClientWritePhase3()
			}
		}
	}
}

func (node *Node) ClientWritePhase3() {
	log.Printf("[node %d] write bulletinboard in phase 3", node.label)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	C := node.dpc.NewG1()
	node.dpc.Commit(C, *node.newPoly)
	msg := &pb.Cmt1Msg{
		Index:   int32(node.label),
		Polycmt: C.CompressedBytes(),
	}
	node.bClient.WritePhase3(ctx, msg)
}

// Read from the bulletinboard and do the verification in phase 3.
func (node *Node) ClientReadPhase3() {
	log.Printf("[node %d] read bulletinboard in phase 3", node.label)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := node.bClient.ReadPhase3(ctx, &pb.EmptyMsg{})
	if err != nil {
		log.Fatalf("client failed to read phase3: %v", err)
	}
	for i := 0; i < node.counter; i++ {
		msg, err := stream.Recv()
		*node.totMsgSize = *node.totMsgSize + proto.Size(msg)
		if err != nil {
			log.Fatalf("client failed to receive in read phase1: %v", err)
		}
		index := msg.GetIndex()
		polycmt := msg.GetPolycmt()
		node.newPolyCmt[index-1].SetCompressedBytes(polycmt)
	}
	for i := 0; i < node.counter; i++ {
		tmp := node.dpc.NewG1()
		if !node.newPolyCmt[i].Equals(tmp.Mul(node.oldPolyCmt[i], node.midPolyCmt[i])) {
			panic("Share Distribution Verification 1 failed")
		}
		if !node.dpc.VerifyEval(node.newPolyCmt[i], gmp.NewInt(int64(node.label)), node.secretShares[i].Y, node.secretShares[i].PolyWit) {
			panic("Share Distribution Verification 2 failed")
		}

	}
	*node.e3 = time.Now()
	f, _ := os.OpenFile(node.metadataPath + "/log" + strconv.Itoa(node.label), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	fmt.Fprintf(f, "totMsgSize,%d\n", *node.totMsgSize)
	fmt.Fprintf(f, "epochLatency,%d\n", node.e3.Sub(*node.s1).Nanoseconds())
	fmt.Fprintf(f, "reconstructionLatency,%d\n", node.e1.Sub(*node.s1).Nanoseconds())
	fmt.Fprintf(f, "proactivizationLatency,%d\n", node.e2.Sub(*node.s2).Nanoseconds())
	fmt.Fprintf(f, "sharedistLatency,%d\n", node.e3.Sub(*node.s3).Nanoseconds())
	*node.totMsgSize = 0
	for i:=0; i<node.counter; i++ {
		node.zeroShares[i].SetInt64(0)
	}
	node.zeroShare.SetInt64(0)
}

func ReadIpList(metadataPath string) []string {
	ipData, err := ioutil.ReadFile(metadataPath + "/ip_list")
	if err != nil {
		log.Fatalf("node failed to read iplist %v\n", err)
	}
	return strings.Split(string(ipData), "\n")
}

// New a Network Node Structure
func New(degree int, label int, counter int, metadataPath string) (Node, error) {
	f, _ := os.Create(metadataPath + "/log" + strconv.Itoa(label))
	defer f.Close()

	ipRaw := ReadIpList(metadataPath)[0 : counter+1]
	bip := ipRaw[0]
	ipList := ipRaw[1 : counter+1]

	if label < 0 {
		return Node{}, errors.New(fmt.Sprintf("label must be non-negative, got %d", label))
	}

	if counter < 0 {
		return Node{}, errors.New(fmt.Sprintf("counter must be non-negtive, got %d", counter))
	}

	randState := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	fixedRandState := rand.New(rand.NewSource(int64(3)))
	dc := commitment.DLCommit{}
	dc.SetupFix()
	dpc := commitment.DLPolyCommit{}
	dpc.SetupFix(counter)

	p := gmp.NewInt(0)
	p.SetString("57896044618658097711785492504343953926634992332820282019728792006155588075521", 10)
	lambda := make([]*gmp.Int, counter)
	// Calculate Lagrange Interpolation
	denominator := polyring.NewOne()
	tmp, _ := polyring.New(1)
	tmp.SetCoefficient(1, 1)
	for i := 0; i < counter; i++ {
		tmp.GetPtrToConstant().Neg(gmp.NewInt(int64(i + 1)))
		denominator.MulSelf(tmp)
	}
	for i := 0; i < counter; i++ {
		lambda[i] = gmp.NewInt(0)
		deno, _ := polyring.New(0)
		tmp.GetPtrToConstant().Neg(gmp.NewInt(int64(i + 1)))
		deno.Div2(denominator, tmp)
		deno.EvalMod(gmp.NewInt(0), p, lambda[i])
		inter := gmp.NewInt(0)
		deno.EvalMod(gmp.NewInt(int64(i+1)), p, inter)
		interInv := gmp.NewInt(0)
		interInv.ModInverse(inter, p)
		lambda[i].Mul(lambda[i], interInv)
		lambda[i].Mod(lambda[i], p)
	}

	zeroShares := make([]*gmp.Int, counter)
	for i := 0; i < counter; i++ {
		zeroShares[i] = gmp.NewInt(0)
	}
	zeroCnt := 0
	zeroShare := gmp.NewInt(0)

	recShares := make([]*polypoint.PolyPoint, counter)
	recCnt := 0

	secretShares := make([]*polypoint.PolyPoint, counter)
	poly, err := polyring.NewRand(degree, fixedRandState, p)
	for i := 0; i < counter; i++ {
		if err != nil {
			panic("Error initializing random poly")
		}
		x := int32(label)
		y := gmp.NewInt(0)
		w := dpc.NewG1()
		poly.EvalMod(gmp.NewInt(int64(x)), p, y)
		dpc.CreateWitness(w, poly, gmp.NewInt(int64(x)))
		secretShares[i] = polypoint.NewPoint(x, y, w)
	}

	proPoly, _ := polyring.New(degree)
	recPoly, _ := polyring.New(degree)
	newPoly, _ := polyring.New(degree)
	shareCnt := 0

	oldPolyCmt := make([]*pbc.Element, counter)
	midPolyCmt := make([]*pbc.Element, counter)
	newPolyCmt := make([]*pbc.Element, counter)
	for i := 0; i < counter; i++ {
		oldPolyCmt[i] = dpc.NewG1()
		midPolyCmt[i] = dpc.NewG1()
		newPolyCmt[i] = dpc.NewG1()
	}

	zeroShareCmt := dc.NewG1()
	zeroPolyCmt := dpc.NewG1()
	zeroPolyWit := dpc.NewG1()

	zerosumShareCmt := make([]*pbc.Element, counter)
	zerosumPolyCmt := make([]*pbc.Element, counter)
	zerosumPolyWit := make([]*pbc.Element, counter)

	for i := 0; i < counter; i++ {
		zerosumShareCmt[i] = dc.NewG1()
		zerosumPolyCmt[i] = dpc.NewG1()
		zerosumPolyWit[i] = dpc.NewG1()
	}

	totMsgSize := 0
	s1 := time.Now()
	e1 := time.Now()
	s2 := time.Now()
	e2 := time.Now()
	s3 := time.Now()
	e3 := time.Now()

	nConn := make([]*grpc.ClientConn, counter)
	nClient := make([]pb.NodeServiceClient, counter)

	iniflag := true
	return Node{
		metadataPath:    metadataPath,
		bip:             bip,
		ipList:          ipList,
		degree:          degree,
		label:           label,
		counter:         counter,
		randState:       randState,
		dc:              &dc,
		dpc:             &dpc,
		p:               p,
		lambda:          lambda,
		zeroShares:      zeroShares,
		zeroCnt:         &zeroCnt,
		zeroShare:       zeroShare,
		secretShares:    secretShares,
		recShares:       recShares,
		recCnt:          &recCnt,
		recPoly:         &recPoly,
		proPoly:         &proPoly,
		newPoly:         &newPoly,
		shareCnt:        &shareCnt,
		oldPolyCmt:      oldPolyCmt,
		midPolyCmt:      midPolyCmt,
		newPolyCmt:      newPolyCmt,
		zeroShareCmt:    zeroShareCmt,
		zeroPolyCmt:     zeroPolyCmt,
		zeroPolyWit:     zeroPolyWit,
		zerosumShareCmt: zerosumShareCmt,
		zerosumPolyCmt:  zerosumPolyCmt,
		zerosumPolyWit:  zerosumPolyWit,
		totMsgSize:      &totMsgSize,
		s1:				&s1,
		e1:				&e1,
		s2:				&s2,
		e2:				&e2,
		s3:				&s3,
		e3:				&e3,
		nConn:           nConn,
		nClient:         nClient,
		iniflag:         &iniflag,
	}, nil
}
