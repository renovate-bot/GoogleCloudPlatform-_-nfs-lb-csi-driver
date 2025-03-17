#!/bin/sh
# Copyright 2024 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -o errexit

trap "{ exit 0 }" TERM

service rpcbind start

# If statd is already running, for example becuase of an existing nfs mount, we will fail
# to service nfs-common start. If we successfully query the statd service (rpc program
# number 100024), that means it's running, and we don't need to start it. This command
# is put in /etc/default/nfs-common as the NEED_STATD variable must be set there.

if rpcinfo -T udp 127.0.0.1 100024; then
  echo statd already running
  echo NEED_STATD=no >> /etc/default/nfs-common
else
  echo no statd found
fi

service nfs-common start

sleep infinity
