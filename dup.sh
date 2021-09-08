#!/bin/bash

# I could not figure out how to do this in the Makefile so I just call this script
a=''
for i in $*; do 
   a="$a module2 $i $i@"
done
echo $a
