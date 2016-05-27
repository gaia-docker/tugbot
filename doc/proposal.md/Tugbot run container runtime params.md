# Assumptions

1. We want to leverage the capabilities of common docker PaaS's like Kubernetes, Fleet and Docker compose/swarm.
2. We want that "tugbot run" will be the only one responsible for test containers execution.
3. We want to find unified interface between docker PaaS's and "tugbot run"

# Proposal

> This proposal is valid for fleet, but the concept needs to be validated with kubernetes and docker compose/swarm

1. In the test container fleet service we will have "docker run" that will run the container and exist immidiate.
2. In order to do so, the "docker run" will have to have different CMD and sometimes different ENTRYPOINT
2. In the same fleet we will add a new label (with --label) *tugbot.dockerrun.cmd* and *tugbot.dockerrun.entrypoint*
that will contain the actual parameters for the *"real docker run"*
3. *tugbot run* will look for the *terminated containers* and will start to manage any terminated container that is not terminated due to his own run
4. *tugbot run* will get all of the needed information using "docker inspect". And will use the CMD and ENTRYPOINT from *tugbot.dockerrun.cmd* and *tugbot.dockerrun.entrypoint*
