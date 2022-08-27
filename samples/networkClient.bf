tl:net

# Please starts a netcat (nc) server on port 24001 now

-*  # Set timeout to 25 seconds and a half
++@ # Set port to 24001

# Start an infinite loop
# Write something on your terminal and press enter
[>
    , # Read user input
    ^ # Send user input to localhost port 24001
<]