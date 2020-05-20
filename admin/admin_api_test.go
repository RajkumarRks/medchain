package admin

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/cothority/v3"
	"go.dedis.ch/cothority/v3/byzcoin"
	"go.dedis.ch/cothority/v3/byzcoin/bcadmin/lib"
	"go.dedis.ch/cothority/v3/darc"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
)

func TestAdminClient_AddAdminsToDarc(t *testing.T) {
	// ------------------------------------------------------------------------
	// 0. Set up
	// ------------------------------------------------------------------------
	log.Lvl1("[INFO] Roster and byzcoin set up")
	local := onet.NewTCPTest(cothority.Suite)
	defer local.CloseAll()

	superAdmin := darc.NewSignerEd25519(nil, nil)
	_, roster, _ := local.GenTree(5, true)

	genesisMsg, err := byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, roster,
		[]string{"spawn:value", "spawn:deferred", "invoke:deferred.addProof",
			"invoke:deferred.execProposedTx"}, superAdmin.Identity())
	require.NoError(t, err)
	gDarc := &genesisMsg.GenesisDarc

	genesisMsg.BlockInterval = time.Second / 5
	bcl, _, err := byzcoin.NewLedger(genesisMsg, false)
	require.NoError(t, err)
	log.Lvl1("[INFO] Create admin client")
	admcl, err := NewClientWithAuth(bcl, &superAdmin)
	require.NoError(t, err)
	require.Equal(t, superAdmin, admcl.AuthKey())

	// ------------------------------------------------------------------------
	// 1. Spawn admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Spawn admin darc")
	adminDarc, err := admcl.SpawnNewAdminDarc()
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = lib.GetDarcByID(admcl.bcl, gDarc.GetID()) // Verify that the darc is really in the global state
	require.NoError(t, err)

	// ------------------------------------------------------------------------
	// 2. Add a new admin to the admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Create new admin client (admcl2)")
	admcl2, err := NewClient(bcl)
	require.NoError(t, err)
	log.Lvl1("[INFO] Create deffered tx to add admcl2 identity")
	id, err := admcl.AddAdminToAdminDarc(adminDarc.GetBaseID(), admcl2.AuthKey().Identity())
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = admcl2.bcl.GetDeferredData(id) // Verify that the deferred transaction is registered on chain
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 1 sign the transaction")
	admcl.AddSignatureToDefferedTx(id, 0)
	log.Lvl1("[INFO] Admin 1 try to exec the transaction")
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)
	log.Lvl1("[INFO] Tx successfully executed, admin 2 has been added to the admin darc")

	// ------------------------------------------------------------------------
	// 2. Add a new admin to the admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Create new admin client (admcl3)")
	admcl3, err := NewClient(bcl)
	require.NoError(t, err)
	log.Lvl1("[INFO] Create deffered tx to add admcl3 identity")
	id, err = admcl.AddAdminToAdminDarc(adminDarc.GetBaseID(), admcl3.AuthKey().Identity())
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = admcl2.bcl.GetDeferredData(id) // Verify that the deferred transaction is registered on chain
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 1 sign the transaction")
	err = admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 1 try to exec the transaction but threshold of signature not reached")
	err = admcl.ExecDefferedTx(id)
	require.Error(t, err)
	log.Lvl1("[INFO] Admin 2 sign the transaction")
	err = admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 2 try to exec the transaction")
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)
	log.Lvl1("[INFO] Tx successfully executed, admin 3 has been added to the admin darc")

	// ------------------------------------------------------------------------
	// 2. Add a new admin to the admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Create new admin client (admcl4)")
	admcl4, err := NewClient(bcl)
	require.NoError(t, err)
	log.Lvl1("[INFO] Create deffered tx to add admcl4 identity")
	id, err = admcl.AddAdminToAdminDarc(adminDarc.GetBaseID(), admcl4.AuthKey().Identity())
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = admcl2.bcl.GetDeferredData(id) // Verify that the deferred transaction is registered on chain
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 1 sign the transaction")
	err = admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 3 sign the transaction")
	err = admcl3.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 2 sign the transaction")
	err = admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 3 try to execute the transaction")
	err = admcl3.ExecDefferedTx(id)
	require.NoError(t, err)
	log.Lvl1("[INFO] Tx successfully executed, admin 4 has been added to the admin darc")
}

