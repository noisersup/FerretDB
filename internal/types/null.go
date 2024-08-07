// Copyright 2021 FerretDB Inc.
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

package types

import "log/slog"

type (
	// NullType represents BSON type Null.
	//
	// Most callers should use types.Null value instead.
	NullType struct{}
)

// Null represents BSON value Null.
var Null = NullType{}

// LogValue implements [slog.LogValuer].
func (n NullType) LogValue() slog.Value {
	return slogValue(n, 1)
}

// check interfaces
var (
	_ slog.LogValuer = NullType{}
)
