# Todo List, schedule, plan and task

Here we track all the todo and planned tasks.

Once the item has been completed, it will be moved to the main document as feature or changelog as other types.

## Task List

* [ ] data mapping supports array item converter(mapping each element in the array)
    * Note: top level array, e.g. top level struct is array in json(like htp body is a json array)
    * [x] Pipeline mapping
    * [x] Flow mapping
    * [x] Functions mapping
    * [x] Source/Target connector mapping
    * [ ] Check data type of each value when mapping values according to DataTypeStore
        * Flow in and out parameter type should match DataTypeStore after complete the flow(data assign may happen in
          the flow)
        * Can check data type in parallel when setting value in ModelInst2
    * [x] Implement or update pluginapi.Model interface to allow plugin to assign and retrieve value
    * [x] Convert all flow files
        * [x] flows
        * [x] connectors
        * [x] verify flow files
* [ ] new api for files to avoid save files on local disks, for the usecases of cache/temp files/etc, such as large http
  payload temp files in nginx
    * [ ] Support component file loading, e.g. http template files
    * [ ] Local disk readonly impl / mem fs impl with memory limit / S3 fs impl / etc.
* [ ] http and http rest - hostname matching + default server
* [ ] http - file serving / template rendering
    * Note: using new file apis
* [ ] Make connector mapping embedded
    * [ ] convert all flow files
* [ ] connector lifecycle new approach to support connector level and instance level
* [ ] make container and connector independent
    * connector can spawn instances to use containers
    * when one container is EOL, connector instance should be released and the event should also be delivered to
      connector
* [ ] Instance name for connectors

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

## Before 1st release (planned as v1.0.0)

`v0.0.1-v0.9.9 as alpha stage versions`

1. stable api and toml config definition
    * provide easy-to-read document
2. clear running/error information
3. Event driven support and path decision(sync and event)

## Changelogs

#### v0.0.2 (in-progress)

* Logger

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
