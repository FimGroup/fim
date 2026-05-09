# Path & Process Data flow

## Original

* In other systems, the data flow of the pipeline is like the following:
    * Two patterns of the running route: InOnly(produce nothing) and InOut(produce result)
        * This can be changed during the execution of the route.
    * Every InOut node's output is the next node's input.
    * If the flow reaches InOnly patterns(default pattern of a node or custom pattern), then the result will not be
      returned.
    * If the last node is InOut, then the output will be returned.
    * Exception propagation depends on: thread model/component InOnly or InOut; if in the same thread, exceptions always
      propagated

## Design

* There are 3 different dimensions of the behavior:
    * Data passing: InOnly or InOut
    * Flow ending
    * Node execution concurrency
* In this design, the following concepts will be provided independently and at the same time
    1. Decide data passing mode: InOnly/InOut
        * Result / ExecutionStatus
            * Result mode: InOnly/InOut
            * ExecutionStatus includes:
                * Exception: Propagation/Discard
                * Transaction: Commit/Rollback
        * APIs of node: IsOutSupported - indicate the node can use InOut mode or not
    2. (Optional) Decide continue flow execution: Continue/Stop
        * Reason: Introduce a node with predicate to stop the flow, where the flow is controlled by the definition.
    3. Decide wait for or discard responding to inbound request: OneWay/Call
        * Inbound node only
    4. Decide if the node can be executed in-parallel regardless order and output: Sync(default&implicit)/Async
        * Both sync/async can use separate threads beyond the processing thread.
        * Async != run in other thread.
        * Async will not guarantee the input is from the other's output, which means may share the same input.
        * Last output of the Async block will be chosen as final result.
    5. Decide the node state: Stateless/Stateful
        * Stateful node requires some prerequisites to the requests, e.g. shared across different requests
    6. Lifecycle handler:
        * Take actions at the whole lifecycle view: e.g. transaction/deferral
        * Available in Sync node only. Async flow does not care about output so it is unavailable.
    7. Side effect handler:
        * Take actions at the whole lifecycle view: e.g. message status checker of a transactional node
        * Available in both Sync and Async nodes.
        * Something like Lifecycle handler but requires more mechanisms to take effect.
    8. Flow splitter and aggregator
        * Flows often have single/multicast/split/aggregator nodes.
        * It follows the data passing mode and the execution mode.
    9. About linked flows(subsequences)
        * Same as a single node
        * Can decide data passing mode: InOnly/InOut
        * Exception propagation: same as defined in data passing mode
