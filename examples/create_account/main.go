/*
 * Flow Go SDK
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/examples"
	"github.com/onflow/flow-go-sdk/templates"
)

func main() {
	CreateAccountDemo()
}

func CreateAccountDemo() {
	ctx := context.Background()
	flowClient, err := client.New("127.0.0.1:3569", grpc.WithInsecure())
	examples.Handle(err)

	rootAcctAddr, rootAcctKey, rootSigner := examples.RootAccount(flowClient)

	myPrivateKey := examples.RandomPrivateKey()
	myAcctKey := flow.NewAccountKey().
		FromPrivateKey(myPrivateKey).
		SetHashAlgo(crypto.SHA3_256).
		SetWeight(flow.AccountKeyWeightThreshold)

	// Create a Cadence script which will create an account with one key with weight 1 and
	createAccountScript, err := templates.CreateAccount([]*flow.AccountKey{myAcctKey}, nil)
	examples.Handle(err)

	// Create a transaction that will execute the script. The transaction is signed
	// by the root account.
	createAccountTx := flow.NewTransaction().
		SetScript(createAccountScript).
		SetProposalKey(rootAcctAddr, rootAcctKey.ID, rootAcctKey.SequenceNumber).
		SetPayer(rootAcctAddr)

	// Sign the transaction with the root account, which already exists
	// All new accounts must be created by an existing account
	err = createAccountTx.SignEnvelope(rootAcctAddr, rootAcctKey.ID, rootSigner)
	examples.Handle(err)

	// Send the transaction to the network
	err = flowClient.SendTransaction(ctx, *createAccountTx)
	examples.Handle(err)

	accountCreationTxRes := examples.WaitForSeal(ctx, flowClient, createAccountTx.ID())

	var myAddress flow.Address

	for _, event := range accountCreationTxRes.Events {
		if event.Type == flow.EventAccountCreated {
			accountCreatedEvent := flow.AccountCreatedEvent(event)
			myAddress = accountCreatedEvent.Address()
		}
	}

	fmt.Println("Account created with address:", myAddress.Hex())
}
