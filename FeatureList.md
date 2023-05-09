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
        * Including: Builtin functions / custom functions / flow / event
* Pipeline - TODO

# 2. TODO List

* Timeout for synchronous flow + timeout accumulation when processing each step of the flow
* Data constraints: e.g. not empty/greater than/less than/etc.
* Panic handling: avoid unexpected broken flow
* Pipeline-Flow options
* Builtin/custom methods, source/target connector - howto
* lifecycle of Pipeline/Flow/functions/connector/etc. - e.g. http server connector should be able to shutdown

