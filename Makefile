#
#  Copyright 2019 Nalej
# 

# Name of the target applications to be built
APPS=service-net-agent

# Use global Makefile for common targets
export
%:
	$(MAKE) -f Makefile.golang $@
