#! /bin/sh
#
# Set the environment variables for building the scaffolder.
#
# Change directory to the one containing this file and run it, for example:
#
#    cd $HOME/workspaces/films
#    . setenv.sh

if test -z $GOPATH
then
    GOPATH=`pwd`
    export GOPATH
else
    GOPATH=$GOPATH:`pwd`
    export GOPATH
fi

PATH=`pwd`/bin:$PATH
export PATH
