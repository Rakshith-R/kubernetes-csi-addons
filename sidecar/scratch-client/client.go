package main

import (
	"context"
	"fmt"
	"log"

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
func main() {
	fmt.Println("Hello client ...")

	opts := grpc.WithInsecure()
	cc, err := grpc.Dial("localhost:8081", opts)
	if err != nil {
		log.Fatal(err)
	}
	defer cc.Close()

	client := identity.NewIdentityClient(cc)
	req := identity.ProbeRequest{}
	resp, err := client.Probe(context.Background(), &req)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Receive response => [%v]\n", resp.Ready)
}
