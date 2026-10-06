package main

import (
	"context"
	"testing"

	basev0 "github.com/codefly-dev/core/generated/go/codefly/base/v0"
	agentv0 "github.com/codefly-dev/core/generated/go/codefly/services/agent/v0"
	builderv0 "github.com/codefly-dev/core/generated/go/codefly/services/builder/v0"
)

func TestAdvertisesNoWorkloadValidation(t *testing.T) {
	info, err := NewService().GetAgentInformation(context.Background(), &agentv0.AgentInformationRequest{})
	if err != nil {
		t.Fatal(err)
	}
	// The agent repository owns its Go suites. The managed upstream server
	// exposes no workload source or suite for a consuming workspace to run.
	if info.Validation == nil {
		t.Fatal("an absent contract triggers legacy runtime probing")
	}
	if info.Validation.GetLint().GetSupported() || info.Validation.GetCompile().GetSupported() || info.Validation.GetTest().GetSupported() {
		t.Fatal("managed upstream server must not advertise workload validation")
	}
}

func TestBuilderValidationRemainsAvailable(t *testing.T) {
	info, err := NewService().GetAgentInformation(context.Background(), &agentv0.AgentInformationRequest{})
	if err != nil {
		t.Fatal(err)
	}
	validation := info.GetValidation()
	for name, operation := range map[string]*agentv0.ValidationOperationCapability{
		"audit": validation.GetAudit(), "build": validation.GetArtifactBuild(),
		"sbom": validation.GetSbom(), "sync": validation.GetSync(),
	} {
		if !operation.GetSupported() {
			t.Errorf("%s lost its supported operation", name)
		}
	}
	builder := NewBuilder()
	if err := builder.HeadlessLoad(context.Background(), &basev0.ServiceIdentity{Workspace: "workspace", Module: "module", Name: "service"}); err != nil {
		t.Fatal(err)
	}
	response, err := builder.Sync(context.Background(), &builderv0.SyncRequest{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if response.GetState().GetState() != builderv0.SyncStatus_SUCCESS || len(response.GetChangedFiles()) != 0 {
		t.Fatalf("unexpected sync result: %v", response)
	}
}
