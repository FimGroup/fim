# Flow configure file format

TBD

##### Model/parameter mapping rule basics

1. primitive type - do mapping and assignment directly
   ```text
      ["source_path_level", "target_path_level"]
   ```
2. object、array types - do mapping and assignment on the whole field
    * Meaning that object/array cannot be directly assigned.
    * This format allows multiple levels of array mapping.
   ```text
   ["source_path_level", "target_path_level", [
     # List of sub-level mappings for object and array
     # For primitive array, leave mapping rule empty
   ]]
   ```
    * Note: array type must be array definition(xxx[]) / array access(xxx[2]) is not allowed
3. For rule #2, source_path_level or target_path_level can be empty string which representing the certain level does not
   exist
    * But both cannot be empty at the same time.
4. For building the source and target data structure, both will be recursively built based on the nested mapping rule.
5. Special cases
    * For the case that have multiple layers of arrays, if the number of layers of opposite site does not match it,
      array mapping result may be overwritten due to the multiple layer iteration mapping.
    * For accessing one specific element in an array, filter/aggregation/etc operators should be used
        * Sample1: `["src[1]", "target"]` or `["src", "target[2]"]`
            * Should use operators to partially access array
        * Sample2:
           ```text
             ["src[1]", "target", [
               # List of sub-level mappings for object and array
             ]]
           ```
            * Should use operators to partially access array

Unsupported:

1. Multidimensional array and array operators

Rule parsing explanation:

1. validate basic rule format and check input validation
2. generate converter flow with data mapping and data type checks according to the rule
3. do each level data mapping when invoking converter and perform runtime data type check

### operators(not complete):

##### mapping operators(TODO)

* split - item -> array
* merge - array -> item


