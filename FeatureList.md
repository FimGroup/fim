# 1. Features

* Two toml definitions: FlowModel(shared models) and Flow
* Data Types: primitives and compound data type
    * Primitives: string, int, bool, float
    * Compound data type: array and object
* Path/Model definition
    * Path format
        * Compatible to available characters from xml
        * e.g.letters, digits, hyphens, underscores, and periods
        * Note: colons are reserved
    * Model Definition
        * Fields - same as path format
        * Full path of a field - several levels of path format joint by '/'
* FlowModel - shared model and shared data across flows
    * All FlowModel definition files -> merged into single one
* Flow - components:
    * In - inputs
    * Out - outputs
    * PreOut - operations before output
    * Flow - steps of the flow
        * Including: Builtin functions / custom functions
* Pipeline - components:
    * Pipeline is the top level abstraction for processing a request
        * Each pipeline defines a usecase of the business
    * source connector
        * Req/Res data type mapping
        * Support options to initialize the connector
    * steps
        * Invoke flow/Trigger event
        * Flow/Target connector
            * Use FlowModel as input and output models
            * Support options to initialize the connector
* Customized components
    * Used in flow
        * Builtin functions
        * Customized functions
    * Used in pipeline
        * Flow
        * Connector: Source connector/Target connector

### Core model

#### Connector

* connector generator
    * instance by: generator name
    * store: connectors and states
    * spawn connectors by rules inside the generator
* connector
    * instance by: generator + instance name
    * connector name: used by pipeline
    * lifecycle control in the container(may not release resource immediately since other pipelines may use) and also
      controlled by generator
* connector model
    * generator : connector = 1 : N
    * pipeline : connector = M : N
    * generator: for shared resource like listening to a network address
    * connector: for defining the entrypoint of a request

# 2. TODO List

* Timeout for synchronous flow + timeout accumulation when processing each step of the flow
* Data constraints: e.g. not empty/greater than/less than/etc.
* Panic handling: avoid unexpected broken flow
* version support - keep a single pipeline stuck to a specific version - can be used for upgrade
* specific node - run specific connector/flow/etc. - e.g. accessing internet may require few nodes and this requires the
  flow to be able to run on those nodes other than any node in the cluster
* Add a resource check - e.g. check tcp binding used/conflict to avoid error
* FlowModel/Object/protocols transformation
* Connector instance and lifecycle management
* data mapping supports array item converter(mapping each element in the array)
* support assign one FlowModel field to different local fields(but not vice versa, for the reason that only one value
  can be assigned and effective to one single field)

Top priority

1. Independent components(not nested existing in other component): container, pipeline, connector, FlowModel
    * lifecycle of Pipeline/Flow/functions/connector/etc. - e.g. http server connector should be able to shutdown
    * refactor component apis & lifecycles
    * split container apis
    * lifecycle of a request: user input/scheduling, event is not part of standalone lifecycle
2. Merged toml file definition
3. error handling & interrupt pipeline
4. allow step to run independently from each other to support async invoking

# 3. Work groups

* core - key components and extensibility
* clustering - support any form and type cluster
* components - connector/functions/plugin/etc.
* optimization - performance/operational
* experience - feedback/improvement
* research - experimental research
* observation - operational features
* usecase - project based
* security - security
* interoperability - java/dotnet agent and java/dotnet/etc. service integration
* tooling - debugging/profiling/logging/etc.

