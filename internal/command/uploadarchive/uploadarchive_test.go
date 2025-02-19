package uploadarchive

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/labkit/correlation"

	"gitlab.com/gitlab-org/gitlab-shell/v14/client/testserver"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/command"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/command/readwriter"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/testhelper/requesthandlers"
)

func TestAllowedAccess(t *testing.T) {
	gitalyAddress, _ := testserver.StartGitalyServer(t, "unix")
	requests := requesthandlers.BuildAllowedWithGitalyHandlers(t, gitalyAddress)
	cmd, _ := setup(t, "1", requests)
	cmd.Config.GitalyClient.InitSidechannelRegistry(context.Background())

	correlationID := correlation.SafeRandomID()
	ctx := correlation.ContextWithCorrelation(context.Background(), correlationID)
	ctx = correlation.ContextWithClientName(ctx, "gitlab-shell-tests")
	ctxWithLogData, err := cmd.Execute(ctx)
	require.NoError(t, err)

	data := ctxWithLogData.Value("logData").(command.LogData)
	require.Equal(t, "alex-doe", data.Username)
	require.Equal(t, "group/project-path", data.Meta.Project)
	require.Equal(t, "group", data.Meta.RootNamespace)
}

func TestForbiddenAccess(t *testing.T) {
	requests := requesthandlers.BuildDisallowedByAPIHandlers(t)

	cmd, _ := setup(t, "disallowed", requests)

	_, err := cmd.Execute(context.Background())
	require.Equal(t, "Disallowed by API call", err.Error())
}

func setup(t *testing.T, keyId string, requests []testserver.TestRequestHandler) (*Command, *bytes.Buffer) {
	url := testserver.StartHttpServer(t, requests)

	output := &bytes.Buffer{}
	input := bytes.NewBufferString("input")

	cmd := &Command{
		Config:     &config.Config{GitlabUrl: url},
		Args:       &commandargs.Shell{GitlabKeyId: keyId, SshArgs: []string{"git-upload-archive", "group/repo"}},
		ReadWriter: &readwriter.ReadWriter{ErrOut: output, Out: output, In: input},
	}

	return cmd, output
}
