package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/csi-addons/spec/lib/go/identity"
)

// IdentityServer struct of rbd CSI driver with supported methods of CSI
// identity server spec.
type IdentityServer struct {
	*identity.UnimplementedIdentityServer
}

// Probe is called by the CO plugin to validate that the CSI-Addons Node is
// still healthy.
func (is *IdentityServer) Probe(
	ctx context.Context,
	req *identity.ProbeRequest) (*identity.ProbeResponse, error) {
	// there is nothing that would cause a delay in getting ready
	res := &identity.ProbeResponse{
		Ready: &wrapperspb.BoolValue{Value: true},
	}

	return res, nil
}
func Serve() {
	ip, _ := os.LookupEnv("POD_IP")
	port, _ := os.LookupEnv("PORT")
	address := ip + ":" + port
	fmt.Println(address)
	lis, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	fmt.Printf("Server is listening on %v ...", lis.Addr())

	s := grpc.NewServer()
	identity.RegisterIdentityServer(s, &IdentityServer{})
	err = s.Serve(lis)
	fmt.Println(err)
}
