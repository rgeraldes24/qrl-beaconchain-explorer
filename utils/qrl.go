package utils

import (
	"github.com/theQRL/go-zond/common"
)

// TODO(now.youtrack.cloud/issue/TZB-2)

/*
// VerifyVoluntaryExitSignature verifies the signature of an voluntary_exit message
func VerifyVoluntaryExitSignature(op *phase0.SignedVoluntaryExit, forkVersion, pubkeyBytes []byte) error {
	currentVersion := phase0.Version{}
	genesisValidatorsRoot := phase0.Root{}
	copy(currentVersion[:], forkVersion)
	copy(genesisValidatorsRoot[:], MustParseHex(Config.Chain.GenesisValidatorsRoot))

	forkDataRoot, err := (&phase0.ForkData{
		CurrentVersion:        currentVersion,
		GenesisValidatorsRoot: genesisValidatorsRoot,
	}).HashTreeRoot()
	if err != nil {
		return fmt.Errorf("failed hashing hashtreeroot: %w", err)
	}

	domain := phase0.Domain{}
	domainVoluntaryExit := MustParseHex(Config.Chain.DomainVoluntaryExit)
	copy(domain[:], domainVoluntaryExit[:])
	copy(domain[4:], forkDataRoot[:])

	// root, err := op.Message.HashTreeRoot()
	// if err != nil {
	// 	return fmt.Errorf("failed to generate message root: %w", err)
	// }

	sigBytes := make([]byte, len(op.Signature))
	copy(sigBytes, op.Signature[:])

	// sig, err := e2types.DilithiumSignatureFromBytes(sigBytes)
	// if err != nil {
	// 	return fmt.Errorf("invalid signature: %w", err)
	// }

	// container := &phase0.SigningData{
	// 	ObjectRoot: root,
	// 	Domain:     domain,
	// }
	// signingRoot, err := ssz.HashTreeRoot(container)
	// if err != nil {
	// 	return fmt.Errorf("failed to generate signing root: %w", err)
	// }

	// pubkey, err := e2types.DilithiumPublicKeyFromBytes(pubkeyBytes)
	// if err != nil {
	// 	return fmt.Errorf("invalid public key: %w", err)
	// }
	// if !sig.Verify(signingRoot[:], pubkey) {
	// 	return fmt.Errorf("signature does not verify")
	// }

	return nil
}
*/

func FixAddressCasing(add string) string {
	addr, _ := common.NewAddressFromString(add)
	return addr.Hex()
}
