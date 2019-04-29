/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package run

// Agent operations dispatcher

import (
	"context"
	"sync"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-inventory-manager-go"

	"github.com/nalej/service-net-agent/internal/pkg/client"
	"github.com/nalej/service-net-agent/internal/pkg/plugin"

	"github.com/rs/zerolog/log"
)

type Dispatcher struct {
	client *client.AgentClient

	opQueue chan *grpc_inventory_manager_go.AgentOpRequest
	resQueue chan *grpc_inventory_manager_go.AgentOpResponse

	wg sync.WaitGroup
	cancel context.CancelFunc
}

func NewDispatcher(client *client.AgentClient, queueLen int) (*Dispatcher, derrors.Error) {
	// We want to be able to cancel
	ctx, cancel := context.WithCancel(context.Background())

	d := &Dispatcher{
		client: client,
		opQueue: make(chan *grpc_inventory_manager_go.AgentOpRequest, queueLen),
		resQueue: make(chan *grpc_inventory_manager_go.AgentOpResponse, queueLen),
		cancel: cancel,
	}

	// Start response routine
	d.wg.Add(1)
	// We stop the sending of responses by closing the channel, this allows
	// us to still send all responses that are queued so the edge
	// controller is aware of their status.
	go d.resWorker(context.Background())

	// Start worker routine
	d.wg.Add(1)
	go d.opWorker(ctx)

	// go work(...)?
	// - new worker
	// - read queue
	// - translate
	// - send to worker
	// TBD


	return d, nil
}

func (d *Dispatcher) Stop(timeout time.Duration) (derrors.Error) {
	log.Debug().Msg("stopping operation dispatcher")

	// We cancel the operation worker routine - this potentially
	// finishes the in-progress operation if it doesn't use the
	// context properly - which is ok, we'll have a timeout later.
	d.cancel()

	// We loop over in-progress operation requests and tell the edge
	// controller they are cancelled.
	// We close the operations channel so our loop ends.
	close(d.opQueue)
	for op := range (d.opQueue) {
		status := grpc_inventory_manager_go.AgentOpStatus_FAIL
		info := "agent stopped"
		d.respond(op, status, info)
	}

	// We indicate that we won't send any more responses. Response
	// worker will send queued ones and then quit.
	close(d.resQueue)

	// We wait until all is finished - with a timeout
	// Wait in a separate routine so we can timeout
	ch := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(ch)
	}()

	// Check whether wait was succesfull or we timed out
	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
		return derrors.NewDeadlineExceededError("waiting for operation workers timed out").WithParams(timeout)
	}

	return nil
}

// Add operation to queue and notify Edge Controller
// We don't return an error unless something is really broken. Under normal
// operation we can return an error to Edge Controller.
func (d *Dispatcher) Dispatch(op *grpc_inventory_manager_go.AgentOpRequest) derrors.Error {
	// Make dispatching non-blocking
	var status grpc_inventory_manager_go.AgentOpStatus
	var info string
	select {
	case d.opQueue <- op:
		// log.Debug().Str("operation_id", op.GetOperationId()).Msg("queued operation")
		status = grpc_inventory_manager_go.AgentOpStatus_SCHEDULED
	default:
		// log.Debug().Str("operation_id", op.GetOperationId()).Msg("operation queue full")
		status = grpc_inventory_manager_go.AgentOpStatus_FAIL
		info = "agent operation queue full"
	}

	// Add to response queue
	d.respond(op, status, info)

	return nil
}

func (d *Dispatcher) respond(op *grpc_inventory_manager_go.AgentOpRequest, status grpc_inventory_manager_go.AgentOpStatus, info string) {
	response := &grpc_inventory_manager_go.AgentOpResponse{
		OrganizationId: op.GetOrganizationId(),
		EdgeControllerId: op.GetEdgeControllerId(),
		AssetId: op.GetAssetId(),
		// OperationId: op.GetOperationId(),
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
	// log.Debug().Str("operation_id", op.GetOperationId()).Msg("queued operation response")
}

func (d *Dispatcher) resWorker(ctx context.Context) {
	defer d.wg.Done()

	log.Debug().Msg("starting operation response worker")
	for ctx.Err() == nil && d.resQueue != nil {
		select {
		case response, ok := <-d.resQueue:
			if !ok {
				// Channel closed and end of queue
				d.resQueue = nil
				break
			}
			// log.Debug().Str("operation_id", response.GetOperationId()).Msg("sending operation response")
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
	defer d.wg.Done()

	log.Debug().Msg("starting operation worker")

	// Create worker
	worker := NewWorker()

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
			var opId = "" // op.GetOperationId()

			log.Debug().
				Str("operation_id", opId).
				Str("plugin", pluginName.String()).
				Str("operation", opName.String()).
				Interface("params", params).
				Msg("executing operation request")

			status := grpc_inventory_manager_go.AgentOpStatus_SUCCESS
			result, derr := worker.Execute(ctx, pluginName, opName, params)
			if derr != nil {
				log.Warn().Err(derr).Str("operation_id", opId).Msg("failed executing operation")
				status = grpc_inventory_manager_go.AgentOpStatus_FAIL
				result = derr.Error()
			}
			d.respond(op, status, result)
		case <-ctx.Done():
			break
		}
	}

	log.Debug().Msg("operation worker stopped")
}