func TestAdminClient_RemovingAdminsFromDarc(t *testing.T) {
	// ------------------------------------------------------------------------
	// 0. Set up
	// ------------------------------------------------------------------------
	log.Lvl1("[INFO] Create admin darc")
	local := onet.NewTCPTest(cothority.Suite)
	defer local.CloseAll()

	superAdmin := darc.NewSignerEd25519(nil, nil)
	_, roster, _ := local.GenTree(5, true)

	genesisMsg, err := byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, roster,
		[]string{"spawn:value", "spawn:deferred", "invoke:deferred.addProof",
			"invoke:deferred.execProposedTx"}, superAdmin.Identity())
	require.NoError(t, err)
	gDarc := &genesisMsg.GenesisDarc

	genesisMsg.BlockInterval = time.Second / 5
	bcl, _, err := byzcoin.NewLedger(genesisMsg, false)
	require.NoError(t, err)
	log.Lvl1("[INFO] Create admin client")
	admcl, err := NewClientWithAuth(bcl, &superAdmin)
	require.NoError(t, err)
	require.Equal(t, superAdmin, admcl.AuthKey())

	// ------------------------------------------------------------------------
	// 1. Spawn admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Spawn admin darc")
	adminDarc, err := admcl.SpawnNewAdminDarc()
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = lib.GetDarcByID(admcl.bcl, gDarc.GetID()) // Verify that the darc is really in the global state
	require.NoError(t, err)

	// ------------------------------------------------------------------------
	// 2. Add a new admin to the admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Create new admin client (admcl2)")
	admcl2, err := NewClient(bcl)
	require.NoError(t, err)
	log.Lvl1("[INFO] Create deffered tx to add admcl2 identity")
	id, err := admcl.AddAdminToAdminDarc(adminDarc.GetBaseID(), admcl2.AuthKey().Identity())
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = admcl2.bcl.GetDeferredData(id) // Verify that the deferred transaction is registered on chain
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 1 sign the transaction")
	admcl.AddSignatureToDefferedTx(id, 0)
	log.Lvl1("[INFO] Admin 1 try to exec the transaction")
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)
	log.Lvl1("[INFO] Tx successfully executed, admin 2 has been added to the admin darc")

	// ------------------------------------------------------------------------
	// 2. Add a new admin to the admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Create new admin client (admcl3)")
	admcl3, err := NewClient(bcl)
	require.NoError(t, err)
	log.Lvl1("[INFO] Create deffered tx to add admcl3 identity")
	id, err = admcl.AddAdminToAdminDarc(adminDarc.GetBaseID(), admcl3.AuthKey().Identity())
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = admcl2.bcl.GetDeferredData(id) // Verify that the deferred transaction is registered on chain
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 1 sign the transaction")
	err = admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 2 sign the transaction")
	err = admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 2 try to exec the transaction")
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)
	log.Lvl1("[INFO] Tx successfully executed, admin 3 has been added to the admin darc")

	// ------------------------------------------------------------------------
	// 2. Remove admin 2 from the admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Create deffered tx to remove admcl2 identity")
	id, err = admcl.RemoveAdminFromAdminDarc(adminDarc.GetBaseID(), admcl2.AuthKey().Identity())
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = admcl2.bcl.GetDeferredData(id)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 1 sign the transaction")
	err = admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 3 sign the transaction")
	err = admcl3.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 2 sign the transaction")
	err = admcl2.AddSignatureToDefferedTx(id, 0) // For now the multisignature rule state that every admin sign for sensitive operations (in other rules admin shouldn't sign to be removed)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 3 try to execute the transaction")
	err = admcl3.ExecDefferedTx(id)
	require.NoError(t, err)
	log.Lvl1("[INFO] Tx successfully executed, admin 2 has been removed from the admin darc")
}

