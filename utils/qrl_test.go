package utils

// TODO(now.youtrack.cloud/issue/TZB-2)
/*
mainnet: {"data":{"genesis_time":"1606824023","genesis_validators_root":"0x4b363db94e286120d76eb905340fdd4e54bfe9f06bf33ff6cf5ad27f511bfe95","genesis_fork_version":"0x00000000"}}
*/
/*

func TestVerifyVoluntaryExitSignature(t *testing.T) {
	Config = &types.Config{}
	ReadConfig(Config, "")
	Config.Chain.DomainVoluntaryExit = "0x04000000"
	ZhejiangGenesisForkVersion := "0x00000069"
	ZhejiangCapellaForkVersion := "0x00000072"
	ZhejiangGenesisValidatorsRoot := "0x53a92d8f2bb1d85f62d16a156e6ebcd1bcaba652d0900b2c2f387826f3481f6f"
	tests := []struct {
		CurrentForkVersion    string
		GenesisValidatorsRoot string
		Msg                   []byte
		Pubkey                []byte
		Valid                 bool
	}{
		{
			CurrentForkVersion:    ZhejiangCapellaForkVersion,
			GenesisValidatorsRoot: ZhejiangGenesisValidatorsRoot,
			Msg:                   []byte(`{"message":{"epoch":"3541","validator_index":"62019"},"signature":"0xa0f4ff61e01346b98acb7a8003df6dc5e61760adf54da5e16d138d5171cf64c2429787763973697aee47d9949108fac2106799ba96690e33263e23e079a3213d80c5617c76a350a9eb114a6afd77ec94f1cf230f38e6caae5d7209474b285fc8"}`),
			Pubkey:                MustParseHex("0x9305db483ed03b526f0f70c6201a359e8becd3f584fe6ae52242e44346a5b4f7a74c29f8dbd981cbe885a4ce6b842a11"),
			Valid:                 true,
		},
		{
			CurrentForkVersion:    ZhejiangGenesisForkVersion,
			GenesisValidatorsRoot: ZhejiangGenesisValidatorsRoot,
			Msg:                   []byte(`{"message":{"epoch":"3541","validator_index":"62019"},"signature":"0x963723375cc200ce005b284f03a07cf45775b84cca94955b28c1d4a8d7dfcfc5d9d953a8b3c9f01e337146bff6328e79067edc58cc9ea0f2ef3fa96b48a01264b5bd719ff4177fd3de37689d3df9a599c6906b7cd1952a15dcd2bbb3a2c54341"}`),
			Pubkey:                MustParseHex("0x9305db483ed03b526f0f70c6201a359e8becd3f584fe6ae52242e44346a5b4f7a74c29f8dbd981cbe885a4ce6b842a11"),
			Valid:                 true,
		},
	}
	for _, test := range tests {
		Config.Chain.GenesisValidatorsRoot = test.GenesisValidatorsRoot
		var op *phase0.SignedVoluntaryExit
		err := json.Unmarshal(test.Msg, &op)
		if err != nil {
			t.Errorf("failed unmarshaling msg: %v", err)
		}
		err = VerifyVoluntaryExitSignature(op, MustParseHex(test.CurrentForkVersion), test.Pubkey)
		if test.Valid != (err == nil) {
			t.Errorf("wrong verify, should be valid: %v, err: %v", test.Valid, err)
		}
	}
}
*/
