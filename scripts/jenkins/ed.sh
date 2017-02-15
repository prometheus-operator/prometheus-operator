#!/bin/bash

# Kops requires user input through an editor to update a ressource. Instead of
# interacting with an editor we give Kops a fake editor via the 'EDITOR' env
# var. This editor always writes the content of file '$1' into file '$2'. In the
# Makefile before calling 'kops edit ig nodes' we set the 'EDITOR' env var to
# this script with the wanted file as the first argument. The second argument
# which is the file that is supposed to be edited by the user is passed in by
# kops later.

WANTED_FILE=$1
TO_EDIT_FILE=$2

cat $WANTED_FILE > $TO_EDIT_FILE
