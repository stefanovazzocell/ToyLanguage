Converts inputted characters to a Brainfuck representation
++++++++++                                   # New Line
>+++++++++++++++++++++++++++++++++++++++++++ # Plus
>           # Move to a new memory cell
+[          # Do the following
    ,           # Get user input
    [           # Unless user input is 0 do the following
        <.          # Print the plus
        >-          # Decrease user input by 1
    ]           # Repeat until user input is 0
    ,       # Ignore User new line input
    <<.>>   # New Line
+]          # Loop forever