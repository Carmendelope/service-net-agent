/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package run

// Agent operations dispatcher

import (
	"context"
	grpc_inventory_go "github.com/nalej/grpc-inventory-go"
	"sync"
	"time"

	"github.com/nalej/derrors"
	"github.com/nalej/infra-net-plugin"

	"github.com/nalej/grpc-inventory-manager-go"

	"github.com/nalej/service-net-agent/internal/pkg/client"

	"github.com/rs/zerolog/log"
)

type Dispatcher struct {
	// Client connection to Edge Controller
	client *client.AgentClient
	worker *Worker

	// Queue of operations received from Edge Controller to be processed
	opQueue chan *grpc_inventory_manager_go.AgentOpRequest
	// Queue of responses to be sent to Edge Controller
	resQueue chan *grpc_inventory_manager_go.AgentOpResponse

	// Wait for working handling operations queue to finish
	opWorkerWaitgroup sync.WaitGroup
	// Wait for working handling response queue to finish
	resWorkerWaitgroup sync.WaitGroup

	// Cancel function to interrupt operations worker
	cancelOpWorker context.CancelFunc
}

func NewDispatcher(client *client.AgentClient, worker *Worker, queueLen int) (*Dispatcher, derrors.Error) {
	// We want to be able to cancel
	ctx, cancelOpWorker := context.WithCancel(context.Background())

	d := &Dispatcher{
		client: client,
		worker: worker,
		opQueue: make(chan *grpc_inventory_manager_go.AgentOpRequest, queueLen),
		resQueue: make(chan *grpc_inventory_manager_go.AgentOpResponse, queueLen),
		cancelOpWorker: cancelOpWorker,
	}

	// Start response routine
	d.resWorkerWaitgroup.Add(1)
	// We stop the sending of responses by closing the channel, this allows
	// us to still send all responses that are queued so the edge
	// controller is aware of their status.
	go d.resWorker(context.Background())

	// Start worker routine
	d.opWorkerWaitgroup.Add(1)
	go d.opWorker(ctx)

	return d, nil
}

func (d *Dispatcher) Stop(timeout time.Duration) (derrors.Error) {
	log.Debug().Msg("stopping operation dispatcher")

	// Set timeout for complete shutdown routine
	timeoutChan := time.After(timeout)

	// We cancel the operation worker routine - this potentially
	// finishes the in-progress operation if it doesn't use the
	// context properly - which is ok, we have a timeout.
	d.cancelOpWorker()

	// We loop over queued operation requests and tell the edge
	// controller they are cancelled.
	// We close the operations channel so our loop ends.
	// We do this while the last operation is in progress so we have
	// more time to send these out (in parallel with operation).
	close(d.opQueue)
	for op := range (d.opQueue) {
		status := grpc_inventory_go.OpStatus_FAIL
		info := "agent stopped"
		d.respond(op, status, info)
	}

	// Wait for operation routine
	timedout := wait(&d.opWorkerWaitgroup, timeoutChan)
	if timedout {
		return derrors.NewDeadlineExceededError("waiting for operation workers timed out - in-flight operations taking too long").WithParams(timeout)
	}

	// We indicate that we won't send any more responses. Response
	// worker will send queued ones and then quit.
	// We couldn't close earlier and wait together with operation routine,
	// because operation routine could have sent one last response.
	close(d.resQueue)

	// Wait for response worker with same timeout
	timedout = wait(&d.resWorkerWaitgroup, timeoutChan)
	if timedout {
		return derrors.NewDeadlineExceededError("waiting for response workers timed out - communication taking too long").WithParams(timeout)
	}

	return nil
}

func wait(wg *sync.WaitGroup, timeoutChan <-chan time.Time) bool {
	// We wait until all is finished - with a timeout
	// Wait in a separate routine so we can timeout
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Check whether wait was succesfull or we timed out
	select {
	case <-ch:
		break
	case <-timeoutChan:
		return true
	}

	return false // No timeout
}

// Add operation to queue and notify Edge Controller
// We don't return an error unless something is really broken. Under normal
// operation we can return an error to Edge Controller.
func (d *Dispatcher) Dispatch(op *grpc_inventory_manager_go.AgentOpRequest) derrors.Error {
	// Make dispatching non-blocking
	var status grpc_inventory_go.OpStatus
	var info string
	select {
	case d.opQueue <- op:
		log.Debug().Str("operation_id", op.GetOperationId()).Msg("queued operation")
		status = grpc_inventory_go.OpStatus_SCHEDULED
	default:
		log.Debug().Str("operation_id", op.GetOperationId()).Msg("operation queue full")
		status = grpc_inventory_go.OpStatus_FAIL
		info = "agent operation queue full"
	}

	// Add to response queue
	d.respond(op, status, info)

	return nil
}

func (d *Dispatcher) respond(op *grpc_inventory_manager_go.AgentOpRequest, status grpc_inventory_go.OpStatus, info string) {
	response := &grpc_inventory_manager_go.AgentOpResponse{
		OrganizationId: op.GetOrganizationId(),
		EdgeControllerId: op.GetEdgeControllerId(),
		AssetId: op.GetAssetId(),
		OperationId: op.GetOperationId(),
		Timestamp: time.Now().UTC().Unix(),
		Status: status,
		Info: info,
	}

	// We want to response in order, while not blocking main loop; hence
	// we have a response queue and worker. If the response queue is full
	// it means communication with Edge Controller is really slow and it's
	// ok to block main loop - agent might die but something is wrong anyway
	// so an agent restart might help.
	d.resQueue <- response
	log.Debug().Str("operation_id", op.GetOperationId()).Msg("queued operation response")
}

func (d *Dispatcher) resWorker(ctx context.Context) {
	defer d.resWorkerWaitgroup.Done()

	log.Debug().Msg("starting operation response worker")
	for ctx.Err() == nil && d.resQueue != nil {
		select {
		case response, ok := <-d.resQueue:
			if !ok {
				// Channel closed and end of queue
				d.resQueue = nil
				break
			}
			log.Debug().Str("operation_id", response.GetOperationId()).Msg("sending operation response")
			_, err := d.client.CallbackAgentOperation(d.client.GetContext(), response)
			if err != nil {
				log.Warn().Err(err).Msg("failed sending operation response to edge controller")
			}
		case <-ctx.Done():
			break
		}
	}

	log.Debug().Msg("operation response worker stopped")
}

func (d *Dispatcher) opWorker(ctx context.Context) {
	defer d.opWorkerWaitgroup.Done()

	log.Debug().Msg("starting operation worker")

	for ctx.Err() == nil && d.opQueue != nil {
		select {
		case op, ok := <-d.opQueue:
			if !ok {
				// Channel closed and end of queue
				d.opQueue = nil
				break
			}

			var pluginName = plugin.PluginName(op.GetPlugin())
			var opName = plugin.CommandName(op.GetOperation())
			var params = op.GetParams()
			var opId = op.GetOperationId()

			log.Debug().
				Str("operation_id", opId).
				Str("plugin", pluginName.String()).
				Str("operation", opName.String()).
				Interface("params", params).
				Msg("executing operation request")

			status := grpc_inventory_go.OpStatus_SUCCESS
			result, derr := d.worker.Execute(ctx, pluginName, opName, params)
			if derr != nil {
				log.Warn().Err(derr).Str("operation_id", opId).Msg("failed executing operation")
				status = grpc_inventory_go.OpStatus_FAIL
				result = derr.Error()
			}
			d.respond(op, status, result)
		case <-ctx.Done():
			break
		}
	}

	log.Debug().Msg("operation worker stopped")
}
