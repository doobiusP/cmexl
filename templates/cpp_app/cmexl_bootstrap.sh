#!/bin/bash

git init
git submodule add https://github.com/Neumann-A/my-vcpkg-triplets external/my-vcpkg-triplets
git submodule update --init --recursive
git add .
git commit -m "Initial commit for NetworkTest"
echo "Submodules added"
