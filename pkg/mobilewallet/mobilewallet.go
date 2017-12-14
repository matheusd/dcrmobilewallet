package mobilewallet

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/hdkeychain"
	"github.com/decred/dcrd/txscript"
	"github.com/decred/dcrd/wire"
	"github.com/decred/dcrwallet/chain"
	"github.com/decred/dcrwallet/loader"
	"github.com/decred/dcrwallet/netparams"
	"github.com/decred/dcrwallet/wallet"
	"github.com/decred/dcrwallet/wallet/txrules"
)

type LibWallet struct {
	dbDir      string
	wallet     *wallet.Wallet
	rpcClient  *chain.RPCClient
	loader     *loader.Loader
	syncer     *chain.RPCSyncer
	netBackend wallet.NetworkBackend
}

func NewLibWallet(dbDir string) *LibWallet {
	lw := &LibWallet{
		dbDir: dbDir,
	}
	return lw
}

func (lw *LibWallet) CreateWallet() error {
	stakeOptions := &loader.StakeOptions{
		VotingEnabled: false,
		AddressReuse:  false,
		VotingAddress: nil,
		TicketFee:     10e8,
	}

	pubPass := []byte(wallet.InsecurePubPassphrase)
	privPass := []byte("123")
	seed, err := hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
	if err != nil {
		return err
	}

	loader := loader.NewLoader(netparams.TestNet2Params.Params, lw.dbDir, stakeOptions,
		20, false, 10e5)

	w, err := loader.CreateNewWallet(pubPass, privPass, seed)
	if err != nil {
		return err
	}

	err = w.UpgradeToSLIP0044CoinType()
	if err != nil {
		return err
	}
	return loader.UnloadWallet()
}

func (lw *LibWallet) OpenWallet() error {
	stakeOptions := &loader.StakeOptions{
		VotingEnabled: false,
		AddressReuse:  false,
		VotingAddress: nil,
		TicketFee:     10e8,
	}

	loader := loader.NewLoader(netparams.TestNet2Params.Params, lw.dbDir, stakeOptions,
		20, false, 10e5)

	pubPass := []byte(wallet.InsecurePubPassphrase)
	w, err := loader.OpenExistingWallet(pubPass)

	if err != nil {
		return err
	}
	lw.wallet = w

	certs := []byte(CACert)
	disableTls := false

	ctx := context.Background()

	c, err := chain.NewRPCClient(netparams.TestNet2Params.Params, rpcHost,
		rpcUser, rpcPass, certs, disableTls)
	if err != nil {
		return err
	}
	err = c.Start(ctx, false)
	if err != nil {
		return err
	}

	lw.netBackend = chain.BackendFromRPCClient(c.Client)
	lw.rpcClient = c
	lw.wallet.SetNetworkBackend(lw.netBackend)

	syncer := chain.NewRPCSyncer(lw.wallet, c)
	lw.syncer = syncer
	go syncer.Run(ctx, true) // TODO: separate to other func

	err = c.NotifyBlocks()
	if err != nil {
		return err
	}

	// err = w.DiscoverActiveAddresses(lw.netBackend, true)
	// if err != nil {
	// 	return err
	// }

	err = w.LoadActiveDataFilters(lw.netBackend)
	if err != nil {
		return err
	}

	// lets skip loading all headers for now...
	// _, _, _, _, _, err = w.FetchHeaders(lw.netBackend)
	// if err != nil {
	// 	return err
	// }

	// no need to rescan on this test.
	// if fetchedHeaderCount > 0 {
	// 	err = w.Rescan(ctx, lw.netBackend, &rescanStart)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (lw *LibWallet) Rescan() error {
	ctx := context.Background()
	n, _ := lw.wallet.NetworkBackend()
	return lw.wallet.RescanFromHeight(ctx, n, 180000) // no one was using dcrmobilewallet before block 180000, right?
}

func (lw *LibWallet) SpendableForAccount() (int64, error) {
	var account uint32 = 0
	var reqConfs int32 = 1

	bals, err := lw.wallet.CalculateAccountBalance(account, reqConfs)
	if err != nil {
		return 0, err
	}

	return int64(bals.Spendable), nil
}

func (lw *LibWallet) AddressForAccount() (string, error) {
	var account uint32 = 0

	var callOpts []wallet.NextAddressCallOption
	callOpts = append(callOpts, wallet.WithGapPolicyWrap())

	addr, err := lw.wallet.NewExternalAddress(account, callOpts...)
	if err != nil {
		return "", err
	}

	return addr.EncodeAddress(), nil
}

func (lw *LibWallet) SendTx() (string, error) {

	var amount int64 = 10e7 // 1 DCR
	var srcAccount uint32 = 0
	var RequiredConfirmations int32 = 0

	// output destination
	addr, err := dcrutil.DecodeAddress(destAddr)
	if err != nil {
		return "", fmt.Errorf("invalid address %v: %v", destAddr, err)
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
	}
	version := txscript.DefaultScriptVersion

	// pay output
	outputs := make([]*wire.TxOut, 1)
	outputs[0] = &wire.TxOut{
		Value:    amount,
		Version:  version,
		PkScript: pkScript,
	}
	var algo wallet.OutputSelectionAlgorithm = wallet.OutputSelectionAlgorithmDefault
	feePerKb := txrules.DefaultRelayFeePerKb

	// create tx
	tx, err := lw.wallet.NewUnsignedTransaction(outputs, feePerKb, srcAccount,
		RequiredConfirmations, algo, nil)
	if err != nil {
		return "", err
	}

	// unlock wallet
	lock := make(chan time.Time, 1)
	defer func() {
		lock <- time.Time{} // send matters, not the value
	}()
	err = lw.wallet.Unlock([]byte(passphrase), lock)
	if err != nil {
		return "", err
	}

	// sign tx
	_, err = lw.wallet.SignTransaction(tx.Tx,
		txscript.SigHashAll, nil, nil, nil)
	if err != nil {
		return "", err
	}

	// serialize to send
	var serializedTransaction bytes.Buffer
	serializedTransaction.Grow(tx.Tx.SerializeSize())
	err = tx.Tx.Serialize(&serializedTransaction)
	if err != nil {
		return "", err
	}

	// publish tx
	txHash, err := lw.wallet.PublishTransaction(tx.Tx, serializedTransaction.Bytes(), lw.netBackend)
	if err != nil {
		return "", err
	}

	return txHash.String(), nil
}
