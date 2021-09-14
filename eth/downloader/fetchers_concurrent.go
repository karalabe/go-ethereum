// Copyright 2021 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package downloader

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common/prque"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/log"
)

// typedQueue is an interface defining the adaptor needed to translate the type
// specific downloader/queue schedulers into the type-agnostic general concurrent
// fetcher algorithm calls.
type typedQueue interface {
	// waker returns a notification channel that gets pinged in case more fetches
	// have been queued up, so the fetcher might assign it to idle peers.
	waker() chan bool

	// pending returns the number of wrapped items that are currently queued for
	// fetching by the concurrent downloader.
	pending() int

	// capacity is responsible for calculating how many items of the abstracted
	// type a particular peer is estimated to be able to retrieve within the
	// alloted round trip time.
	capacity(peer *peerConnection, rtt time.Duration) int

	// updateCapacity is responsible for updating how many items of the abstracted
	// type a particular peer is estimated to be able to retrieve in a unit time.
	updateCapacity(peer *peerConnection, items int, elapsed time.Duration)

	// reserve is responsible for allocating a requested number of pending items
	// from the download queue to the specified peer.
	reserve(peer *peerConnection, items int) (*fetchRequest, bool, bool)

	// unreserve is resposible for removing the current retrieval allocation
	// assigned to a specific peer and placing it back into the pool to allow
	// reassigning to some other peer.
	unreserve(peer string) int

	// request is responsible for converting a generic fetch request into a typed
	// one and sending it to the remote peer for fulfillment.
	request(peer *peerConnection, req *fetchRequest, resCh chan *eth.Response) (*eth.Request, error)

	// deliver is responsible for taking a generic response packet from the
	// concurrent fetcher, unpacking the type specific data and delivering
	// it to the downloader's queue.
	deliver(peer *peerConnection, packet *eth.Response) (int, error)
}

