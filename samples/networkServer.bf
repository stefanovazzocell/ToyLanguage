tl:net

# Send bytes to local port 24002 using netcat (nc)
# Note that this accepts only 1 byte at the time
# This limitation might be fixed in the future by optimizing the network server

-*  # Set timeout to 25 seconds and a half
+++@ # Set port to 24002

# Start an infinite loop
[>
    ? # Receive on all IPv4 port 24002 and write to memory
    . # Print to stdout
<]