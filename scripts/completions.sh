#!/bin/sh
set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
	go run . completion "$sh" >"completions/fztea.$sh"
done