package medChainUtils

import (
	"io/ioutil"
	"github.com/dedis/kyber/util/key"
	"github.com/dedis/cothority"
	"strconv"
	"github.com/dedis/cothority/omniledger/darc"
	"encoding/base64"
	"github.com/dedis/cothority/omniledger/service"
	"github.com/dedis/cothority/omniledger/contracts"
	"time"
	"github.com/dedis/onet/network"
	"encoding/hex"
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func InitKeys(numKeys int, directory string) {
	for i := 0; i < numKeys; i++ {
		temp := darc.NewSignerEd25519(nil, nil)

		private, _ := temp.GetPrivate()
		privateInBytes, _ := private.MarshalBinary()
		public := temp.Identity().Ed25519.Point
		publicInBytes, _ := public.MarshalBinary()

		err := ioutil.WriteFile(directory + "/" + strconv.Itoa(i) + "_private", []byte(base64.StdEncoding.EncodeToString(privateInBytes)), 0644)
		Check(err)
		err = ioutil.WriteFile(directory + "/" + strconv.Itoa(i) + "_public", []byte(base64.StdEncoding.EncodeToString(publicInBytes)), 0644)
		Check(err)

		kp := key.NewKeyPair(cothority.Suite)
		err = kp.Public.UnmarshalBinary(publicInBytes)
		Check(err)
		err = kp.Private.UnmarshalBinary(privateInBytes)
		Check(err)
	}
}

func LoadIdentityEd25519(pathToPublic string) darc.Identity {
	dat, err := ioutil.ReadFile(pathToPublic)
	Check(err)
	return LoadIdentityEd25519FromBytes(dat)
}

func LoadIdentityEd25519FromBytes(publicBytes []byte) darc.Identity {
	kp := key.NewKeyPair(cothority.Suite)
	bin, err := base64.StdEncoding.DecodeString(string(publicBytes[:]))
	Check(err)
	err = kp.Public.UnmarshalBinary(bin)
	Check(err)
	return darc.Identity{
		Ed25519: &darc.IdentityEd25519{
			Point: kp.Public,
		},
	}
}

func LoadSignerEd25519(pathToPublic string, pathToPrivate string) darc.Signer {
	dat, err := ioutil.ReadFile(pathToPrivate)
	Check(err)
	kp := key.NewKeyPair(cothority.Suite)
	bin, err := base64.StdEncoding.DecodeString(string(dat[:]))
	Check(err)
	err = kp.Private.UnmarshalBinary(bin)
	Check(err)
	return darc.Signer{Ed25519: &darc.SignerEd25519{
		Point:  LoadIdentityEd25519(pathToPublic).Ed25519.Point,
		Secret: kp.Private,
	}}
}


func LoadSignerEd25519FromBytes(publicBytes []byte, privateBytes []byte) darc.Signer {
	kp := key.NewKeyPair(cothority.Suite)
	bin, err := base64.StdEncoding.DecodeString(string(privateBytes))
	Check(err)
	err = kp.Private.UnmarshalBinary(bin)
	Check(err)
	return darc.Signer{Ed25519: &darc.SignerEd25519{
		Point:  LoadIdentityEd25519FromBytes(publicBytes).Ed25519.Point,
		Secret: kp.Private,
	}}
}

func CreateQueryTransaction(projectDarc string, queryType string, query string, signer darc.Signer) string {
	// We don't need the "darc:" part from the ID, and a
	projectDarcDecoded, err := hex.DecodeString(projectDarc[5:])
	Check(err)

	ctx := service.ClientTransaction{
		Instructions: []service.Instruction{{
			InstanceID: service.NewInstanceID(projectDarcDecoded),
			Nonce:  service.Nonce{},
			Index:  0,
			Length: 1,
			Spawn: &service.Spawn{
				ContractID: contracts.ContractCreateQueryID,
				Args: []service.Argument{{
					Name:  "queryType",
					Value: []byte(queryType),
				}, {
					Name:  "query",
					Value: []byte(query),
				}, {
					Name:  "currentTime",
					Value: []byte(time.Now().String()),
				}},
			},
		}},
	}

	err = ctx.Instructions[0].SignBy(projectDarcDecoded, signer)
	Check(err)
	data, err := network.Marshal(&ctx)
	Check(err)
	return base64.StdEncoding.EncodeToString(data)
}

func CreateLoginTransaction(allUsersDarc string, userProjectsMap string, signer darc.Signer) string {
	allUsersDarcBytes, err := base64.StdEncoding.DecodeString(allUsersDarc)
	Check(err)
	userProjectsMapBytes, err := base64.StdEncoding.DecodeString(userProjectsMap)
	Check(err)

	ctx := service.ClientTransaction{
		Instructions: []service.Instruction{{
			InstanceID: service.NewInstanceID(allUsersDarcBytes),
			Nonce:  service.Nonce{},
			Index:  0,
			Length: 1,
			Spawn: &service.Spawn{
				ContractID: contracts.ContractProjectListID,
				Args: []service.Argument{{
					Name:  "userProjectsMapInstanceID",
					Value: userProjectsMapBytes,
				}, {
					Name:  "currentTime",
					Value: []byte(time.Now().String()),
				}},
			},
		}},
	}

	err = ctx.Instructions[0].SignBy(allUsersDarcBytes, signer)
	Check(err)
	data, err := network.Marshal(&ctx)
	Check(err)
	return base64.StdEncoding.EncodeToString(data)
}