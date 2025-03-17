#!/bin/bash
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

# This is a helper script to determine the IP to node distribution when pods (using the nfs csi lb volumes) are actively deployed on the cluster
node_ip_output=$(./lb_list_assigned_ip_node_ann.sh)
declare -A ip_counts

while read -r line; do
    ip=$(echo "$line" | cut -d' ' -f2)

    if [[ $ip =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
        ((ip_counts[$ip]++)) 
    else
        echo "Skipping line: No valid IP detected - $line"  # Optional: Log skipped lines
    fi
done <<< "$node_ip_output"

for ip in "${!ip_counts[@]}"; do
    echo "IP: $ip, Count: ${ip_counts[$ip]}"
done

