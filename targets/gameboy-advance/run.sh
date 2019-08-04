#!/bin/bash

cp --no-clobber /skel/* ./
exec make "$@"
