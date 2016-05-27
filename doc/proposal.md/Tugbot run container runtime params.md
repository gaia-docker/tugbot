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

**Fleet service for example:**

```
# Copyright 2015 HP Software
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

[Unit]
Description=End-to-End test for Feature X

[Service]
Restart=on-failure
RestartSec=20s

# Change killmode from "control-group" to "none" to let Docker remove
# work correctly.
KillMode=none

ExecStart=/bin/bash -a -c 'docker run \
-p 8050:8080 \
-v /:/rootfs:ro \
-v /var/run:/var/run:rw \
-v /sys:/sys:ro \
-v /var/lib/docker/:/var/lib/docker:ro \
--label=tugbot.dockerrun.cmd=java -jar mytests.jar
my-org/feature-x-tests ps -ef > /dev/null'

[Install]
WantedBy=multi-user.target

[X-Fleet]
Global=true
```
