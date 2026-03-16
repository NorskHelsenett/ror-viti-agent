package rorclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/NorskHelsenett/ror/pkg/clients/rorclient"
	"github.com/NorskHelsenett/ror/pkg/clients/rorclient/v2/transports/resttransport"
	"github.com/NorskHelsenett/ror/pkg/clients/rorclient/v2/transports/resttransport/httpauthprovider"
	"github.com/NorskHelsenett/ror/pkg/clients/rorclient/v2/transports/resttransport/httpclient"
	"github.com/NorskHelsenett/ror/pkg/config/rorversion"
)

type RorClient struct {
	Client  rorclient.RorClient
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
		Client:  *rorclient.NewRorClient(transport),
		Context: ctx,
	}
	err = rorClient.Client.CheckConnection()

	if err != nil {
		return nil, err
	}

	return &rorClient, nil
}
