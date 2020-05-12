package app

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/onet/v3/log"
)

func TestInteractiveConfig(t *testing.T) {
	registerService()
	defer unregisterService()

	tmp, err := ioutil.TempDir("", "conode")
	log.ErrFatal(err)

	setInput("127.0.0.1:2000\nConode1\n" + tmp)
	InteractiveConfig(suites.MustFind("Ed25519"), tmp+"/config.bin")

	cc, _, err := ParseCothority(tmp + "/private.toml")
	require.NoError(t, err)
	require.NotNil(t, cc.Services[testServiceName])
	require.Equal(t, cc.Description, "Conode1")
	require.Equal(t, cc.Address.String(), "tls://127.0.0.1:2000")

	gFile, err := os.Open(tmp + "/public.toml")
	require.NoError(t, err)
	gc, err := ReadGroupDescToml(gFile)
	require.NoError(t, err)
	require.Equal(t, 1, len(gc.Roster.List))
	require.Equal(t, 1, len(gc.Roster.List[0].ServiceIdentities))
	require.Equal(t, "bn256.adapter", gc.Roster.List[0].ServiceIdentities[0].Suite)

	log.ErrFatal(os.RemoveAll(tmp))
}
