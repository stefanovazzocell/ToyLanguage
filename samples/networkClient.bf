tl:net

# Please starts a netcat (nc) server on port 42001 now

-*  # Set timeout to 25 seconds and a half
++@ # Set port to 42001

# Start an infinite loop
# Write something on your terminal and press enter
[>
    , # Read user input
    ^ # Queues the user input to be sent to localhost port 42001
    ; # Flushes the send cache
<]