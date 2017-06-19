#! /usr/bin/env bash

go build;
./gson -collate -inpfile example1 > out1
./gson -n1qlsort -inpfile example1 > out2
