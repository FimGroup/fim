# Todo List, schedule, plan and task

Here we track all the todo and planned tasks.

Once the item has been completed, it will be moved to the main document as feature or changelog as other types.

## Task List

* [ ] data mapping supports array item converter(mapping each element in the array)
    * Note: top level array, e.g. top level struct is array in json
    * Note2: more detailed field mapping, including range of type and type conversion limits
    * [ ] Pipeline mapping
    * [ ] Flow mapping
    * [ ] ModelInst apis
    * [ ] Components/functions mapping
* [ ] new api for files to avoid save files on local disks, for the usecases of cache/temp files/etc, such as large http
  payload temp files in nginx
    * [ ] Support component file loading, e.g. http template files
    * [ ] Local disk readonly impl / mem fs impl with memory limit / S3 fs impl / etc.
* [ ] http and http rest - hostname matching + default server
* [ ] http - file serving / template rendering
    * Note: using new file apis
* [ ] database connection pool count setting

## Todo List

* Add a resource check - e.g. check tcp binding used/conflict to avoid error
* FlowModel/Object/protocols transformation
