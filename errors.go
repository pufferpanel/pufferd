/*
 Copyright 2018 Padduck, LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

 	http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package pufferd

import (
	"github.com/pufferpanel/apufferi/v4"
	"github.com/pufferpanel/apufferi/v4/scope"
)

var ErrServerOffline = apufferi.CreateError("server offline", "ErrServerOffline")
var ErrIllegalFileAccess = apufferi.CreateError("invalid file access", "ErrIllegalFileAccess")
var ErrServerDisabled = apufferi.CreateError("server is disabled", "ErrServerDisabled")
var ErrContainerRunning = apufferi.CreateError("container already running", "ErrContainerRunning")
var ErrImageDownloading = apufferi.CreateError("image downloading", "ErrImageDownloading")
var ErrProcessRunning = apufferi.CreateError("process already running", "ErrProcessRunning")
var ErrMissingFactory = apufferi.CreateError("missing factory", "ErrMissingFactory")
var ErrServerAlreadyExists = apufferi.CreateError("server already exists", "ErrServerAlreadyExists")
var ErrInvalidUnixTime = apufferi.CreateError("time provided is not a valid UNIX time", "ErrInvalidUnixTime")
var ErrKeyNotPEM = apufferi.CreateError("key is not in PEM format", "ErrKeyNotPEM")
var ErrCannotValidateToken = apufferi.CreateError("could not validate access token", "ErrCannotValidateToken")
var ErrMissingAccessToken = apufferi.CreateError("access token not provided", "ErrMissingAccessToken")
var ErrNotBearerToken = apufferi.CreateError("access token must be a Bearer token", "ErrNotBearerToken")
var ErrKeyNotECDSA = apufferi.CreateError("key is not ECDSA key", "ErrKeyNotECDSA")
var ErrMissingScope = apufferi.CreateError("missing scope", "ErrMissingScope")

func CreateErrMissingScope(scope scope.Scope) *apufferi.Error {
	return apufferi.CreateError(ErrMissingScope.Message, ErrMissingScope.Code).Metadata(map[string]interface{}{"scope": scope})
}
