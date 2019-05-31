include scripts/Makefile.common
include scripts/Makefile.golang

.DEFAULT_GOAL := all

docker-build:
	@echo "This component has no docker images"

yaml:
	@echo "This component has no K8S resources"
