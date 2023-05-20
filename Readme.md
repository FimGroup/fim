# 1. Features

## 1.1 Overview

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
            * General error is returned: FlowError
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

### Core model

#### Container

* Contains all components for a specific tenant/application
* Maintains lifecycle of application

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

# 2. TODO List

* Timeout for synchronous flow + timeout accumulation when processing each step of the flow
    * Plus context
* Data constraints: e.g. not empty/greater than/less than/etc.
* version support - keep a single pipeline stuck to a specific version - can be used for upgrade
* specific node - run specific connector/flow/etc. - e.g. accessing internet may require few nodes and this requires the
  flow to be able to run on those nodes other than any node in the cluster
* Add a resource check - e.g. check tcp binding used/conflict to avoid error
* FlowModel/Object/protocols transformation
* support assign one FlowModel field to different local fields(but not vice versa, for the reason that only one value
  can be assigned and effective to one single field)
* external shared service integration: configuration/service discovery/credential+cert/etc.
* Detail and precise error information
* standard event handling(send and entrypoint without mq vendor spec)
* special connector like zookeeper client
* compatible to design patterns, e.g. soa/ddd/etc.
* unit test on each piece of config

Top priority

4. allow step to run independently from each other to support async invoking
5. loop
6. data mapping supports array item converter(mapping each element in the array)

Before 1st release version

1. stable api and toml config definition
2. clear running/error information
3. Event driven support
4. easy to read document

# 3. Work groups

* core - key components and extensibility
* clustering - support any form and type cluster
* components - connector/functions/plugin/etc.
* optimization - performance/operational
* experience - feedback/improvement/i18n/l10n/error information
* research - experimental research
* observation - operational features
* usecase - project based
* security - security
* interoperability - java/dotnet agent and java/dotnet/etc. service integration
* tooling - debugging/profiling/logging/etc.

# A. Version planning

* version format: x.y.z-type
    * could without -type
* git tag version: vx.y.z-type
* -type enums:
    * -dev
    * -alphaN where N is a number from 0 to Int.Max
    * -betaN where N is a number from 0 to Int.Max
    * -rcN where N is a number from 0 to Int.Max
* number format of x/y/z
    * 1st number - length of the following numbers - e.g. 1-1x, 2-2xx, 3-3xxx
    * single number should not be used unless there will not any version number smaller than it
        * e.g. if 0.0.1 and 0.0.2 are used, the next number should start from 0.0.3 or 0.0.3xxx
    * this aims to support literal sort other than semantic versioning

