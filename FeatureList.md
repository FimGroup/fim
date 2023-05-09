
# 1. Features

* Two toml definitions: FlowModel(shared models) and Flow
* Data Types: primitives and compound data type
  * Primitives: string, int, bool, float
  * Compound data type: array and object
* FlowModel - shared model and shared data across flows
  * All FlowModel definition files -> merged into single one
* Flow - components:
  * In - inputs
  * Out - outputs
  * PreOut - operations before output
  * Flow - steps of the flow
    * Including: Builtin functions / custom functions / flow / event

# 2. TODO List

* Timeout for synchronous flow + timeout accumulation when processing each step of the flow
* Data constraints: e.g. not empty/greater than/less than/etc.
* Panic handling: avoid unexpected broken flow


