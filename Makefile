.PHONY: help portainer portainer-down pull run pull-ui run-ui down-ui down clean get-token get-consul-acl-token start-bootstrapper start-thirdparty logs-bootstrapper logs-thirdparty reload stop exited ps
.SILENT: help get-token

help:
	echo "See README.md in this folder"

ARGS:=$(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(eval $(ARGS):;@:)

OPTIONS:=" arm64 no-secty ui " # Must have spaces around words for `filter-out` function to work properly

ifeq (arm64, $(filter arm64,$(ARGS)))
	ARM64=-arm64
	ARM64_OPTION=arm64
endif
ifeq (no-secty, $(filter no-secty,$(ARGS)))
	NO_SECURITY:=-no-secty
endif
ifeq (ui, $(filter ui,$(ARGS)))
	UI:=-with-ui
endif

SERVICES:=$(filter-out $(OPTIONS),$(ARGS))

# Define additional phony targets for all options to enable support for tab-completion in shell
# Note: This must be defined after the options are parsed otherwise it will interfere with them
.PHONY: $(OPTIONS)

portainer:
	docker-compose -p portainer -f docker-compose-portainer.yml up -d

portainer-down:
	docker-compose -p portainer -f docker-compose-portainer.yml down

pull:
	docker-compose -f docker-compose${NO_SECURITY}${ARM64}.yml pull ${SERVICES}

# bootstrapper & thirdparty container
start-bootstrapper:
	cd ./bootstrapper_go/ && \
	docker build -t bootstrapper . && \
	docker run -d --network host --name bootstrapper bootstrapper && \
	cd ..
start-thirdparty:
	cd ./thirdparty_go/ && \
	docker build -t thirdparty . && \
	docker run -d --network host --name thirdparty thirdparty && \
	cd ..
stop-bootstrapper:
	-docker stop bootstrapper && docker rm bootstrapper && docker rmi bootstrapper
stop-thirdparty:
	-docker stop thirdparty && docker rm thirdparty && docker rmi thirdparty
logs-bootstrapper:
	docker logs bootstrapper -f
logs-thirdparty:
	docker logs thirdparty -f
reload:
	make stop-bootstrapper
	make start-bootstrapper
	make logs-bootstrapper
stop:
	-make stop-bootstrapper
	-make stop-thirdparty
	-docker rm $$(docker ps -a -f status=exited -q)
exited:
	docker ps -f "status=exited"
# continue

run:
# boot up all of the EdgeX's microservice containers
	docker-compose -p edgex -f docker-compose${NO_SECURITY}${UI}${ARM64}.yml up -d ${SERVICES}
# start a fake third party REST API at localhost:7070
	-make start-thirdparty
# wait for them properly running
	sleep 15
# get the root token and transfer it to the bootstrapper
	docker run --rm -ti -v edgex_vault-config:/vault/config:ro alpine:latest cat /vault/config/assets/resp-init.json | tee ./bootstrapper_go/common/root_token.json
# get the proper JWT token and transfer it to the bootstrapper
	make get-token | sed '1d;$$d' | tee ./bootstrapper_go/common/gateway_jwt_token
# start the bootstrapper custom service
	-make start-bootstrapper

down:
	docker-compose -p edgex -f docker-compose.yml down

clean: down
	-docker rm $$(docker ps --filter "network=edgex_edgex-network" --filter "network=edgex_default" -aq) 2> /dev/null
	docker volume prune -f && \
	docker network prune -f
	-make stop

get-token:
	DEV=$(DEV) \
	ARCH=$(ARCH) \
	cd ./compose-builder; sh get-api-gateway-token.sh

get-consul-acl-token:
	DEV=$(DEV) \
	ARCH=$(ARCH) \
	cd ./compose-builder; sh ./get-consul-acl-token.sh

ps:
	docker-compose -p edgex ps
