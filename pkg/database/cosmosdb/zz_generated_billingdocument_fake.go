// Code generated by github.com/jim-minter/go-cosmosdb, DO NOT EDIT.

package cosmosdb

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/ugorji/go/codec"

	pkg "github.com/Azure/ARO-RP/pkg/api"
)

type fakeBillingDocumentTriggerHandler func(context.Context, *pkg.BillingDocument) error
type fakeBillingDocumentQueryHandler func(BillingDocumentClient, *Query, *Options) BillingDocumentRawIterator

var _ BillingDocumentClient = &FakeBillingDocumentClient{}

// NewFakeBillingDocumentClient returns a FakeBillingDocumentClient
func NewFakeBillingDocumentClient(h *codec.JsonHandle) *FakeBillingDocumentClient {
	return &FakeBillingDocumentClient{
		billingDocuments: make(map[string][]byte),
		triggerHandlers:  make(map[string]fakeBillingDocumentTriggerHandler),
		queryHandlers:    make(map[string]fakeBillingDocumentQueryHandler),
		jsonHandle:       h,
		lock:             &sync.RWMutex{},
	}
}

// FakeBillingDocumentClient is a FakeBillingDocumentClient
type FakeBillingDocumentClient struct {
	billingDocuments map[string][]byte
	jsonHandle       *codec.JsonHandle
	lock             *sync.RWMutex
	triggerHandlers  map[string]fakeBillingDocumentTriggerHandler
	queryHandlers    map[string]fakeBillingDocumentQueryHandler
	sorter           func([]*pkg.BillingDocument)

	// returns true if documents conflict
	conflictChecker func(*pkg.BillingDocument, *pkg.BillingDocument) bool

	// err, if not nil, is an error to return when attempting to communicate
	// with this Client
	err error
}

func (c *FakeBillingDocumentClient) decodeBillingDocument(s []byte) (billingDocument *pkg.BillingDocument, err error) {
	err = codec.NewDecoderBytes(s, c.jsonHandle).Decode(&billingDocument)
	return
}

func (c *FakeBillingDocumentClient) encodeBillingDocument(billingDocument *pkg.BillingDocument) (b []byte, err error) {
	err = codec.NewEncoderBytes(&b, c.jsonHandle).Encode(billingDocument)
	return
}

// SetError sets or unsets an error that will be returned on any
// FakeBillingDocumentClient method invocation
func (c *FakeBillingDocumentClient) SetError(err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.err = err
}

// SetSorter sets or unsets a sorter function which will be used to sort values
// returned by List() for test stability
func (c *FakeBillingDocumentClient) SetSorter(sorter func([]*pkg.BillingDocument)) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.sorter = sorter
}

// SetConflictChecker sets or unsets a function which can be used to validate
// additional unique keys in a BillingDocument
func (c *FakeBillingDocumentClient) SetConflictChecker(conflictChecker func(*pkg.BillingDocument, *pkg.BillingDocument) bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.conflictChecker = conflictChecker
}

// SetTriggerHandler sets or unsets a trigger handler
func (c *FakeBillingDocumentClient) SetTriggerHandler(triggerName string, trigger fakeBillingDocumentTriggerHandler) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.triggerHandlers[triggerName] = trigger
}

// SetQueryHandler sets or unsets a query handler
func (c *FakeBillingDocumentClient) SetQueryHandler(queryName string, query fakeBillingDocumentQueryHandler) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.queryHandlers[queryName] = query
}

func (c *FakeBillingDocumentClient) deepCopy(billingDocument *pkg.BillingDocument) (*pkg.BillingDocument, error) {
	b, err := c.encodeBillingDocument(billingDocument)
	if err != nil {
		return nil, err
	}

	return c.decodeBillingDocument(b)
}

