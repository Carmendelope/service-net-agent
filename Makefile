
.DEFAULT_GOAL := all

docker-build:
	@echo "This component has no docker images"

yaml:
	@echo "This component has no K8S resources"

include scripts/Makefile.golang
include scripts/Makefile.common