// concurrentFetch iteratively downloads scheduled block parts, taking available
// peers, reserving a chunk of fetch requests for each, waiting for delivery and
// also periodically checking for timeouts.
func (d *Downloader) concurrentFetch(queue typedQueue) error {
	// Create a delivery channel to accept responses from all peers
	responses := make(chan *eth.Response)

	// Track the currently active requests and their timeout order
	requests := make(map[string]*eth.Request)
	defer func() {
		// Abort all requests on sync cycle cancellation. The requests will still
		// be fulfilled by the remote side, but the dispatcher will not wait to
		// deliver them since nobody's going to be listening.
		for _, req := range requests {
			req.Close()
		}
	}()
	ordering := make(map[*eth.Request]int)
	timeouts := prque.New(func(data interface{}, index int) {
		ordering[data.(*eth.Request)] = index
	})

	timeout := time.NewTimer(0)
	if !timeout.Stop() {
		<-timeout.C
	}
	defer timeout.Stop()

	// Prepare the queue and fetch block parts until the block header fetcher's done
	finished := false
	for {
		// Short circuit if we lost all our peers
		if d.peers.Len() == 0 {
			return errNoPeers
		}
		// If there's nothing more to fetch, wait or terminate
		if queue.pending() == 0 {
			if len(requests) == 0 && finished {
				return nil
			}
		} else {
			// Send a download request to all idle peers, until throttled
			var (
				idles []*peerConnection
				caps  []int
			)
			for _, peer := range d.peers.AllPeers() {
				if _, ok := requests[peer.id]; !ok {
					idles = append(idles, peer)
					caps = append(caps, queue.capacity(peer, time.Second))
				}
			}
			sort.Sort(&peerCapacitySort{idles, caps})

			var (
				progressed bool
				throttled  bool
				queued     = queue.pending()
			)
			for _, peer := range idles {
				// Short circuit if throttling activated or there are no more
				// queued tasks to be retrieved
				if throttled {
					break
				}
				if queued = queue.pending(); queued == 0 {
					break
				}
				// Reserve a chunk of fetches for a peer. A nil can mean either that
				// no more headers are available, or that the peer is known not to
				// have them.
				rtt := d.peers.rates.TargetRoundTrip()

				request, progress, throttle := queue.reserve(peer, queue.capacity(peer, rtt))
				if progress {
					progressed = true
				}
				if throttle {
					throttled = true
					throttleCounter.Inc(1)
				}
				if request == nil {
					continue
				}
				// Fetch the chunk and make sure any errors return the hashes to the queue
				req, err := queue.request(peer, request, responses)
				if err != nil {
					// Although we could try and make an attempt to fix this, this error really
					// means that we've double allocated a fetch task to a peer. If that is the
					// case, the internal states of the downloader and the queue are very wrong so
					// better hard crash and note the error instead of silently accumulating into
					// a much bigger issue.
					panic(fmt.Sprintf("%v: fetch assignment failed", peer))
				}
				requests[peer.id] = req
				ordering[req] = timeouts.Size()
				timeouts.Push(req, int64(rtt))
				if timeouts.Size() == 1 {
					timeout.Reset(rtt)
				}
			}
			// Make sure that we have peers available for fetching. If all peers have been tried
			// and all failed throw an error
			if !progressed && !throttled && len(requests) == 0 && len(idles) == d.peers.Len() && queued > 0 {
				return errPeersUnavailable
			}
		}
		// Wait for something to happen
		select {
		case <-d.cancelCh:
			// If sync was cancelled, tear down the parallel retriever. Pending
			// requests will be cancelled locally, and the remote responses will
			// be dropped when they arrive
			return errCanceled

		case <-timeout.C:
			// Retrieve the next request which should have timed out. The check
			// below is purely for to catch programming errors, given the correct
			// code, there's no possible order of events that should result in a
			// timeout firing for a non-existent event.
			item, exp := timeouts.Peek()
			if now, at := time.Now(), time.Unix(0, exp); now.Before(at) {
				log.Error("Timeout triggered but not reached", "now", now, "timeout", at)
				timeout.Reset(at.Sub(now))
				continue
			}
			req := item.(*eth.Request)

			// Stop tracking the timed out request from a timing perspective,
			// but leave it in the current pool of requests to avoid assigning
			// a second one to an already overloaded node.
			delete(ordering, req)

			timeouts.Pop()
			if timeouts.Size() > 0 {
				_, exp := timeouts.Peek()
				timeout.Reset(time.Until(time.Unix(0, exp)))
			}
			// New timeout potentially set if there are more requests pending,
			// reschedule the failed one to a free peer
			fails := queue.unreserve(req.Peer)

			// Finally, update the peer's retrieval capacity, or if it's already
			// below the minimum allowance, drop the peer. If a lot of retrieval
			// elements expired, we might have overestimated the remote peer or
			// perhaps ourselves. Only reset to minimal throughput but don't drop
			// just yet.
			//
			// The reason the minimum threshold is 2 is that the downloader tries
			// to estimate the bandwidth and latency of a peer separately, which
			// requires pushing the measured capacity a bit and seeing how response
			// times reacts, to it always requests one more than the minimum (i.e.
			// min 2).
			peer := d.peers.Peer(req.Peer)
			if peer == nil {
				// If the peer got disconnected in between, we should really have
				// short-circuited it already. Just in case there's some strange
				// codepath, leave this check in not to crash.
				log.Error("Delivery timeout from unknown peer", "peer", req.Peer)
				continue
			}
			if fails > 2 {
				queue.updateCapacity(peer, 0, 0)
			} else {
				d.dropPeer(peer.id)

				// If this peer was the master peer, abort sync immediately
				d.cancelLock.RLock()
				master := peer.id == d.cancelPeer
				d.cancelLock.RUnlock()

				if master {
					d.cancel()
					return errTimeout
				}
			}

		case res := <-responses:
			// Response arrived, it may be for an existing or an already timed
			// out request. If the former, update the timeout heap and perhaps
			// reschedule the timeout timer.
			index, live := ordering[res.Req]
			if live {
				timeouts.Remove(index)
				if index == 0 {
					if !timeout.Stop() {
						<-timeout.C
					}
					if timeouts.Size() > 0 {
						_, exp := timeouts.Peek()
						timeout.Reset(time.Until(time.Unix(0, exp)))
					}
				}
			}
			if _, ok := requests[res.Req.Peer]; ok {
				delete(requests, res.Req.Peer)
			} else {
				log.Error("Response arrived to non-tracked request")
			}
			// Signal the dispatcher that the round trip is done. We'll drop the
			// peer if the data turns out to be junk.
			res.Done <- nil
			res.Req.Close()

			// If the peer was previously banned and failed to deliver its pack
			// in a reasonable time frame, ignore its message.
			if peer := d.peers.Peer(res.Req.Peer); peer != nil {
				// Deliver the received chunk of data and check chain validity
				accepted, err := queue.deliver(peer, res)
				if errors.Is(err, errInvalidChain) {
					return err
				}
				// Unless a peer delivered something completely else than requested (usually
				// caused by a timed out request which came through in the end), set it to
				// idle. If the delivery's stale, the peer should have already been idled.
				if !errors.Is(err, errStaleDelivery) {
					queue.updateCapacity(peer, accepted, res.Time)
				}
			}

		case cont := <-queue.waker():
			// The header fetcher sent a continuation flag, check if it's done
			if !cont {
				finished = true
			}
		}
	}
}