func (c *FakeBillingDocumentClient) apply(ctx context.Context, partitionkey string, billingDocument *pkg.BillingDocument, options *Options, isCreate bool) (*pkg.BillingDocument, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.err != nil {
		return nil, c.err
	}

	billingDocument, err := c.deepCopy(billingDocument) // copy now because pretriggers can mutate billingDocument
	if err != nil {
		return nil, err
	}

	if options != nil {
		err := c.processPreTriggers(ctx, billingDocument, options)
		if err != nil {
			return nil, err
		}
	}

	_, exists := c.billingDocuments[billingDocument.ID]
	if isCreate && exists {
		return nil, &Error{
			StatusCode: http.StatusConflict,
			Message:    "Entity with the specified id already exists in the system",
		}
	}
	if !isCreate && !exists {
		return nil, &Error{StatusCode: http.StatusNotFound}
	}

	if c.conflictChecker != nil {
		for id := range c.billingDocuments {
			billingDocumentToCheck, err := c.decodeBillingDocument(c.billingDocuments[id])
			if err != nil {
				return nil, err
			}

			if c.conflictChecker(billingDocumentToCheck, billingDocument) {
				return nil, &Error{
					StatusCode: http.StatusConflict,
					Message:    "Entity with the specified id already exists in the system",
				}
			}
		}
	}

	b, err := c.encodeBillingDocument(billingDocument)
	if err != nil {
		return nil, err
	}

	c.billingDocuments[billingDocument.ID] = b

	return billingDocument, nil
}

// Create creates a BillingDocument in the database
func (c *FakeBillingDocumentClient) Create(ctx context.Context, partitionkey string, billingDocument *pkg.BillingDocument, options *Options) (*pkg.BillingDocument, error) {
	return c.apply(ctx, partitionkey, billingDocument, options, true)
}

// Replace replaces a BillingDocument in the database
func (c *FakeBillingDocumentClient) Replace(ctx context.Context, partitionkey string, billingDocument *pkg.BillingDocument, options *Options) (*pkg.BillingDocument, error) {
	return c.apply(ctx, partitionkey, billingDocument, options, false)
}

// List returns a BillingDocumentIterator to list all BillingDocuments in the database
func (c *FakeBillingDocumentClient) List(*Options) BillingDocumentIterator {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.err != nil {
		return NewFakeBillingDocumentErroringRawIterator(c.err)
	}

	billingDocuments := make([]*pkg.BillingDocument, 0, len(c.billingDocuments))
	for _, d := range c.billingDocuments {
		r, err := c.decodeBillingDocument(d)
		if err != nil {
			return NewFakeBillingDocumentErroringRawIterator(err)
		}
		billingDocuments = append(billingDocuments, r)
	}

	if c.sorter != nil {
		c.sorter(billingDocuments)
	}

	return NewFakeBillingDocumentIterator(billingDocuments, 0)
}

// ListAll lists all BillingDocuments in the database
func (c *FakeBillingDocumentClient) ListAll(ctx context.Context, options *Options) (*pkg.BillingDocuments, error) {
	iter := c.List(options)
	return iter.Next(ctx, -1)
}

// Get gets a BillingDocument from the database
func (c *FakeBillingDocumentClient) Get(ctx context.Context, partitionkey string, id string, options *Options) (*pkg.BillingDocument, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.err != nil {
		return nil, c.err
	}

	billingDocument, exists := c.billingDocuments[id]
	if !exists {
		return nil, &Error{StatusCode: http.StatusNotFound}
	}

	return c.decodeBillingDocument(billingDocument)
}

// Delete deletes a BillingDocument from the database
func (c *FakeBillingDocumentClient) Delete(ctx context.Context, partitionKey string, billingDocument *pkg.BillingDocument, options *Options) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.err != nil {
		return c.err
	}

	_, exists := c.billingDocuments[billingDocument.ID]
	if !exists {
		return &Error{StatusCode: http.StatusNotFound}
	}

	delete(c.billingDocuments, billingDocument.ID)
	return nil
}

