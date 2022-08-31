tl:net

# Send bytes to local port 42002 using netcat (nc)

-*  # Set timeout to 25 seconds and a half
+++@ # Set port to 42002

# Start an infinite loop
[>
    ? # Receive on all IPv4 port 42002 and write to memory
    . # Print to stdout
<]