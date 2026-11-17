package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const (
	testASGPath = "/api/v2.1/autoscaling-group/domain/test-org/project/test-project/autoscaling-group/auto_scaling_group_with_lb/asg-1"
	testLBPath  = "/api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers"
)

// asgDeleteServer fakes the autoscaling and load balancer APIs. The group GET
// reports a virtual server when withVS is set, and the group DELETE is refused
// with the platform's real 400 while that virtual server still exists.
func asgDeleteServer(t *testing.T, withVS bool) (*Client, *[]string) {
	t.Helper()
	var (
		mu       sync.Mutex
		calls    []string
		vsExists = withVS
	)
	record := func(r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, r.Method+" "+r.URL.Path)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		record(r)
		mu.Lock()
		defer mu.Unlock()
		switch {
		case r.Method == http.MethodGet && r.URL.Path == testASGPath:
			vs := ""
			if withVS {
				vs = `{"id":"vs-1","name":"app-vs","lb_service":{"id":"lb-1","name":"lb"}}`
			} else {
				vs = `{}`
			}
			_, _ = fmt.Fprintf(w, `{"id":"asg-1","name":"app","status":"CREATE_COMPLETE","virtual_server":%s}`, vs)
		case r.Method == http.MethodGet && r.URL.Path == testLBPath+"/lb_service/lb-1":
			_, _ = fmt.Fprint(w, `{"id":"lb-1","name":"lb","network_id":"lb-net"}`)
		case r.Method == http.MethodDelete && r.URL.Path == testLBPath+"/lb-1/virtual-servers/vs-1":
			vsExists = false
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == testLBPath+"/lb-1/virtual-servers/vs-1":
			if vsExists {
				_, _ = fmt.Fprint(w, `{"id":"vs-1","status":"Active"}`)
				return
			}
			http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
		case r.Method == http.MethodDelete && r.URL.Path == testASGPath:
			if vsExists {
				http.Error(w, `{"detail":"Auto Scaling Group app is in use by a virtual server."}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	c, err := NewClient(ts.URL, "test-api-key", "test-api-secret", "north", "test-org", "test-project", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return c, &calls
}

func indexOfCall(calls []string, want string) int {
	for i, call := range calls {
		if call == want {
			return i
		}
	}
	return -1
}

func TestDeleteAutoScalingGroupWithVirtualServerRemovesVirtualServerFirst(t *testing.T) {
	c, calls := asgDeleteServer(t, true)

	if err := c.DeleteAutoScalingGroupWithVirtualServer(context.Background(), "asg-1", time.Minute); err != nil {
		t.Fatalf("DeleteAutoScalingGroupWithVirtualServer() error = %v", err)
	}

	vsDelete := indexOfCall(*calls, "DELETE "+testLBPath+"/lb-1/virtual-servers/vs-1")
	asgDelete := indexOfCall(*calls, "DELETE "+testASGPath)
	if vsDelete < 0 || asgDelete < 0 {
		t.Fatalf("expected both deletes, calls = %v", *calls)
	}
	if vsDelete > asgDelete {
		t.Fatalf("virtual server must be deleted before the group, calls = %v", *calls)
	}
}

func TestDeleteAutoScalingGroupWithVirtualServerSkipsLoadBalancerWhenNone(t *testing.T) {
	c, calls := asgDeleteServer(t, false)

	if err := c.DeleteAutoScalingGroupWithVirtualServer(context.Background(), "asg-1", time.Minute); err != nil {
		t.Fatalf("DeleteAutoScalingGroupWithVirtualServer() error = %v", err)
	}
	for _, call := range *calls {
		if strings.Contains(call, "load-balancers") {
			t.Fatalf("unexpected load balancer call %q, calls = %v", call, *calls)
		}
	}
	if indexOfCall(*calls, "DELETE "+testASGPath) < 0 {
		t.Fatalf("group was not deleted, calls = %v", *calls)
	}
}

func TestDeleteAutoScalingGroupWithVirtualServerNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()
	c, err := NewClient(ts.URL, "test-api-key", "test-api-secret", "north", "test-org", "test-project", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	err = c.DeleteAutoScalingGroupWithVirtualServer(context.Background(), "asg-1", time.Minute)
	if !IsNotFoundError(err) {
		t.Fatalf("error = %v, want a not-found error", err)
	}
}

func TestGetAutoScalingGroupVirtualServerNone(t *testing.T) {
	c, _ := asgDeleteServer(t, false)

	vs, err := c.GetAutoScalingGroupVirtualServer(context.Background(), &models.AutoScalingGroup{})
	if err != nil || vs != nil {
		t.Fatalf("got (%v, %v), want (nil, nil)", vs, err)
	}
}

func TestAutoScalingForm(t *testing.T) {
	req := &models.CreateASGRequest{
		Name: "app-asg", ImageID: 2435, FlavorID: 101, VPCID: "vpc", NetworkID: "network",
		SecurityGroupID: 3661, AvailabilityZone: "S1", ScalingGroupDesiredSize: 1,
		ScalingGroupMaxSize: 3, DesiredCount: 1, ScaleUpStepSize: 1, ScaleDownStepSize: 1,
		ScalingInterval: 2, CooloffPeriod: 10, TerminationPolicy: "oldest",
		BootVolumeSize: 100, Labels: []string{"app", "prod"},
		ScaleUpRules:   []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 60}},
		ScaleDownRules: []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 30}},
		VSConfig: &models.ASGVSConfig{
			LBServiceID: "lb-id", Name: "app-vs", VIPPortID: 75469, Protocol: "TCP",
			Port: 1234, RoutingAlgorithm: "ROUND_ROBIN", PoolName: "pool", PoolPort: 12,
			Interval: 5, Timeout: 16, XForwardedFor: true, MonitorProtocol: "TCP", MaxConn: 100,
		},
	}

	form, err := autoScalingForm(req)
	if err != nil {
		t.Fatalf("autoScalingForm() error = %v", err)
	}
	if form["policy"] != "metrics" {
		t.Fatalf("policy = %v, want metrics", form["policy"])
	}
	if form["labels"] != "app,prod" {
		t.Fatalf("labels = %v, want app,prod", form["labels"])
	}
	up, ok := form["scaleup_rules"].([]string)
	if !ok || len(up) != 1 || !strings.Contains(up[0], `"metric_type":"cpu"`) {
		t.Fatalf("scaleup_rules = %#v", form["scaleup_rules"])
	}
	vs, ok := form["vs_config"].(string)
	if !ok || !strings.Contains(vs, `"x_forwarded_for":true`) || !strings.Contains(vs, `"vip_port_id":75469`) {
		t.Fatalf("vs_config = %#v", form["vs_config"])
	}
}

func TestAutoScalingFormOmitsOptionalSources(t *testing.T) {
	form, err := autoScalingForm(&models.CreateASGRequest{})
	if err != nil {
		t.Fatalf("autoScalingForm() error = %v", err)
	}
	for _, field := range []string{"image_id", "snapshot_id", "vs_config", "labels"} {
		if _, ok := form[field]; ok {
			t.Fatalf("field %q unexpectedly present", field)
		}
	}
}
