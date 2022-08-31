# Toy Language

This is an interpreter for a superset of [BrainF**k](https://en.wikipedia.org/wiki/Brainfuck) implemented in Go.

## Usage

Build the command line tool with `make build` then try to run a bf program using `tl run /path/to/source` (ex: `tl run samples/helloWorld.bf`.)

You can also run tests and benchmarks with `make test` (~85% coverage of `/src`) and `make bench` (~30% coverage of `/src`). The base instructions and parser is almost 100% covered, the missing code coverage comes from the network extension.

## Design

This language has a byte memory array in which it stores data.

The program interacts with the memory array using a pointer and can perform basic operations one cell at the time.

## Commands

The base language syntax is a superset of that of BrainFuck.

### Basic

| Character | Description |
|-----------|-------------|
| `>` | Increment the data pointer |
| `<` | Decrement the data pointer |
| `+` | Increment (by one) the byte at the data pointer |
| `-` | Decrement (by one) the byte at the data pointer |
| `.` | Output the value at the data pointer |
| `,` | Accept one byte of input, store it at the data pointer |
| `[` | If the data pointer byte is zero, jump to the next corresponding `]` |
| `]` | If the data pointer byte is non-zero, jump to the previous corresponding `[` |

### Extensions

To enable a given extensions you must add at the beginning of your file `tl:` followed by a `:`-separated list of extension codes.

#### Networking

The network extension (code: `net`) enables support for basic TCP communication.
This extensions listens for connections on `0.0.0.0` and can send data to `127.0.0.1` on ports ranging from `42000` to `42255`.

This extension operates in 3 states internally:

- `idle`: not connected anywhere.
- `listening`: the user requested a byte from the network.
- `connected`: the user requested a flush of the send queue OR our listener received a connection. In this state the package can send/receive data.

The send cache is emptied when a port is set (`@`) or when we switch to the `listening` state.
Data is received when the status is `connected`, the receive cache is empty, and the user has requested a byte from the network.
Timeouts default to 5 seconds and apply to the send `;` and receive `?` commands, internal timeouts might be different.
The best way to close a connection is set the port again.

Further notes:

- To connect two computers remotely it is suggested use netcat to forward the connection (ex: `nc -k -l {port} | nc {remote} {port}`).
- A successful send is not a guarantee that the receiver got the entire message, if you want that you must write that logic in your code... fun!

| Character | Description |
|-----------|-------------|
| `*` | Set the timeout to 0.1 seconds times the data pointer byte. If 0 will attempt send (`;`) / receive (`?`) operations until successful |
| `@` | Sets the port to `42000` + the data pointer byte |
| `^` | Adds a byte to the send queue. |
| `;` | Sends to `127.0.0.1:{port}` the data in the send queue in up-to 1024 byte packets. Sets the current byte to 0 is successful, 1 otherwise. |
| `?` | Save a recived byte from `0.0.0.0:{port}` to the data pointer. |

## Samples

Some brainfuck code samples are provided in the `/samples` folder.

Note that the repo license might not apply to externally sourced samples as some come from the [Wikipedia page](https://en.wikipedia.org/wiki/Brainfuck#Examples) for this language.

## Future Updates

Here's a list of potential future improvements in non-particular order:

- [build] Offer WASM build
- [debugging] Have a debugging interface to print out detailed error logs
- [tooling:state] More APIs to get access to internal program states
- [tooling:optimizer] Introduce basic code analysis and optimization
- [convert:go] Offer to convert to a simple Go program
- [convert:js] Offer conversion to JavaScript
- [extension:Color] Introduce terminal colors (`~`)
- [extension:MemLoad] APIs to load file into memory and save memory to file (`{`, `}`)
- [extension:External] Extend beyond this library: allows the user to register callbacks (`_`)
- [extension:MaxData] Extend the data 
