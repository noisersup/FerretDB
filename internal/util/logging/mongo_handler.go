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

package logging

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mongoHandler struct {
	opts *NewHandlerOpts

	jsonHandler slog.Handler
}

type mongoLog struct {
	Timestamp primitive.DateTime `bson:"t"`
}

func (h *mongoHandler) Enabled(_ context.Context, l slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}

	return l >= minLevel
}

func (h *mongoHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.jsonHandler.Handle(ctx, r)
}

func (h *mongoHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.jsonHandler.WithAttrs(attrs)
}

func (h *mongoHandler) WithGroup(name string) slog.Handler {
	return h.jsonHandler.WithGroup(name)
}
