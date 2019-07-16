// Copyright 2019 Ray Holder
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package window

// API defines the functions necessary to monitor window activity.
type API interface {

	// ActiveWindowTitle returns the currently active window's title.
	ActiveWindowTitle() string

	// DetectWindowTitleChanges blocks and starts detecting changes in window
	// titles. When a change is detected, the onChange function is called and
	// when a non-fatal error occurs the onError function is called for that
	// error.
	DetectWindowTitleChanges(onChange func(), onError func(error)) error
}
