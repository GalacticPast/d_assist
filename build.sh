#!/bin/bash
templ fmt . | find . -name "*.templ"
templ generate . | find . -name "*.templ"
make build

