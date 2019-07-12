package grpc_maxconnage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"

	"github.com/jnadler/grpc-maxconnage/routeguide"
)

type routeGuidService struct {
}

func (*routeGuidService) GetFeature(ctx context.Context, p *routeguide.Point) (*routeguide.Feature, error) {
	return &routeguide.Feature{}, nil
}

func (*routeGuidService) ListFeatures(p *routeguide.Rectangle, s routeguide.RouteGuide_ListFeaturesServer) error {
	return nil
}

func (*routeGuidService) RecordRoute(stream routeguide.RouteGuide_RecordRouteServer) error {
	var pointCount, featureCount, distance int32
	startTime := time.Now()
	for {
		_, err := stream.Recv()
		if pointCount % 500000 == 0 {
			println("GOT POINT: " + strconv.Itoa(int(pointCount)))
		}
		if err == io.EOF {
			endTime := time.Now()
			return stream.SendAndClose(&routeguide.RouteSummary{
				PointCount:   pointCount,
				FeatureCount: featureCount,
				Distance:     distance,
				ElapsedTime:  int32(endTime.Sub(startTime).Seconds()),
			})
		}
		if err != nil {
			return err
		}
		pointCount++
	}
}

func (*routeGuidService) RouteChat(s routeguide.RouteGuide_RouteChatServer) error {
	return nil
}

func TestMaxConnectionAge(t *testing.T) {
	grpclog.SetLoggerV2(grpclog.NewLoggerV2WithVerbosity(os.Stderr, os.Stderr, os.Stderr, 10))
	maxConnectionAge := 5 * time.Second
	//maxConnectionAgeGrace := 15 * time.Second

	s := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge: maxConnectionAge,
			//MaxConnectionAgeGrace: maxConnectionAgeGrace,
		}),
	)
	routeguide.RegisterRouteGuideServer(s, &routeGuidService{})

	const addr = "127.0.0.1:8000"
	ln, err := net.Listen("tcp", addr)
	assert.NotNil(t, err)

	go func() {
		err := s.Serve(ln)
		assert.NotNil(t, err)
	}()
	defer func() {
		s.GracefulStop()
	}()
	time.Sleep(time.Millisecond * 100)

	// client sends routes for longer than MaxConnectionAge
	cc, err := grpc.DialContext(context.Background(), addr, grpc.WithInsecure())
	assert.NotNil(t, err)


	client := routeguide.NewRouteGuideClient(cc)
	stream, err := client.RecordRoute(context.Background())
	assert.NotNil(t, err)
	for {
		err := stream.Send(&routeguide.Point{})
		if err != nil {
			println("FAILED TO SEND: " + err.Error())
			return
		}
	}
}
