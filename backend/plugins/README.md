# FreezeTag Plugins

Plugins for FreezeTag are written in Python and are launched by the backend server to handle certain operations (e.g. image upload hooks, user action hooks, etc). Each plugin has its own Python environment, and can leverage PyTorch and Tensorflow for machine learning.

## How to write a plugin
Each plugin is a subdirectory of this `plugins` folder, and contains at minimum a `manifest.json` and a `main.py` that includes the `freezetag` module in this folder. Plugins can specify hooks and actions in the manifest and with function decorators in the `main.py` file.

## How plugins run
Each plugin is launched as a separate Python process by the backend server when it's required for some processing, and stays running until all processing tasks are complete. No plugin will ever have more than one Python process running it at a time, and sequences of processing tasks will be run using the same process.

### Plugin Lifecycle
The life cycle of a plugin is separated into 3 phases:
- **Initialization:** The function registered to the init hook (if there is one) is run before the server makes any requests. This is when the plugin can load models, initialize API connections, etc.
- **Processing**: The process hook (different for different plugin types) is executed at least 1 and potentially many times. This happens synchronously (the backend waits for process to complete before requesting another)
- **Shutdown**: The function registered to the teardown hook (if there is one) is run once the server is done making requests. This allows plugins to clean up, save persistent files for the next run, etc...

## Plugin Protocol Implementation
The communication protocol from plugin to server and vice versa is done through the stdin and stdout pipes on the Python process. Each message in the protocol is an extremely simplified packet of an instruction, and optionally a length and byte contents (sometimes raw, sometimes JSON).

The server and client can each have at most one "transaction" in progress at a time in order to avoid needing a complicated queue structure to handle requests. As an example, here's how a small image tagging request might play out behind the scenes (S is server and P is plugin):
```
initialization
S: READY
P: READY

backend makes a request
S: PUT {"action": "process", "id": 12}
S: BIN [giant RGBA blob]

plugin makes a request and backend responds
P: GET {"action": "metadata", "id": 12}
S: PUT {"action": "metadata", "data": {[image metadata]}}

plugin asks for something to be logged (this doesn't require a response)
P: LOG adding 3 tags to image

plugin responds to original request
P: PUT {"action": "tag", "id": 12, "tags": [array of tags]}

shutdown
S: SHUTDOWN
P: SHUTDOWN
```
Communicating in this order means that there's never ambiguity about which messages belong to which transaction, so neither the plugin nor the backend has to buffer or filter messages.

The only special case messages that don't fit into this 2-layer context-free onion exactly are `LOG`, `ERR`, and `BIN`. `LOG` will only ever be sent by plugins, and the server only has to log (or handle the error) without explicitly responding. `BIN` will only ever show up as additional context after a request or response, never as its own thing.

`ERR` will always be considered a response to the most recent request, so e.g. a "process" request sent by the server that receives `ERR` as a response will be considered completed (in an error state). This is to make sure errors are recoverable in a way that doesn't break the protocol.

If the server receives an `ERR` response to the `READY` message, then it will cancel the plugin. Errors during initialization are considered unrecoverable.