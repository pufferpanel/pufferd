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

package errors

import "github.com/pufferpanel/apufferi"

var ErrServerOffline = apufferi.CreateError("server offline", "ErrServerOffline")
var ErrIllegalFileAccess = apufferi.CreateError("invalid file access", "ErrIllegalFileAccess")
var ErrServerDisabled = apufferi.CreateError("server is disabled", "ErrServerDisabled")
var ErrContainerRunning = apufferi.CreateError("container already running", "ErrContainerRunning")
var ErrImageDownloading = apufferi.CreateError("image downloading", "ErrImageDownloading")
var ErrProcessRunning = apufferi.CreateError("process already running", "ErrProcessRunning")
var ErrMissingFactory = apufferi.CreateError("missing factory", "ErrMissingFactory")