func TestAdminClient_UpdateAdminKeys(t *testing.T) {
	// ------------------------------------------------------------------------
	// 0. Set up
	// ------------------------------------------------------------------------
	log.Lvl1("[INFO] Create admin darc")
	local := onet.NewTCPTest(cothority.Suite)
	defer local.CloseAll()

	superAdmin := darc.NewSignerEd25519(nil, nil)
	_, roster, _ := local.GenTree(5, true)

	genesisMsg, err := byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, roster,
		[]string{"spawn:value", "spawn:deferred", "invoke:deferred.addProof",
			"invoke:deferred.execProposedTx"}, superAdmin.Identity())
	require.NoError(t, err)
	gDarc := &genesisMsg.GenesisDarc

	genesisMsg.BlockInterval = time.Second / 5
	bcl, _, err := byzcoin.NewLedger(genesisMsg, false)
	require.NoError(t, err)
	log.Lvl1("[INFO] Create admin client")
	admcl, err := NewClientWithAuth(bcl, &superAdmin)
	require.NoError(t, err)
	require.Equal(t, superAdmin, admcl.AuthKey())

	// ------------------------------------------------------------------------
	// 1. Spawn admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Spawn admin darc")
	adminDarc, err := admcl.SpawnNewAdminDarc()
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = lib.GetDarcByID(admcl.bcl, gDarc.GetID()) // Verify that the darc is really in the global state
	require.NoError(t, err)

	// ------------------------------------------------------------------------
	// 2. Add a new admin to the admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Create new admin client (admcl2)")
	admcl2, err := NewClient(bcl)
	require.NoError(t, err)
	log.Lvl1("[INFO] Create deffered tx to add admcl2 identity")
	id, err := admcl.AddAdminToAdminDarc(adminDarc.GetBaseID(), admcl2.AuthKey().Identity())
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = admcl2.bcl.GetDeferredData(id) // Verify that the deferred transaction is registered on chain
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 1 sign the transaction")
	admcl.AddSignatureToDefferedTx(id, 0)
	log.Lvl1("[INFO] Admin 1 try to exec the transaction")
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)
	log.Lvl1("[INFO] Tx successfully executed, admin 2 has been added to the admin darc")

	// ------------------------------------------------------------------------
	// 2. Add a new admin to the admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Create new admin client (admcl3)")
	newAdmcl2, err := NewClient(bcl)
	require.NoError(t, err)
	log.Lvl1("[INFO] Create deffered tx to add admcl3 identity")
	id, err = admcl.ModifyAdminKeysFromAdminDarc(adminDarc.GetBaseID(), admcl2.AuthKey().Identity(), newAdmcl2.AuthKey().Identity())
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = admcl2.bcl.GetDeferredData(id) // Verify that the deferred transaction is registered on chain
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 1 sign the transaction")
	err = admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 2 sign the transaction")
	err = admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	log.Lvl1("[INFO] Admin 2 try to exec the transaction")
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)
	log.Lvl1("[INFO] Tx successfully executed, admin 3 has been added to the admin darc")
}

