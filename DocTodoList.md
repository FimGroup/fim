# Todo List, schedule, plan and task

Here we track all the todo and planned tasks.

Once the item has been completed, it will be moved to the main document as feature or changelog as other types.

## Task List

* [ ] Http connector plugin support
    * [x] request id/real ip/logging/recover
    * [x] http auth jwt plugin
    * [ ] http session
    * [ ] timeout
    * [ ] anti-spam
    * [x] cors
    * [ ] compression
    * [ ] header filters - content encodings/content types/charset/no cache
    * [ ] ratelimit
    * [ ] https and HSTS
    * [ ] csrf
* [ ] Distributed processing pipeline
    * [x] pipeline dispatch
    * [ ] flow dispatch
* [ ] Flow type support - event type: trigger another Fim flow(pipeline)
* [ ] new api for files to avoid save files on local disks, for the usecases of cache/temp files/etc, such as large http
  payload temp files in nginx
    * [x] Support component file loading, e.g. http template files
    * [ ] Local disk readonly impl / mem fs impl with memory limit / S3 fs impl / etc.
* [ ] http and http rest - hostname matching + default server
* [ ] Provide init functions for application and container to avoid providing set/add in api
    * For user to inject plugins
* [ ] Should merge source and target connector?
* [ ] standard event handling(send and entrypoint without mq vendor spec)
* [ ] Demo project
    * [ ] Forum
    * [ ] Realworld https://github.com/gothinkster/realworld
* [ ] http static file: caching headers for browser/CDN
* [ ] error handling
    * Error information
    * Panic handling
* [ ] Context support
    * go context and framework context
    * connector/flows/components/async/etc.
* [ ] simple object as parameter for flow/custom function

## Todo List

* [ ] Add a resource check - e.g. check tcp binding used/conflict to avoid error
    * including permission check, limiting a certain resource being used by unauthorized flows
* [ ] FlowModel/Object/protocols transformation
* [ ] Add a special source/target connector for procedures
* [ ] support local parameter in pipeline, avoiding defining temporary parameter in FlowModel
* [ ] more detailed field mapping, including range of type and type conversion limits for pipeline, flow, connector,
  function and etc
* [ ] new design: converter based on path
    * support object/array
    * used for pipeline/flow/connector/function/etc
* [ ] Multidimensional array in mapping
    * Prefer to regard such type of array as single dimensional array for convenience
    * protocol that support this: json
    * protocol that not support this: protobuf
* [ ] Array operators for mapping
    * Split into array / merge into object or primitives
    * Read/Write single index, leaving untouched indexes default values(primitive default value or null object)
    * Create empty array
    * Assign/Retrieve from/to array field in an object
* [ ] Parent field assignment and operators
    * assign a field in the child mapping rule using src value from parent levels
    * operators to retrieve parent field value
    * operators to filter/map children values
* [ ] Auto make sure acceptable primitive types when assigning values in ModelInst2
    * [x] For plugins including functions and connectors
    * [ ] For core facilities including flow configure static value
    * [ ] Need an automatic way to ensure the types entering or leaving ModelInst2
* [ ] Determine whether to support non-single type array
    * values of multiple types(mixed of primitive/object/array) in one single array
* [ ] Capture panic and handling error in pipeline
    * blocking error
    * soft error, can be skipped
* [ ] reduce go.mod dependencies
    * Keep minimum dependencies for plugins
    * Using builtin implementations for shared and small pieces
* [ ] InboundAccumulator
* [ ] Add lifecycle management for connector generator
* [ ] DataType check when accessing ModelInst2
* [ ] Precise and easy mechanisms for debugging
* [ ] Disallow modification to application after started up
* [ ] Add swagger support for http rest
* [ ] http file serving, e.g. js/css/images
* [ ] graceful shutdown and restart
* [ ] Design pattern, SOA and DDD design support
    * updated model for design and architecture
* [ ] Testing/Reliability verification
* [ ] version support - keep a single pipeline stuck to a specific version - can be used for upgrade
* [ ] special connector of such behaviors like zookeeper client
* [ ] DAG flow
* [ ] Transaction support
* [ ] refactor of http connector
* Timeout for synchronous flow + timeout accumulation when processing each step of the flow
    * Plus context
* Data constraints: e.g. not empty/greater than/less than/etc.
* specific node - run specific connector/flow/etc. - e.g. accessing internet may require few nodes and this requires the
  flow to be able to run on those nodes other than any node in the cluster
* support assign one FlowModel field to different local fields(but not vice versa, for the reason that only one value
  can be assigned and effective to one single field)
* external shared service integration: configuration/service discovery/credential+cert/etc.
* Detail and precise error information
* unit test on each piece of config
* new connector: imap/pop3/smtp, file, scheduling
* platform specific agent: jee/dotnet/etc...
* metrics/tracing/etc...
* as a function integrated into serverless apis, such as aws lambda
* path/configure literal key format validation
* pre-generated logic for flows/conversions optimization
* container name duplication check when spawning new container
* [ ] serialization method support
    * json
    * protobuf
    * xml
    * etc.
* Connectors enhancement
    * Http connector
        * Auth module
            * Key rotate
            * More options: expire time / etc..

## Before 1st release (planned as v1.0.0)

`v0.0.1-v0.9.9 as alpha stage versions`

1. stable api and toml config definition
    * provide easy-to-read document
2. clear running/error information
3. Event driven support and path decision(sync and event)
4. accurate lifecycle control

Top priority

4. allow step to run independently from each other to support async invoking
5. loop

## Changelogs

#### v0.0.4

* Core facility changes
    * Add Application name / Container name(business name)
    * Connector plugin
        * Option/configure
        * Builtin plugin(embedded in connector)
* Distributed processing pipeline
    * pipeline dispatch
    * Event implementation by nats
* http connector
    * static file
    * plugin
        * auth jwt
        * CORS
* Job scheduler connector
    * cron scheduler support
* Messaging connector
    * Messaging source connector
    * Messaging target connector
    * Implementation by nats

#### v0.0.3

* Core
    * Update core models of application/container/connector
    * define lifecycle of each type of components
* Http connector
    * Support path parameter
* Postgres connector
    * Migrate to new lifecycle

#### v0.0.2

* Core
    * Application now spawns containers. Containers cannot be created without application.
* Data mapping rule
    * Add flow in/out parameter type matching validation
    * Embed source/target connector mapping rule in connector definition
* Logger
    * Logger API
    * Http access log of source connector
    * Standalone logging package
* Http connector
    * Template rendering

#### v0.0.1

* Initial release with basic concepts
    * Container/FlowModel/Pipeline/Flow/Function/Source and target connector
* Plugins
    * http rest source connector
    * postgresql database connector
    * builtin functions
* Data mapping rule
    * Primitive type
    * Object
    * Array
