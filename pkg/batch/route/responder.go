/*
 *    Copyright 2019 Samsung SDS
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package route

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"

	"github.com/nexledger/accelerator/pkg/batch/route/encoding"
	"github.com/nexledger/accelerator/pkg/batch/tx"
)

type responder struct {
	encoder encoding.Encoder
}

func (r *responder) ItemSuccess(notifier chan *tx.Result, payload []byte, resp *channel.Response) {
	notifier <- &tx.Result{
		TxId:            string(resp.TransactionID),
		ValidationCode:  int32(resp.TxValidationCode),
		ChaincodeStatus: resp.ChaincodeStatus,
		Payload:         payload,
	}
}

func (r *responder) ItemFailure(notifier chan *tx.Result, err error) {
	notifier <- &tx.Result{Error: err}
}

func (r *responder) JobSuccess(job *tx.Job, fabresp *channel.Response) {
	results, err := r.encoder.DecodeResponse(fabresp.Payload)
	if err != nil {
		r.JobFailure(job, err)
		return
	} else if job.Size() != len(results) {
		r.JobFailure(job, errors.New("response length mismatch"))
		return
	}

	for i, result := range results {
		r.ItemSuccess(job.Notifiers()[i], result, fabresp)
	}
}

func (r *responder) JobFailure(job *tx.Job, err error) {
	for _, notifier := range job.Notifiers() {
		r.ItemFailure(notifier, err)
	}
}