func TestAdminClient_SpawnProject(t *testing.T) {
	// ------------------------------------------------------------------------
	// 0. Set up
	// ------------------------------------------------------------------------
	log.Lvl1("[INFO] Setup")
	local := onet.NewTCPTest(cothority.Suite)
	defer local.CloseAll()

	superAdmin := darc.NewSignerEd25519(nil, nil)
	_, roster, _ := local.GenTree(3, true)

	genesisMsg, err := byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, roster,
		[]string{}, superAdmin.Identity())
	require.NoError(t, err)
	gDarc := &genesisMsg.GenesisDarc

	genesisMsg.BlockInterval = time.Second / 5
	bcl, _, err := byzcoin.NewLedger(genesisMsg, false)
	require.NoError(t, err)
	spawnNamingTx, err := bcl.CreateTransaction(byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(gDarc.GetBaseID()),
		Spawn: &byzcoin.Spawn{
			ContractID: byzcoin.ContractNamingID,
		},
		SignerCounter: []uint64{1},
	})
	require.NoError(t, spawnNamingTx.FillSignersAndSignWith(superAdmin))
	_, err = bcl.AddTransactionAndWait(spawnNamingTx, 10)
	require.NoError(t, err)

	log.Lvl1("[INFO] Create admin client")
	admcl, err := NewClientWithAuth(bcl, &superAdmin)
	admcl.incrementSignerCounter() // We increment the counter as the superadmin performed a transaction
	require.NoError(t, err)
	slc := []string{admcl.AuthKey().Identity().String()}
	admcl.SyncDarc(slc)

	// ------------------------------------------------------------------------
	// 1. Spawn admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Spawn admin darc")
	adminDarc, err := admcl.SpawnNewAdminDarc()
	require.NoError(t, err)

	// check if the admin darc is registered in the global state
	admcl.bcl.WaitPropagation(1)
	_, err = lib.GetDarcByID(admcl.bcl, adminDarc.GetBaseID())
	require.NoError(t, err)

	// ------------------------------------------------------------------------
	// 2. Create a new project named project A
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Spawn a new project darc")
	instID, pdarcID, err := admcl.CreateNewProject(adminDarc.GetBaseID(), "Project A")
	err = admcl.AddSignatureToDefferedTx(instID, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(instID)
	require.NoError(t, err)
	err = local.WaitDone(genesisMsg.BlockInterval)
	require.Nil(t, err)

	log.Lvl1("[INFO] Spawn a new accessright instance for the project")
	id, err := admcl.CreateAccessRight(pdarcID, adminDarc.GetBaseID())
	require.NoError(t, err)
	admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)

	err = local.WaitDone(genesisMsg.BlockInterval)
	require.Nil(t, err)

	// get the accessright contract instance id, to bind it to the project darc
	ddata, err := admcl.bcl.GetDeferredData(id)
	arID := ddata.ProposedTransaction.Instructions[0].DeriveID("")

	// bind the accessright instance to the project darc
	err = admcl.AttachAccessRightToProject(arID)
	require.NoError(t, err)

	id, err = admcl.bcl.ResolveInstanceID(pdarcID, "AR") // check that the access right value contract is correctly named
	require.Equal(t, id, arID)
	log.Lvl1("[INFO] The access right value contract is set")
}
func TestAdminClient_TestProjectWorkflow(t *testing.T) {
	// ------------------------------------------------------------------------
	// 0. Set up
	// ------------------------------------------------------------------------
	log.Lvl1("[INFO] Setup")
	local := onet.NewTCPTest(cothority.Suite)
	defer local.CloseAll()

	superAdmin := darc.NewSignerEd25519(nil, nil)
	_, roster, _ := local.GenTree(3, true)

	genesisMsg, err := byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, roster,
		[]string{}, superAdmin.Identity())
	require.NoError(t, err)
	gDarc := &genesisMsg.GenesisDarc

	genesisMsg.BlockInterval = time.Second / 5
	bcl, _, err := byzcoin.NewLedger(genesisMsg, false)
	require.NoError(t, err)
	spawnNamingTx, err := bcl.CreateTransaction(byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(gDarc.GetBaseID()),
		Spawn: &byzcoin.Spawn{
			ContractID: byzcoin.ContractNamingID,
		},
		SignerCounter: []uint64{1},
	})
	require.NoError(t, spawnNamingTx.FillSignersAndSignWith(superAdmin))
	_, err = bcl.AddTransactionAndWait(spawnNamingTx, 10)
	require.NoError(t, err)

	log.Lvl1("[INFO] Create admin client")
	admcl, err := NewClientWithAuth(bcl, &superAdmin)
	admcl.incrementSignerCounter() // We increment the counter as the superadmin keys performed a transaction before creating the clinet
	require.NoError(t, err)
	slc := []string{admcl.AuthKey().Identity().String()}
	admcl.SyncDarc(slc)

	// ------------------------------------------------------------------------
	// 1. Spawn admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Spawn admin darc")
	adminDarc, err := admcl.SpawnNewAdminDarc()
	adid := adminDarc.GetBaseID()
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	// verify the admin darc is registered in the global state
	_, err = lib.GetDarcByID(admcl.bcl, adid)
	require.NoError(t, err)

	log.Lvl1("[INFO] Create new admin client (admcl2)")
	admcl2, err := NewClient(bcl)
	require.NoError(t, err)

	log.Lvl1("[INFO] Create deffered tx to add admcl2 identity")
	id, err := admcl.AddAdminToAdminDarc(adid, admcl2.AuthKey().Identity())
	require.NoError(t, err)
	err = admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)

	err = local.WaitDone(genesisMsg.BlockInterval)
	require.Nil(t, err)
	slc = append(slc, admcl2.AuthKey().Identity().String())

	// TODO Change that to value of slc in adminDarc update the list of known admin into clients
	admcl.SyncDarc(slc)
	admcl2.SyncDarc(slc)

	// ------------------------------------------------------------------------
	// 2. Create a new project named project A
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Spawn a new project darc")
	instID, pdarcID, err := admcl.CreateNewProject(adid, "Project A")
	admcl.AddSignatureToDefferedTx(instID, 0)
	require.NoError(t, err)
	admcl2.AddSignatureToDefferedTx(instID, 0)
	require.NoError(t, err)
	admcl.ExecDefferedTx(instID)
	err = local.WaitDone(genesisMsg.BlockInterval)
	require.Nil(t, err)

	log.Lvl1("[INFO] Spawn a new accessright instance for the project")
	id, err = admcl.CreateAccessRight(pdarcID, adid)
	require.NoError(t, err)
	admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)

	err = local.WaitDone(genesisMsg.BlockInterval)
	require.Nil(t, err)

	// get the accessright contract instance id, to bind it to the project darc
	ddata, err := admcl.bcl.GetDeferredData(id)
	arID := ddata.ProposedTransaction.Instructions[0].DeriveID("")

	// bind the accessright instance to the project darc
	err = admcl.AttachAccessRightToProject(arID)
	require.NoError(t, err)
	id, err = admcl.bcl.ResolveInstanceID(pdarcID, "AR") // check that the access right value contract is correctly named
	require.Equal(t, id, arID)
	log.Lvl1("[INFO] The access right value contract is set")

	// ------------------------------------------------------------------------
	// 3. Interact with access rights for project A
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Add the querier 1:1 to the project")
	id, err = admcl.AddQuerierToProject(pdarcID, adid, "1:1", "count_per_site_shuffled,count_global")
	require.NoError(t, err)
	admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)
	log.Lvl1("[INFO] Modify the querier 1:1 access rights")
	id, err = admcl.ModifyQuerierAccessRightsForProject(pdarcID, adid, "1:1", "count_per_site_shuffled")
	require.NoError(t, err)
	admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)
	err = local.WaitDone(genesisMsg.BlockInterval)
	require.Nil(t, err)

	log.Lvl1("[INFO] Verify access right of querier 1:1 for action : count_per_site_shuffled")
	authorization, err := admcl.VerifyAccessRights("1:1", "count_per_site_shuffled", pdarcID)
	require.NoError(t, err)
	require.True(t, authorization)

	log.Lvl1("[INFO] Verify access right of querier 2:1 (not registered) for action :count_per_site_shuffled (Expected to fail)")
	authorization, err = admcl.VerifyAccessRights("2:1", "count_per_site_shuffled", pdarcID)
	require.Error(t, err)
	require.False(t, authorization)

	log.Lvl1("[INFO] Create new admin client (admcl3)")
	admcl3, err := NewClient(bcl)
	require.NoError(t, err)

	log.Lvl1("[INFO] Create deffered tx to add admcl3 identity")
	id, err = admcl.AddAdminToAdminDarc(adid, admcl3.AuthKey().Identity())
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = admcl2.bcl.GetDeferredData(id)
	require.NoError(t, err)

	log.Lvl1("[INFO] Admin 1 sign the transaction")
	err = admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)

	log.Lvl1("[INFO] Admin 2 sign the transaction")
	err = admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)

	log.Lvl1("[INFO] Admin 2 try to exec the transaction")
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)

	log.Lvl1("[INFO] Add the querier 3:1 to the project")
	id, err = admcl.AddQuerierToProject(pdarcID, adid, "3:1", "count_global")
	require.NoError(t, err)
	err = admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id) // This transaction is expected to fail as admin 3 is required to sign (for now multisignature rule is such that every admin sign)
	require.Error(t, err)
	err = admcl3.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)

	authorization, err = admcl.VerifyAccessRights("3:1", "count_per_site_shuffled", pdarcID)
	require.NoError(t, err)
	require.False(t, authorization)

	id, err = admcl.ModifyQuerierAccessRightsForProject(pdarcID, adid, "3:1", "count_per_site_shuffled")
	require.NoError(t, err)
	admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	admcl3.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)

	admcl.bcl.WaitPropagation(1)

	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)

	authorization, err = admcl.VerifyAccessRights("3:1", "count_per_site_shuffled", pdarcID)
	require.NoError(t, err)
	require.True(t, authorization)
}

