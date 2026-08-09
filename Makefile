.DEFAULT_GOAL := help

.PHONY: help
help:
	@task --list

# Compatibility bridge for existing docs and muscle memory.
# Examples:
#   make test
#   make migrate-create NAME=add_new_feature
#   make test-profile-queries ENDPOINT=feed_list DURATION=60
%:
	@task $@
