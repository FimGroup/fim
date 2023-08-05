# 1. Features

## 1.1 Overview

`Make it easy and flexible to integration business logic and external integrations.`

And

`Quick implementation of a prototype or even a production ready application.`

* Using configurable flow file as preference rather than API with coding
    * Config file format: Toml
    * [Format definition](DocFormatDefinition.md)
* Provide flexible APIs to be able to extend the functionalities for complex business and on-demand requirements

#### Notable characteristics

* Running core functionality on a 256MB memory vm with reasonable performance(keep required plugins)
    * Require minimum resource
    * See [sample forum system based on fim](https://github.com/FimGroup/sample-fim-forum-system)
* Running as the following types of service
    * Standalone server with single or clustered support
        * On Bare metal machines
        * On VMs on both local or cloud
        * On containers with/without scheduler(e.g. k8s/docker compose)
    * Cloud PaaS/Local APPs on cloud providers / Embedded in APPs
    * Function/Serverless on cloud providers
    * As API provider
    * Support flexible scaling out and scaling up
* Optimized for the following scenarios
    * Throughput: e.g. internet application, data pipeline, IoT connections
    * Latency: e.g. realtime control, decision engine
    * Business logic: e.g. Internet Services, ERP, CRM
    * Cost: e.g. resource cost, development cost, quality cost, resource usage
    * Integration/Open: e.g. various tech stacks, various customer requirements

#### Usage documents that you only need to know to use Flexible Integration Mesh :)

* Write configure file: [Format definition](DocFormatDefinition.md)
* Write entrypoint of service using minimum APIs: [DocEndUserAPI.md](DocEndUserAPI.md)

## 1.2 Detail Features

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
        * Use FlowModel as in/out parameters
        * Flow
        * Connector: Source connector/Target connector
* Check functions
    * For the purpose to validate data/break flow/respond error information
    * Two types of check functions
        * break current flow: check_XXX_break - break current flow and respond error
            * General error is returned: *FlowError
        * non-breaking: check_XXX - check and set error information in local parameter for branching logic
* Lifecycle of requests
    * Start of requests: user request or scheduled job
    * Note: Events can be regarded as start of request or not. Recommended not to regard events as start point.
    * Note2: Events types - standard event / delayed(scheduled) event
    * Start point: source connector which is the entrypoint of user request/scheduled job/event listener(not
      recommended)
* Branching
    * Branching feature is part of Flow/Pipeline to support dynamic step selection when processing request
    * Format of branching in a common programming language: if/switch-case/pattern-matching/etc.
    * Introduce pattern-matching like operators to support branching in Flow/Pipeline definition
    * Current supported branching operator:
        * (flow/pipeline) @case-true
        * (flow/pipeline) @case-false
        * (flow) @case-equals: two parameters(may use together with @assign function)
        * (flow) @case-not-equals: two parameters(may use together with @assign function)
        * (flow/pipeline) @case-empty: parameter is null or empty string - ""
        * (flow/pipeline) @case-non-empty: parameter is not empty string
* Configure replacement
    * Configure placement allows to custom configure via various injection ways rather than hardcoded value
    * Avoid credential/cert/decryption or other configurations being exposed unexpectedly via configure injection
      support
    * There are two types of configure replacement
        * static configure replacement: such as a placeholder in the configure file
            * The configuration will be replaced once when startup
            * Prefix 'configure-static://' in values in configure files
        * dynamic configure replacement
            * The configuration will be replaced every time it is used
            * Prefix 'configure-dynamic://' in values in configure files
        * Even though the two types are nearly the same, the reason to separate them is allowing more options for
          plugins to deal with configure replacement
    * Currently, supports
        * source connector and steps(including target connector) in pipeline
* Resource management
    * Resource manager is used for providing different types of resources with the core facilities and the plugins
    * Types of resources, e.g. files, sockets, etc.
* Distributed support
    * About distributed support:
        * Enabling the possibility to run a request from a source connector to be processed on distributed cluster
    * Naming:
        * Pipeline level: "ContainerName(business name)/PipelineName"
        * Flow level: (currently not supported)
    * Type of distribution:
        * RPC: realtime call
        * Event: trigger asynchronous flow without response, e.g. message/delayed message/job
    * Level of distribution:
        * Pipeline level: source connector -> (dispatch) -> pipeline handler
        * Flow level: previous flow -> (dispatch) -> next flow
* Connector
    * Source & Target connector
    * (Refer to below section)

### Core model

#### Application

* Provide application level lifecycle management
* Start/Stop with the whole application
* Name property: physical application that provides environment for businesses running on
    * Name convention: same as path format

#### Container

* Contains all components for a specific tenant/app/business
* Maintains lifecycle of tenant/app/business
* Name property: container name or business name that defines a set of business flows
    * Name convention: same as path format

#### FlowModel

(Refer to above section)

#### Pipeline

(Refer to above section)

#### Connector

* connector generator
    * instance by: generator name
    * store: connectors and states
    * spawn connectors by rules inside the generator
    * may or may not have lifecycle control / require self-management based on connector lifecycle if possible
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
* Note: the lifecycle of connector generally speaking is independent to container. But for the instance of connector,
  since it should be unique within one container, the lifecycle of connector instance should be tight to container
  lifecycle.
* connector plugin
    * types
        * builtin plugin - provided by connector by default and can be used directly
        * custom plugin - user customized plugin for the specific connector which providing plugin feature
    * plugin customization capabilities
        * plugin options + logging/configure manager/resource manager/etc.(fundamental mechanisms)
        * custom code/logic
        * lifecycle
        * in/out parameter per request
        * order
    * Implementation solution
        * embedded in connector
        * standard mechanism provided by framework

### Supporting model

#### Flow

(Refer to above section)

#### Builtin/Custom Function

(Refer to above section)

#### Project entrypoint

The following apis can be used in projects(for starting container and custom functions)

* package basicapi
* components.InitComponent
* fimcore.NewUseContainer

### Hierarchy and order of core components and trigger points

* Note: new type may be added in the future and will be kept the same principles below.
* Note2: since go doesn't have isolation mechanism such as classload in java, some components that provide resources
  like connectors have to be in Application level in order to be shared across all containers.

Details

* (Shared/default components)
    * default LoggingManager
        * no lifecycle
* Application
    * (Add the following components)
    * ConfigureManager
    * FileResourceManager
    * Target connector generator
    * Source connector generator
    * Source/Target connector pre-initialized generator definition
    * (start application)
        * Do lifecycle management of the items above
            * ConfigureManager
            * FileResourceManager
            * Target connector generator
            * Source connector generator
            * Target connector pre-initialized generator initialization
            * Source connector pre-initialized generator initialization
    * (Spawn container)
* Container
    * (Add the following components)
    * ConfigureManager (separate from Application level)
        * no lifecycle (may change in the future)
    * Builtin and custom functions
    * FlowModel/Pipeline/Flow definitions
    * Spawn source and target connector
    * (start container)
        * Combining pipeline and source connector
        * Source and target connector lifecycle. Same instance name may be shared in several places.

# 2. Project planning details, Project progress and TODO List

See detailed document: [DocTodoList.md](DocTodoList.md)

# 3. Work groups

* core - key components and extensibility
* clustering - support any form and type cluster
* components - connector/functions/plugin/etc.
* optimization - performance/operational
* experience - feedback/improvement/i18n/l10n/error information/documentation
* research - experimental research
* observation - operational features
* usecase - project based
* security - security
* interoperability - java/dotnet agent and java/dotnet/etc. service integration
* tooling - IDE/debugging/profiling/logging/etc.

Top priority tasks

* optimize knowledge to use fim
* more connectors
* Interoperability with existing third party components
* service forms - standalone, as part of function, peer agent

# A. Version planning

* version format: x.y.z-type
    * could without -type
* git tag version: vx.y.z-type
* -type enums:
    * -dev
    * -alphaN where N is a number from 0 to Int.Max, prefer no more than 10 or to following the number format below
    * -betaN where N is a number from 0 to Int.Max, prefer no more than 10 or to following the number format below
    * -rcN where N is a number from 0 to Int.Max, prefer no more than 10 or to following the number format below
* number format of x/y/z
    * 1st number - length of the following numbers - e.g. 1-1x, 2-2xx, 3-3xxx
    * single number should not be used unless there will not be any version number smaller than it
        * e.g. if 0.0.1 and 0.0.2 are used, the next number should start from 0.0.3 or 0.0.3xxx
    * this aims to support literal sort other than semantic versioning