func TestAdminClient_TestDummy(t *testing.T) {
	// log.SetDebugVisible(3)
	// ------------------------------------------------------------------------
	// 0. Set up
	// ------------------------------------------------------------------------
	log.Lvl1("[INFO] Setup")
	local := onet.NewTCPTest(cothority.Suite)
	defer local.CloseAll()

	superAdmin := darc.NewSignerEd25519(nil, nil)
	hosts, roster, _ := local.GenTree(3, true)

	genesisMsg, err := byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, roster,
		[]string{}, superAdmin.Identity())
	require.NoError(t, err)
	gDarc := &genesisMsg.GenesisDarc

	genesisMsg.BlockInterval = time.Second / 5
	bcl, _, err := byzcoin.NewLedger(genesisMsg, false)
	require.NoError(t, err)
	spawnNamingTx, err := bcl.CreateTransaction(byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(gDarc.GetBaseID()),
		Spawn: &byzcoin.Spawn{
			ContractID: byzcoin.ContractNamingID,
		},
		SignerCounter: []uint64{1},
	})
	require.NoError(t, spawnNamingTx.FillSignersAndSignWith(superAdmin))
	_, err = bcl.AddTransactionAndWait(spawnNamingTx, 10)
	require.NoError(t, err)

	log.Lvl1("[INFO] Create admin client")
	admcl, err := NewClientWithAuth(bcl, &superAdmin)
	admcl.incrementSignerCounter() // We increment the counter as the superadmin keys performed a transaction before creating the clinet
	require.NoError(t, err)
	// err = admcl.DummyTest()
	// services := local.GetServices(hosts, templateID)
	// log.Lvl1("service lenght ", services)
	// for _, s := range services {
	// 	resp, err := s.(*Service).HandleShareID(
	// 		&ShareID{"TEST"},
	// 	)
	// 	require.NoError(t, err)
	// 	fmt.Println(resp)
	// }
	services := local.GetServices(hosts, adminID)
	for _, s := range services {
		log.Lvl2("Sending request to", s)
		resp, err := s.(*Service).HandleGetDefferedIDs(
			&GetDeferredIDs{},
		)
		require.Nil(t, err)
		fmt.Println(resp)
	}

	type Client struct {
		*onet.Client
	}

	oooo := Client{Client: onet.NewClient(cothority.Suite, "ShareID")}

	err = local.WaitDone(genesisMsg.BlockInterval)
	require.Nil(t, err)
	ttt := DefferedIDReply{}
	err = oooo.SendProtobuf(roster.List[1], &DefferedID{spawnNamingTx.Instructions[0].DeriveID(""), roster}, &ttt)
	require.NoError(t, err)
	err = oooo.SendProtobuf(roster.List[1], &DefferedID{byzcoin.InstanceID{}, roster}, &ttt)
	require.NoError(t, err)
	services = local.GetServices(hosts, adminID)
	for _, s := range services {
		log.Lvl2("Sending request to", s)
		resp, err := s.(*Service).HandleGetDefferedIDs(
			&GetDeferredIDs{},
		)
		require.Nil(t, err)
		fmt.Println(resp)
	}
}
func TestAdminClient_TestProjectWorkflowWithShare(t *testing.T) {
	// ------------------------------------------------------------------------
	// 0. Set up
	// ------------------------------------------------------------------------
	log.Lvl1("[INFO] Setup")
	local := onet.NewTCPTest(cothority.Suite)
	defer local.CloseAll()

	superAdmin := darc.NewSignerEd25519(nil, nil)
	_, roster, _ := local.GenTree(3, true)

	genesisMsg, err := byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, roster,
		[]string{}, superAdmin.Identity())
	require.NoError(t, err)
	gDarc := &genesisMsg.GenesisDarc

	genesisMsg.BlockInterval = time.Second / 5
	bcl, _, err := byzcoin.NewLedger(genesisMsg, false)
	require.NoError(t, err)
	spawnNamingTx, err := bcl.CreateTransaction(byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(gDarc.GetBaseID()),
		Spawn: &byzcoin.Spawn{
			ContractID: byzcoin.ContractNamingID,
		},
		SignerCounter: []uint64{1},
	})
	require.NoError(t, spawnNamingTx.FillSignersAndSignWith(superAdmin))
	_, err = bcl.AddTransactionAndWait(spawnNamingTx, 10)
	require.NoError(t, err)

	log.Lvl1("[INFO] Create admin client")
	admcl, err := NewClientWithAuth(bcl, &superAdmin)
	admcl.incrementSignerCounter() // We increment the counter as the superadmin keys performed a transaction before creating the clinet
	require.NoError(t, err)
	slc := []string{admcl.AuthKey().Identity().String()}
	admcl.SyncDarc(slc)

	// ------------------------------------------------------------------------
	// 1. Spawn admin darc
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Spawn admin darc")
	adminDarc, err := admcl.SpawnNewAdminDarc()
	adid := adminDarc.GetBaseID()
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	// verify the admin darc is registered in the global state
	_, err = lib.GetDarcByID(admcl.bcl, adid)
	require.NoError(t, err)

	log.Lvl1("[INFO] Create new admin client (admcl2)")
	admcl2, err := NewClient(bcl)
	require.NoError(t, err)

	log.Lvl1("[INFO] Create deffered tx to add admcl2 identity")
	id, err := admcl.AddAdminToAdminDarc(adid, admcl2.AuthKey().Identity())
	require.NoError(t, err)
	err = admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)

	err = local.WaitDone(genesisMsg.BlockInterval)
	require.Nil(t, err)
	slc = append(slc, admcl2.AuthKey().Identity().String())

	// TODO Change that to value of slc in adminDarc update the list of known admin into clients
	admcl.SyncDarc(slc)
	admcl2.SyncDarc(slc)

	// ------------------------------------------------------------------------
	// 2. Create a new project named project A
	// ------------------------------------------------------------------------

	type Client struct {
		*onet.Client
	}

	log.Lvl1("[INFO] Spawn a new project darc")
	instID, pdarcID, err := admcl.CreateNewProject(adid, "Project A")
	require.NoError(t, err)
	admcl.AddSignatureToDefferedTx(instID, 0)
	time.Sleep(time.Second)
	require.NoError(t, err)
	res, err := admcl2.FetchNewDefferedInstanceIDs()
	fmt.Println(res)
	require.NoError(t, err)
	fmt.Println(res)
	instID = res.Ids[1]
	admcl2.AddSignatureToDefferedTx(instID, 0)
	require.NoError(t, err)
	admcl.ExecDefferedTx(instID)
	err = local.WaitDone(genesisMsg.BlockInterval)
	require.Nil(t, err)

	log.Lvl1("[INFO] Spawn a new accessright instance for the project")
	id, err = admcl.CreateAccessRight(pdarcID, adid)
	require.NoError(t, err)
	err = admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)

	err = local.WaitDone(genesisMsg.BlockInterval)
	require.Nil(t, err)

	// get the accessright contract instance id, to bind it to the project darc
	ddata, err := admcl.bcl.GetDeferredData(id)
	arID := ddata.ProposedTransaction.Instructions[0].DeriveID("")

	// bind the accessright instance to the project darc
	err = admcl.AttachAccessRightToProject(arID)
	require.NoError(t, err)
	id, err = admcl.bcl.ResolveInstanceID(pdarcID, "AR") // check that the access right value contract is correctly named
	require.Equal(t, id, arID)
	log.Lvl1("[INFO] The access right value contract is set")

	// ------------------------------------------------------------------------
	// 3. Interact with access rights for project A
	// ------------------------------------------------------------------------

	log.Lvl1("[INFO] Add the querier 1:1 to the project")
	id, err = admcl.AddQuerierToProject(pdarcID, adid, "1:1", "count_per_site_shuffled,count_global")
	require.NoError(t, err)
	admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)
	log.Lvl1("[INFO] Modify the querier 1:1 access rights")
	id, err = admcl.ModifyQuerierAccessRightsForProject(pdarcID, adid, "1:1", "count_per_site_shuffled")
	require.NoError(t, err)
	admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)
	err = local.WaitDone(genesisMsg.BlockInterval)
	require.Nil(t, err)

	log.Lvl1("[INFO] Verify access right of querier 1:1 for action : count_per_site_shuffled")
	authorization, err := admcl.VerifyAccessRights("1:1", "count_per_site_shuffled", pdarcID)
	require.NoError(t, err)
	require.True(t, authorization)

	log.Lvl1("[INFO] Verify access right of querier 2:1 (not registered) for action :count_per_site_shuffled (Expected to fail)")
	authorization, err = admcl.VerifyAccessRights("2:1", "count_per_site_shuffled", pdarcID)
	require.Error(t, err)
	require.False(t, authorization)

	log.Lvl1("[INFO] Create new admin client (admcl3)")
	admcl3, err := NewClient(bcl)
	require.NoError(t, err)

	log.Lvl1("[INFO] Create deffered tx to add admcl3 identity")
	id, err = admcl.AddAdminToAdminDarc(adid, admcl3.AuthKey().Identity())
	require.NoError(t, err)
	admcl.bcl.WaitPropagation(1)
	_, err = admcl2.bcl.GetDeferredData(id)
	require.NoError(t, err)

	log.Lvl1("[INFO] Admin 1 sign the transaction")
	err = admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)

	log.Lvl1("[INFO] Admin 2 sign the transaction")
	err = admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)

	log.Lvl1("[INFO] Admin 2 try to exec the transaction")
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)

	log.Lvl1("[INFO] Add the querier 3:1 to the project")
	id, err = admcl.AddQuerierToProject(pdarcID, adid, "3:1", "count_global")
	require.NoError(t, err)
	err = admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id) // This transaction is expected to fail as admin 3 is required to sign (for now multisignature rule is such that every admin sign)
	require.Error(t, err)
	err = admcl3.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)

	authorization, err = admcl.VerifyAccessRights("3:1", "count_per_site_shuffled", pdarcID)
	require.NoError(t, err)
	require.False(t, authorization)

	id, err = admcl.ModifyQuerierAccessRightsForProject(pdarcID, adid, "3:1", "count_per_site_shuffled")
	require.NoError(t, err)
	admcl.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	admcl2.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)
	admcl3.AddSignatureToDefferedTx(id, 0)
	require.NoError(t, err)

	admcl.bcl.WaitPropagation(1)

	err = admcl.ExecDefferedTx(id)
	require.NoError(t, err)

	authorization, err = admcl.VerifyAccessRights("3:1", "count_per_site_shuffled", pdarcID)
	require.NoError(t, err)
	require.True(t, authorization)
}
