/*
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

package clocki

import (
	"time"

	"github.com/jonboulle/clockwork"
)

// FakeClock duplicates its namesake interface as defined by clockwork versions
// prior to v0.5.0.
type FakeClock interface {
	clockwork.Clock

	// Advance advances the FakeClock to a new point in time, ensuring any existing
	// waiters are notified appropriately before returning.
	Advance(d time.Duration)
	// BlockUntil blocks until the FakeClock has the given number of waiters.
	BlockUntil(waiters int)
}