// ChangeFeed is unimplemented
func (c *FakeBillingDocumentClient) ChangeFeed(*Options) BillingDocumentIterator {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.err != nil {
		return NewFakeBillingDocumentErroringRawIterator(c.err)
	}

	return NewFakeBillingDocumentErroringRawIterator(ErrNotImplemented)
}

func (c *FakeBillingDocumentClient) processPreTriggers(ctx context.Context, billingDocument *pkg.BillingDocument, options *Options) error {
	for _, triggerName := range options.PreTriggers {
		if triggerHandler := c.triggerHandlers[triggerName]; triggerHandler != nil {
			err := triggerHandler(ctx, billingDocument)
			if err != nil {
				return err
			}
		} else {
			return ErrNotImplemented
		}
	}

	return nil
}

// Query calls a query handler to implement database querying
func (c *FakeBillingDocumentClient) Query(name string, query *Query, options *Options) BillingDocumentRawIterator {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.err != nil {
		return NewFakeBillingDocumentErroringRawIterator(c.err)
	}

	if queryHandler := c.queryHandlers[query.Query]; queryHandler != nil {
		return queryHandler(c, query, options)
	}

	return NewFakeBillingDocumentErroringRawIterator(ErrNotImplemented)
}

// QueryAll calls a query handler to implement database querying
func (c *FakeBillingDocumentClient) QueryAll(ctx context.Context, partitionkey string, query *Query, options *Options) (*pkg.BillingDocuments, error) {
	iter := c.Query("", query, options)
	return iter.Next(ctx, -1)
}

func NewFakeBillingDocumentIterator(billingDocuments []*pkg.BillingDocument, continuation int) BillingDocumentRawIterator {
	return &fakeBillingDocumentIterator{billingDocuments: billingDocuments, continuation: continuation}
}

type fakeBillingDocumentIterator struct {
	billingDocuments []*pkg.BillingDocument
	continuation     int
	done             bool
}

func (i *fakeBillingDocumentIterator) NextRaw(ctx context.Context, maxItemCount int, out interface{}) error {
	return ErrNotImplemented
}

func (i *fakeBillingDocumentIterator) Next(ctx context.Context, maxItemCount int) (*pkg.BillingDocuments, error) {
	if i.done {
		return nil, nil
	}

	var billingDocuments []*pkg.BillingDocument
	if maxItemCount == -1 {
		billingDocuments = i.billingDocuments[i.continuation:]
		i.continuation = len(i.billingDocuments)
		i.done = true
	} else {
		max := i.continuation + maxItemCount
		if max > len(i.billingDocuments) {
			max = len(i.billingDocuments)
		}
		billingDocuments = i.billingDocuments[i.continuation:max]
		i.continuation += max
		i.done = i.Continuation() == ""
	}

	return &pkg.BillingDocuments{
		BillingDocuments: billingDocuments,
		Count:            len(billingDocuments),
	}, nil
}

func (i *fakeBillingDocumentIterator) Continuation() string {
	if i.continuation >= len(i.billingDocuments) {
		return ""
	}
	return fmt.Sprintf("%d", i.continuation)
}

// NewFakeBillingDocumentErroringRawIterator returns a BillingDocumentRawIterator which
// whose methods return the given error
func NewFakeBillingDocumentErroringRawIterator(err error) BillingDocumentRawIterator {
	return &fakeBillingDocumentErroringRawIterator{err: err}
}

type fakeBillingDocumentErroringRawIterator struct {
	err error
}

func (i *fakeBillingDocumentErroringRawIterator) Next(ctx context.Context, maxItemCount int) (*pkg.BillingDocuments, error) {
	return nil, i.err
}

func (i *fakeBillingDocumentErroringRawIterator) NextRaw(context.Context, int, interface{}) error {
	return i.err
}

func (i *fakeBillingDocumentErroringRawIterator) Continuation() string {
	return ""
}