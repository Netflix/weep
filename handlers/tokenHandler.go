/*
 * Copyright 2020 Netflix, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handlers

import (
	"fmt"
	"net/http"
)

const staticToken = "AQAEANQlVdnIoNfmJQHofbSTjkIm8eoMIBZZZX05Xk9jLiFuJuL2_A=="

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	// Returning a static token allows us to support IMDSv2 with minimal effort.
	fmt.Fprint(w, staticToken)
}
