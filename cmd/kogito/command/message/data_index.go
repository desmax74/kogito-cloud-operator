// Copyright 2019 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package message

// TODO: start migrating all message to this package to make it easy for docs to review
const (
	DataIndexErrCreating            = "Error while trying to create a new Kogito Data Index Service: %s "
	DataIndexSuccessfulInstalled    = "Kogito Data Index Service successfully installed in the Project %s."
	DataIndexCheckStatus            = "Check the Service status by running 'oc describe kogitodataindex/%s -n %s'"
	DataIndexNotInstalledNoOperator = "Skipping Data Index install since there's no operator available. Use 'kogito install data-index' after installing the operator"
)
