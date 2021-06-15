#!/bin/sh

# Copyright 2016 The Kubernetes Authors.
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
set -o nounset

if [ ! -f "build/build.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

#Delete old swagger. Sometimes overwriting doesnt seem to work properly.
rm -rf ./docs/docs.go
rm -rf ./docs/swagger

if [ -z "$SWAGGER_INFO" ]; then
	#No info file specified.
	swag init
else
	#If an info file is specified.
	swag init -g "$SWAGGER_INFO"
fi