package rorclient

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/NorskHelsenett/ror/pkg/clients/rorclient"
	"github.com/NorskHelsenett/ror/pkg/clients/rorclient/v2/transports/resttransport"
	"github.com/NorskHelsenett/ror/pkg/clients/rorclient/v2/transports/resttransport/httpauthprovider"
	"github.com/NorskHelsenett/ror/pkg/clients/rorclient/v2/transports/resttransport/httpclient"
	"github.com/NorskHelsenett/ror/pkg/config/rorversion"
	"github.com/NorskHelsenett/ror/pkg/rorresources"
)

type RorClient struct {
	Client  *rorclient.RorClient
	Cache   *rorcache
	Context context.Context
}

type UpdateError struct {
	Errors []UpdateInstanceError
}

func (u UpdateError) Error() string {
	builder := strings.Builder{}
	_, err := fmt.Fprintf(&builder, "total count: %v. ", len(u.Errors))
	if err != nil {
		return fmt.Sprintf("failed to build error string. %s", err)
	}
	for _, err := range u.Errors {
		_, werr := fmt.Fprintf(&builder, "uuid: %v, status: %v, message: %v. ", err.UUID, err.status, err.message)
		if werr != nil {
			return fmt.Sprintf("failed to build error string. %s", werr)
		}
	}
	return builder.String()
}

type UpdateInstanceError struct {
	UUID    string
	message string
	status  int
}

type RorDeleteAction struct {
	Err     error
	UUID    string
	Message string
	Exists  bool
}

func NewRorDeleteAction(uuid string) *RorDeleteAction {
	return &RorDeleteAction{
		UUID:    uuid,
		Exists:  false,
		Message: "",
		Err:     nil,
	}
}

// NewRorClient creates a new rorclient wrapper
//
// Source is the source url used in storing the resources, this has to match the url used in gathering the resources.
func NewRorClient(ctx context.Context, apikey, url, role, version, commit string) (*RorClient, error) {
	apiKeyProvider := httpauthprovider.NewAuthProvider(httpauthprovider.AuthPoviderTypeAPIKey, apikey)
	clientConfig, err := httpclient.NewHttpTransportClientConfig(
		url,
		apiKeyProvider,
		role,
		rorversion.NewRorVersion(version, commit),
	)
	if err != nil {
		return nil, err
	}

	transport := resttransport.NewRorHttpTransport(clientConfig)

	rorClient := RorClient{
		Client:  rorclient.NewRorClient(transport),
		Cache:   newRorCache(ctx),
		Context: ctx,
	}
	err = rorClient.Client.CheckConnection()

	if err != nil {
		return nil, err
	}

	return &rorClient, nil
}

func (r *RorClient) UpdateRorResources(rorresources []*rorresources.Resource) error {

	sets := chunkResourceToSet(rorresources, 49)

	errs := []error{}
	for _, set := range sets {
		_, err := r.saveRorSet(set)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to send rorsets to ror. %v", errs)
	}

	return nil
}

func (r *RorClient) saveRorSet(rorSet *rorresources.ResourceSet) (*rorresources.ResourceUpdateResults, error) {

	results, err := r.updateRorResource(rorSet)
	if err != nil {
		return results, fmt.Errorf("failed to update ror resource(s). %w", err)
	}

	return results, nil
}

func (r *RorClient) updateRorResource(rorSet *rorresources.ResourceSet) (*rorresources.ResourceUpdateResults, error) {

	response, err := r.Client.ResourcesV2().Update(r.Context, rorSet)
	if err != nil {
		return nil, fmt.Errorf("failed to update %v ror resource(s). %w", rorSet.Len(), err)
	}
	var errors UpdateError
	// This is only to please the linter
	anyGoodResponseCodes := 299
	for UUID, response := range response.Results {
		if response.Status > anyGoodResponseCodes {
			updateErr := UpdateInstanceError{
				UUID:    UUID,
				status:  response.Status,
				message: response.Message,
			}

			errors.Errors = append(errors.Errors, updateErr)
		}
	}

	if len(errors.Errors) != 0 {
		return response, errors
	}

	return response, nil
}

func chunkResourceToSet(resources []*rorresources.Resource, size int) []*rorresources.ResourceSet {

	output := make([]*rorresources.ResourceSet, 0)
	resourcechunks := chunkToMax(resources, size)

	for _, resourcechunk := range resourcechunks {
		set := rorresources.NewResourceSet()
		for _, resource := range resourcechunk {
			set.Add(resource)
		}
		output = append(output, set)
	}

	return output
}

// ChunkToMax slices an slice of items into chunks of maxSize.
func chunkToMax(items []*rorresources.Resource, maxSize int) [][]*rorresources.Resource {
	return slices.Collect(slices.Chunk(items, maxSize))
}
